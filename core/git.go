package core

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/google/uuid"
	"github.com/kyokomi/emoji"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// A future like struct to hold the result of git clone
type gitCloneResult struct {
	ClonePath string // The abs path in os.TempDir() where the the item was cloned to
	Error     error  // An error which occurred during the clone
	mu        sync.RWMutex
}

// Mutex safe getter
func (result *gitCloneResult) get() string {
	result.mu.RLock()
	clonePath := result.ClonePath
	result.mu.RUnlock()
	return clonePath
}

// R/W safe map of {[cacheToken]: gitCloneResult}
type gitCache struct {
	mu    sync.RWMutex
	cache map[string]*gitCloneResult
}

// Mutext safe getter
func (cache *gitCache) get(cacheToken string) (*gitCloneResult, bool) {
	cache.mu.RLock()
	value, ok := cache.cache[cacheToken]
	cache.mu.RUnlock()
	return value, ok
}

// Mutex safe setter
func (cache *gitCache) set(cacheToken string, cloneResult *gitCloneResult) {
	cache.mu.Lock()
	cache.cache[cacheToken] = cloneResult
	cache.mu.Unlock()
}

// cacheKey combines a git-repo, branch, and commit into a unique key used for caching to a map
func cacheKey(repo, branch, commit string) string {
	if len(branch) == 0 {
		branch = "master"
	}
	if len(commit) == 0 {
		commit = "head"
	}
	return fmt.Sprintf("%v@%v:%v", repo, branch, commit)
}

// cache is a global git map cache of {[cacheKey]: gitCloneResult}
var cache = gitCache{
	cache: map[string]*gitCloneResult{},
}

// cloneRepo clones a target git repository into the hosts temporary directory
// and returns a gitCloneResult pointing to that location on filesystem
func (cache *gitCache) cloneRepo(repo string, commit string, branch string, accessToken string) chan *gitCloneResult {
	cloneResultChan := make(chan *gitCloneResult)

	go func() {
		cacheToken := cacheKey(repo, branch, commit)

		// Check if the repo is cloned/being-cloned
		cloneResult, ok := cache.get(cacheToken)
		if ok {
			log.Info(emoji.Sprintf(":atm: Previously cloned %s this install; reusing", cacheToken))
			cloneResultChan <- cloneResult
		} else {
			// Add the clone future to cache
			cloneResult := gitCloneResult{
				Error: nil,
			}
			cloneResult.mu.Lock()               // lock the future
			cache.set(cacheToken, &cloneResult) // store future in cache

			// Default options for a clone
			cloneOptions := &git.CloneOptions{
				URL:          repo,
				SingleBranch: true,
			}

			// Only fetch latest commit if commit provided
			if len(commit) == 0 {
				log.Println(emoji.Sprintf(":helicopter: Component requested latest commit: fast cloning at --depth 1"))
				cloneOptions.Depth = 1
			} else {
				log.Println(emoji.Sprintf(":helicopter: Component requested commit '%s': need full clone", commit))
			}

			// Add the auth token to auth if provided
			if len(accessToken) != 0 {
				cloneOptions.Auth = &http.BasicAuth{
					Username: "not-needed-when-using-PAT-but-cant-be-blank",
					Password: accessToken,
				}
			}

			// Add branch reference option if provided
			if len(branch) != 0 {
				log.Println(emoji.Sprintf(":helicopter: Component requested branch '%s'", branch))
				cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(branch)
			}

			// Clone into a random path in the host temp dir
			uuid, err := uuid.NewRandom()
			if err != nil {
				cloneResultChan <- &gitCloneResult{
					Error: err,
				}
			}
			clonePathOnFS := path.Join(os.TempDir(), uuid.String())
			log.Println(emoji.Sprintf(":helicopter: Cloning %s into %s", cacheToken, clonePathOnFS))
			r, err := git.PlainClone(clonePathOnFS, false, cloneOptions)
			if err != nil {
				cloneResultChan <- &gitCloneResult{
					Error: err,
				}
			}

			// If commit provided, checkout the commit
			if len(commit) != 0 {
				w, err := r.Worktree()
				if err != nil {
					cloneResultChan <- &gitCloneResult{
						Error: err,
					}
				}

				err = w.Checkout(&git.CheckoutOptions{
					Hash: plumbing.NewHash(commit),
				})
				if err != nil {
					cloneResultChan <- &gitCloneResult{
						Error: err,
					}
				}
			}

			// Save the gitCloneResult into cache
			cloneResult.ClonePath = clonePathOnFS
			cloneResult.mu.Unlock() // unlock/resolve the future

			// Push the cached result to the channel
			cloneResultChan <- &cloneResult
		}
	}()

	return cloneResultChan
}

// CloneRepo is a helper func to centralize cloning a repository with the spec
// provided by its arguments.
func CloneRepo(repo string, commit string, intoPath string, branch string, accessToken string) (err error) {
	// Clone and get the location of where it was cloned to in tmp
	result := <-cache.cloneRepo(repo, commit, branch, accessToken)
	clonePath := result.get()
	if result.Error != nil {
		return result.Error
	}

	// copy the repo from tmp cache to component path
	log.Println(emoji.Sprintf(":truck: Copying %s into %s", clonePath, intoPath))
	if err = copy.Copy(clonePath, intoPath); err != nil {
		return err
	}

	return err
}

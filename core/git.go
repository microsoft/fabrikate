package core

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sync"

	"github.com/google/uuid"
	"github.com/kyokomi/emoji"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
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

// Thread safe store of {[gitRepo]: token}
type gitAccessTokenMap struct {
	mu     sync.RWMutex
	tokens map[string]string
}

// Get is a thread safe getter to do a map lookup in a getAccessTokens
func (t *gitAccessTokenMap) Get(repo string) (string, bool) {
	t.mu.RLock()
	token, exists := t.tokens[repo]
	t.mu.RUnlock()
	return token, exists
}

// Set is a thread safe setter method to modify a gitAccessTokenMap
func (t *gitAccessTokenMap) Set(repo, token string) {
	t.mu.Lock()
	t.tokens[repo] = token
	t.mu.Unlock()
}

// cacheKey combines a git-repo, branch, and commit into a unique key used for
// caching to a map
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

// GitAccessTokens is a thread-safe global store of Personal Access Tokens which
// is used to store PATs as they are discovered throughout the Install lifecycle
var GitAccessTokens = gitAccessTokenMap{
	tokens: map[string]string{},
}

// cloneRepo clones a target git repository into the hosts temporary directory
// and returns a gitCloneResult pointing to that location on filesystem
func (cache *gitCache) cloneRepo(repo string, commit string, branch string) chan *gitCloneResult {
	cloneResultChan := make(chan *gitCloneResult)

	go func() {
		cacheToken := cacheKey(repo, branch, commit)

		// Check if the repo is cloned/being-cloned
		if cloneResult, ok := cache.get(cacheToken); ok {
			log.Info(emoji.Sprintf(":atm: Previously cloned %s this install; reusing", cacheToken))
			cloneResultChan <- cloneResult
		} else {
			// Add the clone future to cache
			cloneResult := gitCloneResult{}
			cloneResult.mu.Lock()               // lock the future
			cache.set(cacheToken, &cloneResult) // store future in cache

			// Default options for a clone
			cloneCommandArgs := []string{"clone"}

			// check for access token and append to repo if present
			if token, exists := GitAccessTokens.Get(repo); exists {
				// Only match when the repo string does not contain a an access token already
				// "(https?)://(?!(.+:)?.+@)(.+)" would be preferred but go does not support negative lookahead
				pattern, err := regexp.Compile("^(https?)://([^@]+@)?(.+)$")
				if err != nil {
					cloneResultChan <- &gitCloneResult{Error: err}
				}
				// If match is found, inject the access token into the repo string
				matches := pattern.FindStringSubmatch(repo)
				if matches != nil {
					protocol := matches[1]
					// credentialsWithAtSign := matches[2]
					cleanedRepoString := matches[3]
					repo = fmt.Sprintf("%v://%v@%v", protocol, token, cleanedRepoString)
				}
			}

			// Add repo to clone args
			cloneCommandArgs = append(cloneCommandArgs, repo)

			// Only fetch latest commit if commit provided
			if len(commit) == 0 {
				log.Println(emoji.Sprintf(":helicopter: Component requested latest commit: fast cloning at --depth 1"))
				cloneCommandArgs = append(cloneCommandArgs, "--depth", "1")
			} else {
				log.Println(emoji.Sprintf(":helicopter: Component requested commit '%s': need full clone", commit))
			}

			// Add branch reference option if provided
			if len(branch) != 0 {
				log.Println(emoji.Sprintf(":helicopter: Component requested branch '%s'", branch))
				cloneCommandArgs = append(cloneCommandArgs, "--branch", branch)
			}

			// Clone into a random path in the host temp dir
			randomFolderName, err := uuid.NewRandom()
			if err != nil {
				cloneResultChan <- &gitCloneResult{Error: err}
			}
			clonePathOnFS := path.Join(os.TempDir(), randomFolderName.String())
			log.Println(emoji.Sprintf(":helicopter: Cloning %s into %s", cacheToken, clonePathOnFS))
			cloneCommandArgs = append(cloneCommandArgs, clonePathOnFS)
			cloneCommand := exec.Command("git", cloneCommandArgs...)
			cloneCommand.Env = append(cloneCommand.Env, "GIT_TERMINAL_PROMPT=0") // tell git to fail if it asks for credentials
			if err = cloneCommand.Run(); err != nil {
				cloneResultChan <- &gitCloneResult{Error: err}
			}

			// If commit provided, checkout the commit
			if len(commit) != 0 {
				log.Println(emoji.Sprintf(":helicopter: Performing checkout at commit %s", commit))
				checkoutCommit := exec.Command("git", "checkout", commit)
				checkoutCommit.Dir = clonePathOnFS
				if err = checkoutCommit.Run(); err != nil {
					cloneResultChan <- &gitCloneResult{Error: err}
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
func CloneRepo(repo string, commit string, intoPath string, branch string) (err error) {
	// Clone and get the location of where it was cloned to in tmp
	result := <-cache.cloneRepo(repo, commit, branch)
	clonePath := result.get()
	if result.Error != nil {
		return result.Error
	}

	// copy the repo from tmp cache to component path
	log.Info(emoji.Sprintf(":truck: Copying %s into %s", clonePath, intoPath))
	if err = copy.Copy(clonePath, intoPath); err != nil {
		return err
	}

	return err
}

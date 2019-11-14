package core

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/google/uuid"
	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/logger"
	"github.com/otiai10/copy"
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

// Mutex safe getter
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
			logger.Info(emoji.Sprintf(":atm: Previously cloned '%s' this install; reusing cached result", cacheToken))
			cloneResultChan <- cloneResult
			close(cloneResultChan)
			return
		}

		// Add the clone future to cache
		cloneResult := gitCloneResult{}
		cloneResult.mu.Lock() // lock the future
		defer func() {
			cloneResult.mu.Unlock() // ensure the lock is released
			close(cloneResultChan)
		}()
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
				return
			}
			// If match is found, inject the access token into the repo string
			if matches := pattern.FindStringSubmatch(repo); matches != nil {
				protocol := matches[1]
				// credentialsWithAtSign := matches[2]
				cleanedRepoString := matches[3]
				repo = fmt.Sprintf("%v://%v@%v", protocol, token, cleanedRepoString)
			}
		}

		// Add repo to clone args
		cloneCommandArgs = append(cloneCommandArgs, repo)

		// Only fetch latest commit if commit not provided
		if len(commit) == 0 {
			logger.Info(emoji.Sprintf(":helicopter: Component requested latest commit: fast cloning at --depth 1"))
			cloneCommandArgs = append(cloneCommandArgs, "--depth", "1")
		} else {
			logger.Info(emoji.Sprintf(":helicopter: Component requested commit '%s': need full clone", commit))
		}

		// Add branch reference option if provided
		if len(branch) != 0 {
			logger.Info(emoji.Sprintf(":helicopter: Component requested branch '%s'", branch))
			cloneCommandArgs = append(cloneCommandArgs, "--branch", branch)
		}

		// Clone into a random path in the host temp dir
		randomFolderName, err := uuid.NewRandom()
		if err != nil {
			cloneResultChan <- &gitCloneResult{Error: err}
			return
		}
		clonePathOnFS := path.Join(os.TempDir(), randomFolderName.String())
		logger.Info(emoji.Sprintf(":helicopter: Cloning %s => %s", cacheToken, clonePathOnFS))
		cloneCommandArgs = append(cloneCommandArgs, clonePathOnFS)
		cloneCommand := exec.Command("git", cloneCommandArgs...)
		cloneCommand.Env = append(cloneCommand.Env, os.Environ()...)         // pass all env variables to git command so proper SSH config is passed if needed
		cloneCommand.Env = append(cloneCommand.Env, "GIT_TERMINAL_PROMPT=0") // tell git to fail if it asks for credentials

		if output, err := cloneCommand.CombinedOutput(); err != nil {
			logger.Error(emoji.Sprintf(":no_entry_sign: Error occurred while cloning: '%s'\n%s: %s", cacheToken, err, output))
			cloneResultChan <- &gitCloneResult{Error: err}
			return
		}

		// If commit provided, checkout the commit
		if len(commit) != 0 {
			logger.Info(emoji.Sprintf(":helicopter: Performing checkout commit '%s' for repo '%s'", commit, repo))
			checkoutCommit := exec.Command("git", "checkout", commit)
			checkoutCommit.Dir = clonePathOnFS
			if output, err := checkoutCommit.CombinedOutput(); err != nil {
				logger.Error(emoji.Sprintf(":no_entry_sign: Error occurred checking out commit '%s' from repo '%s'\n%s: %s", commit, repo, err, output))
				cloneResultChan <- &gitCloneResult{Error: err}
				return
			}
		}

		// Save the gitCloneResult into cache
		cloneResult.ClonePath = clonePathOnFS

		// Push the cached result to the channel
		cloneResultChan <- &cloneResult
	}()

	return cloneResultChan
}

type git struct{}

// Git is function wrapper to expose host git functionality
var Git = git{}

// CloneRepo is a helper func to centralize cloning a repository with the spec
// provided by its arguments.
func (g git) CloneRepo(repo string, commit string, intoPath string, branch string) (err error) {
	// Clone and get the location of where it was cloned to in tmp
	result := <-cache.cloneRepo(repo, commit, branch)
	clonePath := result.get()
	if result.Error != nil {
		return result.Error
	}

	// Remove the into directory if it already exists
	if err = os.RemoveAll(intoPath); err != nil {
		return err
	}

	// copy the repo from tmp cache to component path
	absIntoPath, err := filepath.Abs(intoPath)
	if err != nil {
		return err
	}
	logger.Info(emoji.Sprintf(":truck: Copying %s => %s", clonePath, absIntoPath))
	if err = copy.Copy(clonePath, intoPath); err != nil {
		return err
	}

	return err
}

// CleanGitCache deletes all temporary folders created as temporary cache for
// git clones.
func (g git) CleanGitCache() (err error) {
	logger.Info(emoji.Sprintf(":bomb: Cleaning up git cache..."))
	cache.mu.Lock()
	for key, value := range cache.cache {
		logger.Info(emoji.Sprintf(":bomb: Removing git cache directory '%s'", value.ClonePath))
		if err = os.RemoveAll(value.ClonePath); err != nil {
			logger.Error(emoji.Sprintf(":exclamation: Error deleting temporary directory '%s'", value.ClonePath))
			cache.mu.Unlock()
			return err
		}
		delete(cache.cache, key)
	}
	cache.mu.Unlock()
	logger.Info(emoji.Sprintf(":white_check_mark: Completed cache clean!"))
	return err
}

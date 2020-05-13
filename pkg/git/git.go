package git

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
	"github.com/microsoft/fabrikate/pkg/logger"
	"github.com/otiai10/copy"
)

// A future like struct to hold the result of git clone
type cloneResult struct {
	ClonePath string // The abs path in os.TempDir() where the the item was cloned to
	Error     error  // An error which occurred during the clone
	mu        sync.RWMutex
}

// Mutex safe getter
func (result *cloneResult) get() string {
	result.mu.RLock()
	clonePath := result.ClonePath
	result.mu.RUnlock()
	return clonePath
}

// R/W safe map of {[cacheToken]: cloneResult}
type cacheMap struct {
	mu    sync.RWMutex
	cache map[string]*cloneResult
}

// Mutex safe getter
func (c *cacheMap) get(cacheToken string) (*cloneResult, bool) {
	c.mu.RLock()
	value, ok := c.cache[cacheToken]
	c.mu.RUnlock()
	return value, ok
}

// Mutex safe setter
func (c *cacheMap) set(token string, result *cloneResult) {
	c.mu.Lock()
	c.cache[token] = result
	c.mu.Unlock()
}

// Thread safe store of {[gitRepo]: token}
type accessTokenMap struct {
	mu     sync.RWMutex
	tokens map[string]string
}

// Get is a thread safe getter to do a map lookup in a getAccessTokens
func (t *accessTokenMap) Get(repo string) (string, bool) {
	t.mu.RLock()
	token, exists := t.tokens[repo]
	t.mu.RUnlock()
	return token, exists
}

// Set is a thread safe setter method to modify a accessTokenMap
func (t *accessTokenMap) Set(repo, token string) {
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

// cache is a global git map cache of {[cacheKey]: cloneResult}
var cache = cacheMap{
	cache: map[string]*cloneResult{},
}

// GitAccessTokens is a thread-safe global store of Personal Access Tokens which
// is used to store PATs as they are discovered throughout the Install lifecycle
var AccessTokens = accessTokenMap{
	tokens: map[string]string{},
}

// cloneRepo clones a target git repository into the hosts temporary directory
// and returns a cloneResult pointing to that location on filesystem
func (c *cacheMap) cloneRepo(repo string, commit string, branch string) chan *cloneResult {
	cloneResultChan := make(chan *cloneResult)

	go func() {
		cacheToken := cacheKey(repo, branch, commit)

		// Check if the repo is cloned/being-cloned
		if cloneResult, ok := c.get(cacheToken); ok {
			logger.Info(emoji.Sprintf(":atm: Previously cloned '%s' this install; reusing cached result", cacheToken))
			cloneResultChan <- cloneResult
			close(cloneResultChan)
			return
		}

		// Add the clone future to c
		result := cloneResult{}
		result.mu.Lock() // lock the future
		defer func() {
			result.mu.Unlock() // ensure the lock is released
			close(cloneResultChan)
		}()
		c.set(cacheToken, &result) // store future in c

		// Default options for a clone
		cloneCmdArgs := []string{"clone"}

		// check for access token and append to repo if present
		if token, exists := AccessTokens.Get(repo); exists {
			// Only match when the repo string does not contain a an access token already
			// "(https?)://(?!(.+:)?.+@)(.+)" would be preferred but go does not support negative lookahead
			pattern, err := regexp.Compile("^(https?)://([^@]+@)?(.+)$")
			if err != nil {
				cloneResultChan <- &cloneResult{Error: err}
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
		cloneCmdArgs = append(cloneCmdArgs, repo)

		// Only fetch latest commit if commit not provided
		if len(commit) == 0 {
			logger.Info(emoji.Sprintf(":helicopter: Component requested latest commit: fast cloning at --depth 1"))
			cloneCmdArgs = append(cloneCmdArgs, "--depth", "1")
		} else {
			logger.Info(emoji.Sprintf(":helicopter: Component requested commit '%s': need full clone", commit))
		}

		// Add branch reference option if provided
		if len(branch) != 0 {
			logger.Info(emoji.Sprintf(":helicopter: Component requested branch '%s'", branch))
			cloneCmdArgs = append(cloneCmdArgs, "--branch", branch)
		}

		// Clone into a random path in the host temp dir
		randomFolderName, err := uuid.NewRandom()
		if err != nil {
			cloneResultChan <- &cloneResult{Error: err}
			return
		}
		clonePathOnFS := path.Join(os.TempDir(), randomFolderName.String())
		logger.Info(emoji.Sprintf(":helicopter: Cloning %s => %s", cacheToken, clonePathOnFS))
		cloneCmdArgs = append(cloneCmdArgs, clonePathOnFS)
		cloneCmd := exec.Command("git", cloneCmdArgs...)
		cloneCmd.Env = append(cloneCmd.Env, os.Environ()...)         // pass all env variables to git command so proper SSH config is passed if needed
		cloneCmd.Env = append(cloneCmd.Env, "GIT_TERMINAL_PROMPT=0") // tell git to fail if it asks for credentials

		if output, err := cloneCmd.CombinedOutput(); err != nil {
			logger.Error(emoji.Sprintf(":no_entry_sign: Error occurred while cloning: '%s'\n%s: %s", cacheToken, err, output))
			cloneResultChan <- &cloneResult{Error: err}
			return
		}

		// If commit provided, checkout the commit
		if len(commit) != 0 {
			logger.Info(emoji.Sprintf(":helicopter: Performing checkout commit '%s' for repo '%s'", commit, repo))
			checkoutCommit := exec.Command("git", "checkout", commit)
			checkoutCommit.Dir = clonePathOnFS
			if output, err := checkoutCommit.CombinedOutput(); err != nil {
				logger.Error(emoji.Sprintf(":no_entry_sign: Error occurred checking out commit '%s' from repo '%s'\n%s: %s", commit, repo, err, output))
				cloneResultChan <- &cloneResult{Error: err}
				return
			}
		}

		// Save the cloneResult into c
		result.ClonePath = clonePathOnFS

		// Push the cached result to the channel
		cloneResultChan <- &result
	}()

	return cloneResultChan
}

// Clone is a helper func to centralize cloning a repository with the spec
// provided by its arguments.
func Clone(repo string, commit string, intoPath string, branch string) (err error) {
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

// CleanCache deletes all temporary folders created as temporary cache for
// git clones.
func CleanCache() (err error) {
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

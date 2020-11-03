package git

import "sync"

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

// cache is a global git map cache of {[cacheKey]: gitCloneResult}
var cache = gitCache{
	cache: map[string]*gitCloneResult{},
}

// GitAccessTokens is a thread-safe global store of Personal Access Tokens which
// is used to store PATs as they are discovered throughout the Install lifecycle
var GitAccessTokens = gitAccessTokenMap{
	tokens: map[string]string{},
}

package kra

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
)

// cacheEntry represents a cached item with expiration metadata
type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// isExpired reports whether the entry has passed its TTL
func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiration)
}

// CacheManager provides a groupcache-backed LRU cache with TTL semantics
type CacheManager struct {
	cache      *lru.Cache
	mu         sync.RWMutex
	enabled    bool
	debug      bool
	maxEntries int
}

// NewCacheManager creates a new cache manager backed by groupcache's LRU implementation
func NewCacheManager(enabled bool, debug bool, maxEntries int) *CacheManager {
	if maxEntries <= 0 {
		maxEntries = 1024
	}

	var lruCache *lru.Cache
	if enabled {
		lruCache = lru.New(maxEntries)
	}

	return &CacheManager{
		cache:      lruCache,
		enabled:    enabled,
		debug:      debug,
		maxEntries: maxEntries,
	}
}

// Get retrieves a value from the cache
//
// Returns the cached value and true if found and not expired,
// otherwise returns nil and false.
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	if !cm.enabled {
		return nil, false
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	value, ok := cm.cache.Get(key)
	if !ok {
		if cm.debug {
			fmt.Printf("[Cache] MISS: %s\n", key)
		}
		return nil, false
	}

	entry, _ := value.(*cacheEntry)
	if entry == nil || entry.isExpired() {
		cm.cache.Remove(key)
		if cm.debug {
			fmt.Printf("[Cache] EXPIRED: %s\n", key)
		}
		return nil, false
	}

	if cm.debug {
		fmt.Printf("[Cache] HIT: %s\n", key)
	}
	return entry.value, true
}

// Set stores a value in the cache with the specified TTL
//
// If TTL is 0 or negative, the entry will never expire (not recommended).
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) {
	if !cm.enabled {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	entry := &cacheEntry{
		value:      value,
		expiration: time.Now().Add(ttl),
	}

	cm.cache.Add(key, entry)

	if cm.debug {
		fmt.Printf("[Cache] SET: %s (TTL: %v)\n", key, ttl)
	}
}

// Delete removes an entry from the cache
func (cm *CacheManager) Delete(key string) {
	if !cm.enabled {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.cache.Remove(key)

	if cm.debug {
		fmt.Printf("[Cache] DELETE: %s\n", key)
	}
}

// Clear removes all entries from the cache
func (cm *CacheManager) Clear() {
	if !cm.enabled {
		return
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.cache = lru.New(cm.maxEntries)

	if cm.debug {
		fmt.Println("[Cache] CLEAR: All entries removed")
	}
}

// GetOrSet retrieves a value from cache or computes it using the provided function
//
// This is a convenience method that combines Get and Set operations.
// If the value is not in cache, it calls the compute function, stores the result,
// and returns it.
//
// The compute function is called outside the lock to prevent deadlocks.
func (cm *CacheManager) GetOrSet(
	key string,
	compute func() (interface{}, error),
	ttl time.Duration,
) (interface{}, error) {
	// Try to get from cache first
	if value, found := cm.Get(key); found {
		return value, nil
	}

	// Compute the value (outside the lock)
	value, err := compute()
	if err != nil {
		return nil, err
	}

	// Store in cache
	cm.Set(key, value, ttl)

	return value, nil
}

// Size returns the current number of entries in the cache
func (cm *CacheManager) Size() int {
	if !cm.enabled {
		return 0
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.cache.Len()
}

// GenerateCacheKey creates a cache key from operation name and parameters
//
// This is a helper function to create consistent cache keys across the SDK.
//
// Example:
//
//	key := GenerateCacheKey("pin_verification", "P051234567A")
//	// Returns: "pin_verification:P051234567A"
func GenerateCacheKey(operation string, params ...string) string {
	if len(params) == 0 {
		return operation
	}

	key := operation
	for _, param := range params {
		key += ":" + param
	}
	return key
}

package kra

import (
	"fmt"
	"sync"
	"time"
)

// cacheEntry represents a cached item with expiration
type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

// isExpired checks if the cache entry has expired
func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiration)
}

// CacheManager manages in-memory caching with TTL support
//
// The cache is goroutine-safe and automatically cleans up expired entries.
type CacheManager struct {
	cache   map[string]*cacheEntry
	mu      sync.RWMutex
	enabled bool
	debug   bool
}

// NewCacheManager creates a new cache manager
//
// If enabled is false, the cache will be a no-op (all operations will return cache misses).
//
// The cleanup goroutine runs every minute to remove expired entries.
func NewCacheManager(enabled bool, debug bool) *CacheManager {
	cm := &CacheManager{
		cache:   make(map[string]*cacheEntry),
		enabled: enabled,
		debug:   debug,
	}

	if enabled {
		// Start background cleanup goroutine
		go cm.cleanupLoop()
	}

	return cm
}

// Get retrieves a value from the cache
//
// Returns the cached value and true if found and not expired,
// otherwise returns nil and false.
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	if !cm.enabled {
		return nil, false
	}

	cm.mu.RLock()
	defer cm.mu.RUnlock()

	entry, exists := cm.cache[key]
	if !exists {
		if cm.debug {
			fmt.Printf("[Cache] MISS: %s\n", key)
		}
		return nil, false
	}

	if entry.isExpired() {
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

	expiration := time.Now().Add(ttl)
	cm.cache[key] = &cacheEntry{
		value:      value,
		expiration: expiration,
	}

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

	delete(cm.cache, key)

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

	cm.cache = make(map[string]*cacheEntry)

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

	return len(cm.cache)
}

// cleanupLoop runs periodically to remove expired entries
func (cm *CacheManager) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.cleanup()
	}
}

// cleanup removes all expired entries from the cache
func (cm *CacheManager) cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	expiredCount := 0

	for key, entry := range cm.cache {
		if now.After(entry.expiration) {
			delete(cm.cache, key)
			expiredCount++
		}
	}

	if cm.debug && expiredCount > 0 {
		fmt.Printf("[Cache] CLEANUP: Removed %d expired entries\n", expiredCount)
	}
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

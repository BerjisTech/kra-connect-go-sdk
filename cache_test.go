package kra

import (
	"fmt"
	"testing"
	"time"
)

func newTestCacheManager(enabled bool) *CacheManager {
	return NewCacheManager(enabled, false, 32)
}

func TestCacheManager_SetAndGet(t *testing.T) {
	cm := newTestCacheManager(true)

	// Set a value
	cm.Set("test-key", "test-value", 1*time.Hour)

	// Get the value
	value, found := cm.Get("test-key")
	if !found {
		t.Error("Expected to find cached value")
	}

	if value.(string) != "test-value" {
		t.Errorf("Expected 'test-value', got %v", value)
	}
}

func TestCacheManager_GetNonExistent(t *testing.T) {
	cm := newTestCacheManager(true)

	// Try to get non-existent key
	_, found := cm.Get("non-existent")
	if found {
		t.Error("Expected not to find value for non-existent key")
	}
}

func TestCacheManager_Expiration(t *testing.T) {
	cm := newTestCacheManager(true)

	// Set a value with very short TTL
	cm.Set("test-key", "test-value", 100*time.Millisecond)

	// Immediate get should succeed
	_, found := cm.Get("test-key")
	if !found {
		t.Error("Expected to find value immediately after setting")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Get should fail after expiration
	_, found = cm.Get("test-key")
	if found {
		t.Error("Expected not to find value after expiration")
	}
}

func TestCacheManager_Delete(t *testing.T) {
	cm := newTestCacheManager(true)

	// Set a value
	cm.Set("test-key", "test-value", 1*time.Hour)

	// Verify it exists
	_, found := cm.Get("test-key")
	if !found {
		t.Error("Expected to find cached value")
	}

	// Delete it
	cm.Delete("test-key")

	// Verify it's gone
	_, found = cm.Get("test-key")
	if found {
		t.Error("Expected not to find value after deletion")
	}
}

func TestCacheManager_Clear(t *testing.T) {
	cm := newTestCacheManager(true)

	// Set multiple values
	cm.Set("key1", "value1", 1*time.Hour)
	cm.Set("key2", "value2", 1*time.Hour)
	cm.Set("key3", "value3", 1*time.Hour)

	// Verify size
	if size := cm.Size(); size != 3 {
		t.Errorf("Expected size 3, got %d", size)
	}

	// Clear cache
	cm.Clear()

	// Verify all values are gone
	if size := cm.Size(); size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", size)
	}

	_, found := cm.Get("key1")
	if found {
		t.Error("Expected not to find key1 after clear")
	}
}

func TestCacheManager_GetOrSet(t *testing.T) {
	cm := newTestCacheManager(true)

	callCount := 0
	compute := func() (interface{}, error) {
		callCount++
		return "computed-value", nil
	}

	// First call should compute
	value, err := cm.GetOrSet("test-key", compute, 1*time.Hour)
	if err != nil {
		t.Fatalf("GetOrSet() error = %v", err)
	}
	if value.(string) != "computed-value" {
		t.Errorf("Expected 'computed-value', got %v", value)
	}
	if callCount != 1 {
		t.Errorf("Expected compute to be called once, called %d times", callCount)
	}

	// Second call should use cache
	value, err = cm.GetOrSet("test-key", compute, 1*time.Hour)
	if err != nil {
		t.Fatalf("GetOrSet() error = %v", err)
	}
	if value.(string) != "computed-value" {
		t.Errorf("Expected 'computed-value', got %v", value)
	}
	if callCount != 1 {
		t.Errorf("Expected compute to be called once total, called %d times", callCount)
	}
}

func TestCacheManager_Disabled(t *testing.T) {
	cm := newTestCacheManager(false)

	// Set a value
	cm.Set("test-key", "test-value", 1*time.Hour)

	// Get should always miss when disabled
	_, found := cm.Get("test-key")
	if found {
		t.Error("Expected not to find value when cache is disabled")
	}

	// Size should be 0 when disabled
	if size := cm.Size(); size != 0 {
		t.Errorf("Expected size 0 when disabled, got %d", size)
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		params    []string
		want      string
	}{
		{
			name:      "no params",
			operation: "operation",
			params:    []string{},
			want:      "operation",
		},
		{
			name:      "single param",
			operation: "operation",
			params:    []string{"param1"},
			want:      "operation:param1",
		},
		{
			name:      "multiple params",
			operation: "operation",
			params:    []string{"param1", "param2", "param3"},
			want:      "operation:param1:param2:param3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCacheKey(tt.operation, tt.params...)
			if got != tt.want {
				t.Errorf("GenerateCacheKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheManager_Concurrent(t *testing.T) {
	cm := newTestCacheManager(true)

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := GenerateCacheKey("test", fmt.Sprintf("%d", id))
			cm.Set(key, id, 1*time.Hour)
			if value, found := cm.Get(key); found {
				if value.(int) != id {
					t.Errorf("Expected %d, got %v", id, value)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCacheManager_EvictsLeastRecentlyUsed(t *testing.T) {
	cm := NewCacheManager(true, false, 2)

	cm.Set("a", "A", time.Hour)
	cm.Set("b", "B", time.Hour)

	if cm.Size() != 2 {
		t.Fatalf("expected size 2, got %d", cm.Size())
	}

	// Access "a" so that "b" becomes LRU
	if _, ok := cm.Get("a"); !ok {
		t.Fatalf("expected to get key a")
	}

	// Insert "c" to trigger eviction
	cm.Set("c", "C", time.Hour)

	if _, ok := cm.Get("b"); ok {
		t.Fatalf("expected key b to be evicted")
	}

	if _, ok := cm.Get("a"); !ok {
		t.Fatalf("expected key a to remain")
	}
}

func TestCacheManager_DebugLogging(t *testing.T) {
	cm := NewCacheManager(true, true, 4)
	cm.Set("key", "value", time.Millisecond)
	cm.Get("key")
	cm.Delete("key")
	cm.Clear()
}

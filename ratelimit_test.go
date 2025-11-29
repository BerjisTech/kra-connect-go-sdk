package kra

import (
	"testing"
	"time"
)

func TestRateLimiter_TryAcquire(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Minute, true, false)

	// Should be able to acquire 5 tokens
	for i := 0; i < 5; i++ {
		if !rl.TryAcquire() {
			t.Errorf("Expected to acquire token %d", i+1)
		}
	}

	// 6th attempt should fail
	if rl.TryAcquire() {
		t.Error("Expected to fail acquiring 6th token")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Second, true, false)

	// Acquire 2 tokens
	rl.TryAcquire()
	rl.TryAcquire()

	// Wait for refill (should block briefly)
	start := time.Now()
	rl.Wait()
	duration := time.Since(start)

	// Should have waited at least 500ms for one token to refill
	if duration < 400*time.Millisecond {
		t.Errorf("Expected to wait at least 400ms, waited %v", duration)
	}
}

func TestRateLimiter_AvailableTokens(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Minute, true, false)

	// Initial tokens should be 10
	if tokens := rl.AvailableTokens(); tokens != 10 {
		t.Errorf("Expected 10 initial tokens, got %d", tokens)
	}

	// Acquire 3 tokens
	rl.TryAcquire()
	rl.TryAcquire()
	rl.TryAcquire()

	// Should have 7 tokens left
	if tokens := rl.AvailableTokens(); tokens != 7 {
		t.Errorf("Expected 7 tokens after acquiring 3, got %d", tokens)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Minute, true, false)

	// Acquire all tokens
	for i := 0; i < 5; i++ {
		rl.TryAcquire()
	}

	// Should have 0 tokens
	if tokens := rl.AvailableTokens(); tokens != 0 {
		t.Errorf("Expected 0 tokens, got %d", tokens)
	}

	// Reset
	rl.Reset()

	// Should have all tokens back
	if tokens := rl.AvailableTokens(); tokens != 5 {
		t.Errorf("Expected 5 tokens after reset, got %d", tokens)
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	// Create limiter with 10 tokens per second
	rl := NewRateLimiter(10, 1*time.Second, true, false)

	// Acquire all tokens
	for i := 0; i < 10; i++ {
		rl.TryAcquire()
	}

	// Wait for some tokens to refill
	time.Sleep(500 * time.Millisecond)

	// Should have refilled approximately 5 tokens
	tokens := rl.AvailableTokens()
	if tokens < 4 || tokens > 6 {
		t.Errorf("Expected approximately 5 tokens after 500ms, got %d", tokens)
	}
}

func TestRateLimiter_EstimateWaitTime(t *testing.T) {
	rl := NewRateLimiter(10, 1*time.Second, true, false)

	// With tokens available, wait time should be 0
	if wait := rl.EstimateWaitTime(); wait != 0 {
		t.Errorf("Expected 0 wait time with tokens available, got %v", wait)
	}

	// Acquire all tokens
	for i := 0; i < 10; i++ {
		rl.TryAcquire()
	}

	// Wait time should be positive
	wait := rl.EstimateWaitTime()
	if wait <= 0 {
		t.Error("Expected positive wait time with no tokens available")
	}
}

func TestRateLimiter_Disabled(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Minute, false, false)

	// Should always succeed when disabled
	for i := 0; i < 100; i++ {
		if !rl.TryAcquire() {
			t.Error("Expected to always acquire token when disabled")
		}
	}

	// Available tokens should be -1 (unlimited)
	if tokens := rl.AvailableTokens(); tokens != -1 {
		t.Errorf("Expected -1 tokens when disabled, got %d", tokens)
	}

	// Estimate wait time should be 0
	if wait := rl.EstimateWaitTime(); wait != 0 {
		t.Errorf("Expected 0 wait time when disabled, got %v", wait)
	}
}

func TestRateLimiter_Concurrent(t *testing.T) {
	rl := NewRateLimiter(100, 1*time.Second, true, false)

	successCount := 0
	done := make(chan bool)

	// Try to acquire 150 tokens concurrently
	for i := 0; i < 150; i++ {
		go func() {
			if rl.TryAcquire() {
				successCount++
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 150; i++ {
		<-done
	}

	// Should have acquired approximately 100 tokens (rate limit)
	if successCount < 95 || successCount > 105 {
		t.Errorf("Expected approximately 100 successful acquisitions, got %d", successCount)
	}
}

package kra

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
//
// The rate limiter is goroutine-safe and uses the token bucket algorithm
// to control the rate of API requests.
type RateLimiter struct {
	maxTokens    int
	tokens       int
	refillRate   float64 // tokens per second
	lastRefill   time.Time
	mu           sync.Mutex
	enabled      bool
	debug        bool
	windowPeriod time.Duration
}

// NewRateLimiter creates a new rate limiter
//
// Parameters:
//   - maxRequests: Maximum number of requests allowed in the window
//   - window: Time window for rate limiting (e.g., 1 minute)
//   - enabled: Whether rate limiting is enabled
//   - debug: Whether to log rate limiting operations
//
// Example:
//
//	limiter := NewRateLimiter(100, 1*time.Minute, true, false)
func NewRateLimiter(maxRequests int, window time.Duration, enabled bool, debug bool) *RateLimiter {
	if !enabled {
		return &RateLimiter{enabled: false}
	}

	refillRate := float64(maxRequests) / window.Seconds()

	return &RateLimiter{
		maxTokens:    maxRequests,
		tokens:       maxRequests,
		refillRate:   refillRate,
		lastRefill:   time.Now(),
		enabled:      enabled,
		debug:        debug,
		windowPeriod: window,
	}
}

// Wait blocks until a token is available
//
// This method will block the current goroutine until a token becomes available.
// Use this method when you want to ensure the request waits for rate limiting.
//
// Example:
//
//	limiter.Wait()
//	// Proceed with API request
func (rl *RateLimiter) Wait() {
	if !rl.enabled {
		return
	}

	for {
		if rl.tryAcquire() {
			return
		}

		// Calculate how long to wait for next token (inline to avoid deadlock)
		timePerToken := time.Second / time.Duration(rl.refillRate)
		waitDuration := timePerToken + (10 * time.Millisecond)

		if rl.debug {
			fmt.Printf("[RateLimit] WAIT: Sleeping for %v\n", waitDuration)
		}
		time.Sleep(waitDuration)
	}
}

// TryAcquire attempts to acquire a token without blocking
//
// Returns true if a token was acquired, false if no tokens are available.
// Use this method when you want to check rate limits without waiting.
//
// Example:
//
//	if limiter.TryAcquire() {
//	    // Proceed with API request
//	} else {
//	    // Handle rate limit exceeded
//	    return RateLimitExceededError
//	}
func (rl *RateLimiter) TryAcquire() bool {
	if !rl.enabled {
		return true
	}

	return rl.tryAcquire()
}

// tryAcquire internal method that attempts to acquire a token
func (rl *RateLimiter) tryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens > 0 {
		rl.tokens--
		if rl.debug {
			fmt.Printf("[RateLimit] ACQUIRE: Token acquired (remaining: %d/%d)\n", rl.tokens, rl.maxTokens)
		}
		return true
	}

	if rl.debug {
		fmt.Printf("[RateLimit] EXCEED: No tokens available (0/%d)\n", rl.maxTokens)
	}
	return false
}

// refill adds tokens based on elapsed time since last refill
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	tokensToAdd := int(elapsed * rl.refillRate)

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now

		if rl.debug && tokensToAdd > 0 {
			fmt.Printf("[RateLimit] REFILL: Added %d tokens (now: %d/%d)\n", tokensToAdd, rl.tokens, rl.maxTokens)
		}
	}
}

// AvailableTokens returns the current number of available tokens
//
// This is useful for monitoring rate limit status.
func (rl *RateLimiter) AvailableTokens() int {
	if !rl.enabled {
		return -1 // Indicate unlimited
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()
	return rl.tokens
}

// Reset resets the rate limiter to full capacity
//
// This is useful for testing or when you want to clear rate limit state.
func (rl *RateLimiter) Reset() {
	if !rl.enabled {
		return
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tokens = rl.maxTokens
	rl.lastRefill = time.Now()

	if rl.debug {
		fmt.Printf("[RateLimit] RESET: Tokens reset to %d/%d\n", rl.tokens, rl.maxTokens)
	}
}

// EstimateWaitTime estimates how long it would take to acquire a token
//
// Returns 0 if tokens are available, otherwise returns estimated wait duration.
func (rl *RateLimiter) EstimateWaitTime() time.Duration {
	if !rl.enabled {
		return 0
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens > 0 {
		return 0
	}

	// Time to generate one token (inline to avoid deadlock)
	timePerToken := time.Second / time.Duration(rl.refillRate)

	// Add a small buffer to ensure token is available
	return timePerToken + (10 * time.Millisecond)
}

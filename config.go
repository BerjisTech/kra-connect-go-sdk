package kra

import (
	"time"
)

// Config holds the configuration for the KRA Connect client
type Config struct {
	// API configuration
	APIKey  string
	BaseURL string
	Timeout time.Duration

	// Retry configuration
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration

	// Rate limiting configuration
	RateLimitEnabled bool
	MaxRequests      int
	RateLimitWindow  time.Duration

	// Cache configuration
	CacheEnabled       bool
	PINVerificationTTL time.Duration
	TCCVerificationTTL time.Duration
	EslipValidationTTL time.Duration
	TaxpayerDetailsTTL time.Duration
	NILReturnTTL       time.Duration
	CacheMaxEntries    int

	// Debug configuration
	DebugMode bool
}

// Option is a functional option for configuring the KRA Connect client
type Option func(*Config) error

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://api.kra.go.ke/gavaconnect/v1",
		Timeout: 30 * time.Second,

		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     32 * time.Second,

		RateLimitEnabled: true,
		MaxRequests:      100,
		RateLimitWindow:  1 * time.Minute,

		CacheEnabled:       true,
		PINVerificationTTL: 1 * time.Hour,
		TCCVerificationTTL: 30 * time.Minute,
		EslipValidationTTL: 15 * time.Minute,
		TaxpayerDetailsTTL: 2 * time.Hour,
		NILReturnTTL:       24 * time.Hour,
		CacheMaxEntries:    1024,

		DebugMode: false,
	}
}

// WithAPIKey sets the API key for authentication
//
// The API key is required and must be at least 16 characters long.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key-here"),
//	)
func WithAPIKey(apiKey string) Option {
	return func(c *Config) error {
		if err := ValidateAPIKey(apiKey); err != nil {
			return err
		}
		c.APIKey = apiKey
		return nil
	}
}

// WithBaseURL sets the base URL for the KRA API
//
// Default: https://api.kra.go.ke/gavaconnect/v1
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithBaseURL("https://sandbox.kra.go.ke/gavaconnect/v1"),
//	)
func WithBaseURL(baseURL string) Option {
	return func(c *Config) error {
		if baseURL == "" {
			return NewValidationError("base_url", "Base URL cannot be empty")
		}
		c.BaseURL = baseURL
		return nil
	}
}

// WithTimeout sets the HTTP request timeout
//
// Default: 30 seconds
// Valid range: 1 second to 10 minutes
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithTimeout(60 * time.Second),
//	)
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) error {
		if err := ValidateTimeout(timeout); err != nil {
			return err
		}
		c.Timeout = timeout
		return nil
	}
}

// WithRetry configures retry behavior for failed requests
//
// Default: maxRetries=3, initialDelay=1s, maxDelay=32s
//
// The retry mechanism uses exponential backoff with jitter.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithRetry(5, 2*time.Second, 60*time.Second),
//	)
func WithRetry(maxRetries int, initialDelay, maxDelay time.Duration) Option {
	return func(c *Config) error {
		if err := ValidateRetryConfig(maxRetries, initialDelay, maxDelay); err != nil {
			return err
		}
		c.MaxRetries = maxRetries
		c.InitialDelay = initialDelay
		c.MaxDelay = maxDelay
		return nil
	}
}

// WithRateLimit configures rate limiting for API requests
//
// Default: enabled=true, maxRequests=100, window=1 minute
//
// Rate limiting uses a token bucket algorithm to prevent exceeding API limits.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithRateLimit(200, 1*time.Minute),
//	)
func WithRateLimit(maxRequests int, window time.Duration) Option {
	return func(c *Config) error {
		if err := ValidateRateLimitConfig(maxRequests, window); err != nil {
			return err
		}
		c.RateLimitEnabled = true
		c.MaxRequests = maxRequests
		c.RateLimitWindow = window
		return nil
	}
}

// WithoutRateLimit disables rate limiting
//
// Use this option if you have your own rate limiting mechanism
// or if the API you're connecting to doesn't have rate limits.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithoutRateLimit(),
//	)
func WithoutRateLimit() Option {
	return func(c *Config) error {
		c.RateLimitEnabled = false
		return nil
	}
}

// WithCache enables caching with custom TTL values
//
// Default TTLs:
//   - PIN verification: 1 hour
//   - TCC verification: 30 minutes
//   - E-slip validation: 15 minutes
//   - Taxpayer details: 2 hours
//   - NIL return: 24 hours
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithCache(true, 2*time.Hour),
//	)
func WithCache(enabled bool, defaultTTL time.Duration) Option {
	return func(c *Config) error {
		if enabled && defaultTTL > 0 {
			if err := ValidateCacheTTL(defaultTTL); err != nil {
				return err
			}
		}
		c.CacheEnabled = enabled
		if defaultTTL > 0 {
			c.PINVerificationTTL = defaultTTL
			c.TCCVerificationTTL = defaultTTL
			c.EslipValidationTTL = defaultTTL
			c.TaxpayerDetailsTTL = defaultTTL
			c.NILReturnTTL = defaultTTL
		}
		return nil
	}
}

// WithCacheCapacity sets the maximum number of entries the cache can hold before evicting
//
// Default: 1024 entries
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithCacheCapacity(2048),
//	)
func WithCacheCapacity(entries int) Option {
	return func(c *Config) error {
		if entries <= 0 {
			return NewValidationError("cache_max_entries", "Cache max entries must be positive")
		}
		c.CacheMaxEntries = entries
		return nil
	}
}

// WithCustomCacheTTLs sets custom TTL values for each operation type
//
// This allows fine-grained control over cache duration for different operations.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithCustomCacheTTLs(
//	        2*time.Hour,    // PIN verification
//	        1*time.Hour,    // TCC verification
//	        30*time.Minute, // E-slip validation
//	        4*time.Hour,    // Taxpayer details
//	        48*time.Hour,   // NIL return
//	    ),
//	)
func WithCustomCacheTTLs(
	pinTTL, tccTTL, eslipTTL, taxpayerTTL, nilReturnTTL time.Duration,
) Option {
	return func(c *Config) error {
		if err := ValidateCacheTTL(pinTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(tccTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(eslipTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(taxpayerTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(nilReturnTTL); err != nil {
			return err
		}

		c.PINVerificationTTL = pinTTL
		c.TCCVerificationTTL = tccTTL
		c.EslipValidationTTL = eslipTTL
		c.TaxpayerDetailsTTL = taxpayerTTL
		c.NILReturnTTL = nilReturnTTL

		return nil
	}
}

// WithoutCache disables caching
//
// Use this option if you want to always get fresh data from the API
// or if you have your own caching mechanism.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithoutCache(),
//	)
func WithoutCache() Option {
	return func(c *Config) error {
		c.CacheEnabled = false
		return nil
	}
}

// WithDebug enables debug mode
//
// In debug mode, the client logs detailed information about requests,
// responses, retries, cache hits/misses, and rate limiting.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithDebug(true),
//	)
func WithDebug(enabled bool) Option {
	return func(c *Config) error {
		c.DebugMode = enabled
		return nil
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := ValidateAPIKey(c.APIKey); err != nil {
		return err
	}

	if c.BaseURL == "" {
		return NewValidationError("base_url", "Base URL is required")
	}

	if err := ValidateTimeout(c.Timeout); err != nil {
		return err
	}

	if err := ValidateRetryConfig(c.MaxRetries, c.InitialDelay, c.MaxDelay); err != nil {
		return err
	}

	if c.RateLimitEnabled {
		if err := ValidateRateLimitConfig(c.MaxRequests, c.RateLimitWindow); err != nil {
			return err
		}
	}

	if c.CacheEnabled {
		if c.CacheMaxEntries <= 0 {
			return NewValidationError("cache_max_entries", "Cache max entries must be positive")
		}
		if err := ValidateCacheTTL(c.PINVerificationTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(c.TCCVerificationTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(c.EslipValidationTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(c.TaxpayerDetailsTTL); err != nil {
			return err
		}
		if err := ValidateCacheTTL(c.NILReturnTTL); err != nil {
			return err
		}
	}

	return nil
}

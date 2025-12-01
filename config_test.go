package kra

import (
	"strings"
	"testing"
	"time"
)

func TestWithCacheCapacity(t *testing.T) {
	if _, err := NewClient(
		WithAPIKey(strings.Repeat("A", 16)),
		WithCacheCapacity(2048),
	); err != nil {
		t.Fatalf("expected cache capacity option to succeed, got %v", err)
	}

	_, err := NewClient(
		WithAPIKey(strings.Repeat("A", 16)),
		WithCacheCapacity(0),
	)
	if err == nil {
		t.Fatalf("expected error for zero cache capacity")
	}
}

func TestConfigValidateCacheSettings(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = strings.Repeat("B", 16)
	cfg.CacheEnabled = true
	cfg.CacheMaxEntries = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for negative cache entries")
	}

	cfg.CacheMaxEntries = 10
	cfg.PINVerificationTTL = -1
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for negative TTL")
	}

	cfg.PINVerificationTTL = time.Hour
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected validation to pass, got %v", err)
	}
}

func TestConfigOptionsCoverage(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = strings.Repeat("C", 16)

	options := []Option{
		WithTimeout(45 * time.Second),
		WithRetry(2, 10*time.Millisecond, time.Second),
		WithRateLimit(200, time.Minute),
		WithCache(true, 15*time.Minute),
		WithCustomCacheTTLs(
			10*time.Minute,
			11*time.Minute,
			12*time.Minute,
			13*time.Minute,
			14*time.Minute,
		),
		WithDebug(true),
	}

	for _, opt := range options {
		if err := opt(cfg); err != nil {
			t.Fatalf("option returned error: %v", err)
		}
	}

	if !cfg.DebugMode {
		t.Fatal("expected debug mode to be enabled")
	}

	if err := WithoutCache()(cfg); err != nil {
		t.Fatalf("WithoutCache() error = %v", err)
	}

	if cfg.CacheEnabled {
		t.Fatal("expected cache to be disabled")
	}
}

func TestConfigValidateSuccess(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = strings.Repeat("E", 16)
	cfg.RateLimitEnabled = true
	cfg.MaxRequests = 25
	cfg.RateLimitWindow = time.Second
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateErrors(t *testing.T) {
	invalidAPI := DefaultConfig()
	if err := invalidAPI.Validate(); err == nil {
		t.Fatal("expected error for missing API key")
	}

	timeoutCfg := DefaultConfig()
	timeoutCfg.APIKey = strings.Repeat("F", 16)
	timeoutCfg.Timeout = 0
	if err := timeoutCfg.Validate(); err == nil {
		t.Fatal("expected timeout validation error")
	}

	rateCfg := DefaultConfig()
	rateCfg.APIKey = strings.Repeat("G", 16)
	rateCfg.MaxRequests = 0
	if err := rateCfg.Validate(); err == nil {
		t.Fatal("expected rate limit validation error")
	}

	cacheCfg := DefaultConfig()
	cacheCfg.APIKey = strings.Repeat("H", 16)
	cacheCfg.PINVerificationTTL = -time.Minute
	if err := cacheCfg.Validate(); err == nil {
		t.Fatal("expected cache TTL validation error")
	}
}

func TestWithCustomCacheTTLsInvalid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = strings.Repeat("I", 16)
	err := WithCustomCacheTTLs(-time.Minute, time.Minute, time.Minute, time.Minute, time.Minute)(cfg)
	if err == nil {
		t.Fatal("expected error for negative PIN TTL")
	}
}

func TestConfigValidateTTLErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Config)
	}{
		{"tcc", func(c *Config) { c.TCCVerificationTTL = -time.Minute }},
		{"eslip", func(c *Config) { c.EslipValidationTTL = -time.Minute }},
		{"taxpayer", func(c *Config) { c.TaxpayerDetailsTTL = -time.Minute }},
		{"nil return", func(c *Config) { c.NILReturnTTL = -time.Minute }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.APIKey = strings.Repeat("Z", 16)
			tc.mutate(cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected error for %s TTL", tc.name)
			}
		})
	}
}

func TestConfigOptionValidationErrors(t *testing.T) {
	cfg := DefaultConfig()
	if err := WithAPIKey("short")(cfg); err == nil {
		t.Fatal("expected WithAPIKey to fail for short key")
	}

	if err := WithBaseURL("")(cfg); err == nil {
		t.Fatal("expected WithBaseURL to fail for empty URL")
	}

	if err := WithTimeout(0)(cfg); err == nil {
		t.Fatal("expected WithTimeout to fail for zero duration")
	}
}

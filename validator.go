package kra

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Regular expressions for validation
var (
	// PIN format: P followed by 9 digits and a letter (e.g., P051234567A)
	pinRegex = regexp.MustCompile(`^P\d{9}[A-Z]$`)

	// TCC format: TCC followed by digits (e.g., TCC123456)
	tccRegex = regexp.MustCompile(`^TCC\d+$`)

	// E-slip format: digits only
	eslipRegex = regexp.MustCompile(`^\d+$`)

	// Period format: YYYYMM (e.g., 202401)
	periodRegex = regexp.MustCompile(`^\d{6}$`)

	// Obligation ID format: alphanumeric with optional hyphens/underscores
	obligationIDRegex = regexp.MustCompile(`^[A-Z0-9_-]+$`)
)

// ValidateAndNormalizePIN validates and normalizes a PIN number
//
// PIN format: P followed by 9 digits and a letter (e.g., P051234567A)
// The function converts the PIN to uppercase and removes whitespace.
//
// Returns the normalized PIN or an error if validation fails.
func ValidateAndNormalizePIN(pin string) (string, error) {
	if pin == "" {
		return "", NewValidationError("pin", "PIN number is required")
	}

	// Normalize: trim whitespace and convert to uppercase
	normalized := strings.ToUpper(strings.TrimSpace(pin))

	// Validate format
	if !pinRegex.MatchString(normalized) {
		return "", NewInvalidPINFormatError(pin)
	}

	return normalized, nil
}

// ValidateAndNormalizeTCC validates and normalizes a TCC number
//
// TCC format: TCC followed by digits (e.g., TCC123456)
// The function converts the TCC to uppercase and removes whitespace.
//
// Returns the normalized TCC or an error if validation fails.
func ValidateAndNormalizeTCC(tcc string) (string, error) {
	if tcc == "" {
		return "", NewValidationError("tcc", "TCC number is required")
	}

	// Normalize: trim whitespace and convert to uppercase
	normalized := strings.ToUpper(strings.TrimSpace(tcc))

	// Validate format
	if !tccRegex.MatchString(normalized) {
		return "", NewInvalidTCCFormatError(tcc)
	}

	return normalized, nil
}

// ValidateEslipNumber validates an e-slip number
//
// E-slip format: digits only
//
// Returns an error if validation fails.
func ValidateEslipNumber(eslip string) error {
	if eslip == "" {
		return NewValidationError("eslip", "E-slip number is required")
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(eslip)

	// Validate format
	if !eslipRegex.MatchString(trimmed) {
		return NewValidationError("eslip", fmt.Sprintf("Invalid e-slip format: '%s'. Expected digits only", eslip))
	}

	return nil
}

// ValidatePeriod validates a tax period string
//
// Period format: YYYYMM (e.g., 202401)
// Validates that the year is reasonable (1900-2100) and month is 01-12.
//
// Returns an error if validation fails.
func ValidatePeriod(period string) error {
	if period == "" {
		return NewValidationError("period", "Period is required")
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(period)

	// Validate format
	if !periodRegex.MatchString(trimmed) {
		return NewValidationError("period", fmt.Sprintf("Invalid period format: '%s'. Expected YYYYMM (e.g., 202401)", period))
	}

	// Parse and validate date components
	year := trimmed[:4]
	month := trimmed[4:]

	// Validate year range
	if year < "1900" || year > "2100" {
		return NewValidationError("period", fmt.Sprintf("Invalid year in period: %s. Year must be between 1900 and 2100", year))
	}

	// Validate month range
	if month < "01" || month > "12" {
		return NewValidationError("period", fmt.Sprintf("Invalid month in period: %s. Month must be between 01 and 12", month))
	}

	return nil
}

// ValidateObligationID validates an obligation ID
//
// Obligation ID format: alphanumeric with optional hyphens/underscores
//
// Returns an error if validation fails.
func ValidateObligationID(obligationID string) error {
	if obligationID == "" {
		return NewValidationError("obligation_id", "Obligation ID is required")
	}

	// Trim whitespace and convert to uppercase
	normalized := strings.ToUpper(strings.TrimSpace(obligationID))

	// Validate format
	if !obligationIDRegex.MatchString(normalized) {
		return NewValidationError("obligation_id", fmt.Sprintf("Invalid obligation ID format: '%s'. Expected alphanumeric with optional hyphens/underscores", obligationID))
	}

	return nil
}

// ValidateAPIKey validates an API key
//
// API key should be non-empty and at least 16 characters long for security.
//
// Returns an error if validation fails.
func ValidateAPIKey(apiKey string) error {
	if apiKey == "" {
		return NewValidationError("api_key", "API key is required")
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(apiKey)

	if len(trimmed) < 16 {
		return NewValidationError("api_key", "API key must be at least 16 characters long")
	}

	return nil
}

// ValidateTimeout validates a timeout duration
//
// Timeout should be positive and reasonable (between 1 second and 10 minutes).
//
// Returns an error if validation fails.
func ValidateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return NewValidationError("timeout", "Timeout must be positive")
	}

	if timeout > 10*time.Minute {
		return NewValidationError("timeout", "Timeout cannot exceed 10 minutes")
	}

	return nil
}

// ValidateRetryConfig validates retry configuration
//
// Returns an error if validation fails.
func ValidateRetryConfig(maxRetries int, initialDelay, maxDelay time.Duration) error {
	if maxRetries < 0 {
		return NewValidationError("max_retries", "Max retries cannot be negative")
	}

	if maxRetries > 10 {
		return NewValidationError("max_retries", "Max retries cannot exceed 10")
	}

	if initialDelay <= 0 {
		return NewValidationError("initial_delay", "Initial delay must be positive")
	}

	if maxDelay <= 0 {
		return NewValidationError("max_delay", "Max delay must be positive")
	}

	if maxDelay < initialDelay {
		return NewValidationError("max_delay", "Max delay cannot be less than initial delay")
	}

	return nil
}

// ValidateRateLimitConfig validates rate limit configuration
//
// Returns an error if validation fails.
func ValidateRateLimitConfig(maxRequests int, window time.Duration) error {
	if maxRequests <= 0 {
		return NewValidationError("max_requests", "Max requests must be positive")
	}

	if window <= 0 {
		return NewValidationError("window", "Window duration must be positive")
	}

	return nil
}

// ValidateCacheTTL validates cache TTL duration
//
// Returns an error if validation fails.
func ValidateCacheTTL(ttl time.Duration) error {
	if ttl < 0 {
		return NewValidationError("cache_ttl", "Cache TTL cannot be negative")
	}

	if ttl > 24*time.Hour {
		return NewValidationError("cache_ttl", "Cache TTL cannot exceed 24 hours")
	}

	return nil
}

package kra

import (
	"fmt"
	"time"
)

// Error types for the KRA Connect SDK

// SDKError is the base error type for all SDK errors.
type SDKError struct {
	Message    string
	Details    map[string]interface{}
	StatusCode int
	Err        error
}

func (e *SDKError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *SDKError) Unwrap() error {
	return e.Err
}

// ValidationError represents input validation errors
type ValidationError struct {
	SDKError
	Field string
}

// NewValidationError constructs a validation error for a given field.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		SDKError: SDKError{
			Message: message,
			Details: map[string]interface{}{
				"field": field,
			},
		},
		Field: field,
	}
}

// InvalidPINFormatError represents an invalid PIN format error.
type InvalidPINFormatError struct {
	ValidationError
	PIN string
}

// NewInvalidPINFormatError constructs an error for invalid PIN formats.
func NewInvalidPINFormatError(pin string) *InvalidPINFormatError {
	base := NewValidationError(
		"pin",
		fmt.Sprintf("Invalid PIN format: '%s'. Expected format: P followed by 9 digits and a letter (e.g., P051234567A)", pin),
	)

	return &InvalidPINFormatError{
		ValidationError: *base,
		PIN:             pin,
	}
}

// InvalidTCCFormatError represents an invalid TCC format error.
type InvalidTCCFormatError struct {
	ValidationError
	TCC string
}

// NewInvalidTCCFormatError constructs an error for invalid TCC formats.
func NewInvalidTCCFormatError(tcc string) *InvalidTCCFormatError {
	base := NewValidationError(
		"tcc",
		fmt.Sprintf("Invalid TCC format: '%s'. Expected format: TCC followed by digits (e.g., TCC123456)", tcc),
	)

	return &InvalidTCCFormatError{
		ValidationError: *base,
		TCC:             tcc,
	}
}

// AuthenticationError represents API authentication failures
type AuthenticationError struct {
	SDKError
}

// NewAuthenticationError constructs an authentication error.
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		SDKError: SDKError{
			Message:    message,
			StatusCode: 401,
		},
	}
}

// RateLimitError represents rate limit exceeded errors
type RateLimitError struct {
	SDKError
	RetryAfter time.Duration
	Limit      int
	Window     time.Duration
}

// NewRateLimitError constructs a rate limit error with retry hints.
func NewRateLimitError(retryAfter time.Duration, limit int, window time.Duration) *RateLimitError {
	message := fmt.Sprintf("Rate limit exceeded. Retry after %v", retryAfter)
	if limit > 0 {
		message += fmt.Sprintf(" (Limit: %d requests per %v)", limit, window)
	}

	return &RateLimitError{
		SDKError: SDKError{
			Message:    message,
			StatusCode: 429,
			Details: map[string]interface{}{
				"retry_after": retryAfter,
				"limit":       limit,
				"window":      window,
			},
		},
		RetryAfter: retryAfter,
		Limit:      limit,
		Window:     window,
	}
}

// TimeoutError represents request timeout errors
type TimeoutError struct {
	SDKError
	Endpoint      string
	Timeout       time.Duration
	AttemptNumber int
}

// NewTimeoutError constructs a timeout error for a given endpoint.
func NewTimeoutError(endpoint string, timeout time.Duration, attemptNumber int) *TimeoutError {
	return &TimeoutError{
		SDKError: SDKError{
			Message:    fmt.Sprintf("Request to '%s' timed out after %v (attempt %d)", endpoint, timeout, attemptNumber),
			StatusCode: 408,
			Details: map[string]interface{}{
				"endpoint":       endpoint,
				"timeout":        timeout,
				"attempt_number": attemptNumber,
			},
		},
		Endpoint:      endpoint,
		Timeout:       timeout,
		AttemptNumber: attemptNumber,
	}
}

// APIError represents general API errors
type APIError struct {
	SDKError
	Endpoint     string
	ResponseBody string
}

// NewAPIError constructs a generic API error for non-timeout failures.
func NewAPIError(statusCode int, message, endpoint, responseBody string) *APIError {
	return &APIError{
		SDKError: SDKError{
			Message:    message,
			StatusCode: statusCode,
			Details: map[string]interface{}{
				"endpoint":      endpoint,
				"response_body": responseBody,
			},
		},
		Endpoint:     endpoint,
		ResponseBody: responseBody,
	}
}

// IsServerError returns true if the error is a server error (5xx)
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsClientError returns true if the error is a client error (4xx)
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// NetworkError represents network-related errors
type NetworkError struct {
	SDKError
	Endpoint string
}

// NewNetworkError constructs an error for network failures.
func NewNetworkError(endpoint string, err error) *NetworkError {
	return &NetworkError{
		SDKError: SDKError{
			Message: fmt.Sprintf("Network error while calling '%s'", endpoint),
			Err:     err,
			Details: map[string]interface{}{
				"endpoint": endpoint,
			},
		},
		Endpoint: endpoint,
	}
}

// CacheError represents cache operation errors
type CacheError struct {
	SDKError
	Operation string
	Key       string
}

// NewCacheError constructs an error for cache read/write failures.
func NewCacheError(operation, key, reason string) *CacheError {
	return &CacheError{
		SDKError: SDKError{
			Message: fmt.Sprintf("Cache %s failed for key '%s': %s", operation, key, reason),
			Details: map[string]interface{}{
				"operation": operation,
				"key":       key,
				"reason":    reason,
			},
		},
		Operation: operation,
		Key:       key,
	}
}

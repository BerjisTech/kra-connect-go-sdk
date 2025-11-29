package kra

import (
	"fmt"
	"time"
)

// Error types for the KRA Connect SDK

// KRAError is the base error type for all SDK errors
type KRAError struct {
	Message    string
	Details    map[string]interface{}
	StatusCode int
	Err        error
}

func (e *KRAError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *KRAError) Unwrap() error {
	return e.Err
}

// ValidationError represents input validation errors
type ValidationError struct {
	KRAError
	Field string
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		KRAError: KRAError{
			Message: message,
			Details: map[string]interface{}{
				"field": field,
			},
		},
		Field: field,
	}
}

// InvalidPINFormatError represents an invalid PIN format error
type InvalidPINFormatError struct {
	ValidationError
	PIN string
}

func NewInvalidPINFormatError(pin string) *InvalidPINFormatError {
	return &InvalidPINFormatError{
		ValidationError: ValidationError{
			KRAError: KRAError{
				Message: fmt.Sprintf("Invalid PIN format: '%s'. Expected format: P followed by 9 digits and a letter (e.g., P051234567A)", pin),
				Details: map[string]interface{}{
					"pin": pin,
				},
			},
			Field: "pin",
		},
		PIN: pin,
	}
}

// InvalidTCCFormatError represents an invalid TCC format error
type InvalidTCCFormatError struct {
	ValidationError
	TCC string
}

func NewInvalidTCCFormatError(tcc string) *InvalidTCCFormatError {
	return &InvalidTCCFormatError{
		ValidationError: ValidationError{
			KRAError: KRAError{
				Message: fmt.Sprintf("Invalid TCC format: '%s'. Expected format: TCC followed by digits (e.g., TCC123456)", tcc),
				Details: map[string]interface{}{
					"tcc": tcc,
				},
			},
			Field: "tcc",
		},
		TCC: tcc,
	}
}

// AuthenticationError represents API authentication failures
type AuthenticationError struct {
	KRAError
}

func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		KRAError: KRAError{
			Message:    message,
			StatusCode: 401,
		},
	}
}

// RateLimitError represents rate limit exceeded errors
type RateLimitError struct {
	KRAError
	RetryAfter time.Duration
	Limit      int
	Window     time.Duration
}

func NewRateLimitError(retryAfter time.Duration, limit int, window time.Duration) *RateLimitError {
	message := fmt.Sprintf("Rate limit exceeded. Retry after %v", retryAfter)
	if limit > 0 {
		message += fmt.Sprintf(" (Limit: %d requests per %v)", limit, window)
	}

	return &RateLimitError{
		KRAError: KRAError{
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
	KRAError
	Endpoint      string
	Timeout       time.Duration
	AttemptNumber int
}

func NewTimeoutError(endpoint string, timeout time.Duration, attemptNumber int) *TimeoutError {
	return &TimeoutError{
		KRAError: KRAError{
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
	KRAError
	Endpoint     string
	ResponseBody string
}

func NewAPIError(statusCode int, message, endpoint, responseBody string) *APIError {
	return &APIError{
		KRAError: KRAError{
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
	KRAError
	Endpoint string
}

func NewNetworkError(endpoint string, err error) *NetworkError {
	return &NetworkError{
		KRAError: KRAError{
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
	KRAError
	Operation string
	Key       string
}

func NewCacheError(operation, key, reason string) *CacheError {
	return &CacheError{
		KRAError: KRAError{
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

package kra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// HTTPClient handles HTTP communication with the KRA API
type HTTPClient struct {
	client       *http.Client
	config       *Config
	rateLimiter  *RateLimiter
	cacheManager *CacheManager
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(config *Config, rateLimiter *RateLimiter, cacheManager *CacheManager) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config:       config,
		rateLimiter:  rateLimiter,
		cacheManager: cacheManager,
	}
}

// apiRequest represents a request to the KRA API
type apiRequest struct {
	Method   string
	Endpoint string
	Body     interface{}
	Headers  map[string]string
}

// apiResponse represents the structure of KRA API responses
type apiResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   *apiErrorResponse      `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// apiErrorResponse represents error details in API responses
type apiErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Post sends a POST request to the API with retry logic
func (h *HTTPClient) Post(ctx context.Context, endpoint string, body interface{}) (map[string]interface{}, error) {
	req := &apiRequest{
		Method:   "POST",
		Endpoint: endpoint,
		Body:     body,
	}

	return h.executeWithRetry(ctx, req)
}

// Get sends a GET request to the API with retry logic
func (h *HTTPClient) Get(ctx context.Context, endpoint string) (map[string]interface{}, error) {
	req := &apiRequest{
		Method:   "GET",
		Endpoint: endpoint,
	}

	return h.executeWithRetry(ctx, req)
}

// executeWithRetry executes a request with exponential backoff retry logic
func (h *HTTPClient) executeWithRetry(ctx context.Context, req *apiRequest) (map[string]interface{}, error) {
	var lastErr error
	delay := h.config.InitialDelay

	for attempt := 0; attempt <= h.config.MaxRetries; attempt++ {
		// Check if context is cancelled
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		// Wait for rate limiter
		if !h.waitForRateLimit(ctx) {
			return nil, ctx.Err()
		}

		// Execute the request
		response, err := h.execute(ctx, req, attempt+1)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on client errors (4xx) except 429 (rate limit)
		if apiErr, ok := err.(*APIError); ok {
			if apiErr.IsClientError() && apiErr.StatusCode != 429 {
				return nil, err
			}
		}

		// Don't retry on validation errors
		if _, ok := err.(*ValidationError); ok {
			return nil, err
		}

		// Don't retry on authentication errors
		if _, ok := err.(*AuthenticationError); ok {
			return nil, err
		}

		// Last attempt - don't wait
		if attempt >= h.config.MaxRetries {
			break
		}

		// Log retry attempt
		if h.config.DebugMode {
			fmt.Printf("[HTTP] RETRY: Attempt %d/%d for %s after error: %v\n",
				attempt+1, h.config.MaxRetries+1, req.Endpoint, err)
		}

		// Calculate backoff with jitter
		backoff := h.calculateBackoff(delay, attempt)

		// Wait with context cancellation support
		select {
		case <-time.After(backoff):
			// Continue to next retry
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		// Exponential backoff for next iteration
		delay = time.Duration(float64(delay) * 2)
		if delay > h.config.MaxDelay {
			delay = h.config.MaxDelay
		}
	}

	return nil, lastErr
}

// execute sends a single HTTP request
func (h *HTTPClient) execute(ctx context.Context, apiReq *apiRequest, attemptNumber int) (map[string]interface{}, error) {
	// Build full URL
	url := h.config.BaseURL + apiReq.Endpoint

	// Create request body
	var bodyReader io.Reader
	if apiReq.Body != nil {
		jsonBody, err := json.Marshal(apiReq.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, apiReq.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+h.config.APIKey)
	httpReq.Header.Set("User-Agent", "KRA-Connect-Go-SDK/0.1.1")

	// Add custom headers
	for key, value := range apiReq.Headers {
		httpReq.Header.Set(key, value)
	}

	// Log request
	if h.config.DebugMode {
		fmt.Printf("[HTTP] REQUEST: %s %s (attempt %d)\n", apiReq.Method, url, attemptNumber)
	}

	// Send request
	startTime := time.Now()
	httpResp, err := h.client.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		if h.config.DebugMode {
			fmt.Printf("[HTTP] ERROR: Request failed after %v: %v\n", duration, err)
		}
		return nil, NewNetworkError(apiReq.Endpoint, err)
	}
	defer httpResp.Body.Close()

	// Log response
	if h.config.DebugMode {
		fmt.Printf("[HTTP] RESPONSE: %d in %v\n", httpResp.StatusCode, duration)
	}

	// Read response body
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Handle non-200 status codes
	if httpResp.StatusCode != http.StatusOK {
		return nil, h.handleErrorResponse(httpResp.StatusCode, respBody, apiReq.Endpoint)
	}

	// Parse response
	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, NewAPIError(
			httpResp.StatusCode,
			"Failed to parse API response",
			apiReq.Endpoint,
			string(respBody),
		)
	}

	// Check API-level errors
	if !apiResp.Success {
		errorMsg := "API request failed"
		if apiResp.Error != nil {
			errorMsg = apiResp.Error.Message
		} else if apiResp.Message != "" {
			errorMsg = apiResp.Message
		}

		return nil, NewAPIError(
			httpResp.StatusCode,
			errorMsg,
			apiReq.Endpoint,
			string(respBody),
		)
	}

	return apiResp.Data, nil
}

// handleErrorResponse handles HTTP error responses
func (h *HTTPClient) handleErrorResponse(statusCode int, body []byte, endpoint string) error {
	bodyStr := string(body)

	// Try to parse error response
	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err == nil && apiResp.Error != nil {
		bodyStr = apiResp.Error.Message
		if apiResp.Error.Details != "" {
			bodyStr += ": " + apiResp.Error.Details
		}
	}

	// Handle specific status codes
	switch statusCode {
	case http.StatusUnauthorized:
		return NewAuthenticationError("Authentication failed. Please check your API key.")

	case http.StatusForbidden:
		return NewAuthenticationError("Access forbidden. Your API key may not have the required permissions.")

	case http.StatusTooManyRequests:
		// Try to extract retry-after from response
		retryAfter := 60 * time.Second
		return NewRateLimitError(retryAfter, h.config.MaxRequests, h.config.RateLimitWindow)

	case http.StatusRequestTimeout:
		return NewTimeoutError(endpoint, h.config.Timeout, 1)

	case http.StatusBadRequest:
		return NewAPIError(statusCode, "Bad request: "+bodyStr, endpoint, bodyStr)

	case http.StatusNotFound:
		return NewAPIError(statusCode, "Endpoint not found: "+endpoint, endpoint, bodyStr)

	default:
		return NewAPIError(statusCode, bodyStr, endpoint, bodyStr)
	}
}

// waitForRateLimit waits for rate limiter with context support
func (h *HTTPClient) waitForRateLimit(ctx context.Context) bool {
	if !h.config.RateLimitEnabled {
		return true
	}

	// Try to acquire without blocking first
	if h.rateLimiter.TryAcquire() {
		return true
	}

	// Need to wait - check estimated wait time
	waitTime := h.rateLimiter.EstimateWaitTime()

	if h.config.DebugMode {
		fmt.Printf("[HTTP] RATE_LIMIT: Waiting %v for token\n", waitTime)
	}

	// Wait with context cancellation support
	select {
	case <-time.After(waitTime):
		h.rateLimiter.Wait()
		return true
	case <-ctx.Done():
		return false
	}
}

// calculateBackoff calculates backoff duration with jitter
func (h *HTTPClient) calculateBackoff(baseDelay time.Duration, attempt int) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	backoff := float64(baseDelay) * math.Pow(2, float64(attempt))

	// Cap at max delay
	if backoff > float64(h.config.MaxDelay) {
		backoff = float64(h.config.MaxDelay)
	}

	// Add jitter (±25%)
	jitter := backoff * 0.25 * (rand.Float64()*2 - 1)
	backoff += jitter

	// Ensure minimum delay of 100ms
	if backoff < 100 {
		backoff = 100
	}

	return time.Duration(backoff)
}

package kra

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClientRetriesOnServerError(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(t, w, apiResponse{
				Success: false,
				Error: &apiErrorResponse{
					Code:    "ERR",
					Message: "temporary",
				},
			})
			return
		}
		writeJSON(t, w, apiResponse{
			Success: true,
			Data:    map[string]interface{}{"is_valid": true},
		})
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.APIKey = "ABCDEFGHIJKLMNOP"
	cfg.BaseURL = server.URL
	cfg.RateLimitEnabled = false
	cfg.CacheEnabled = false
	cfg.InitialDelay = 10 * time.Millisecond
	cfg.MaxDelay = 10 * time.Millisecond

	rateLimiter := NewRateLimiter(cfg.MaxRequests, cfg.RateLimitWindow, cfg.RateLimitEnabled, cfg.DebugMode)
	cacheManager := NewCacheManager(cfg.CacheEnabled, cfg.DebugMode, cfg.CacheMaxEntries)
	client := NewHTTPClient(cfg, rateLimiter, cacheManager)
	client.client = server.Client()

	ctx := context.Background()
	data, err := client.Post(ctx, "/verify-pin", map[string]string{"pin": "P051234567A"})
	if err != nil {
		t.Fatalf("Post() error = %v", err)
	}
	if val, ok := data["is_valid"].(bool); !ok || !val {
		t.Fatalf("expected valid pin, got %v", data)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestHTTPClientHandleErrorResponse(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "ABCDEFGHIJKLMNOP"
	rateLimiter := NewRateLimiter(cfg.MaxRequests, cfg.RateLimitWindow, false, cfg.DebugMode)
	cacheManager := NewCacheManager(false, cfg.DebugMode, cfg.CacheMaxEntries)
	client := NewHTTPClient(cfg, rateLimiter, cacheManager)

	err := client.handleErrorResponse(http.StatusUnauthorized, []byte(`{"error":{"message":"bad"}}`), "/verify-pin")
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("expected AuthenticationError, got %v", err)
	}

	err = client.handleErrorResponse(http.StatusTooManyRequests, []byte(`{"error":{"message":"limit"}}`), "/verify-pin")
	if _, ok := err.(*RateLimitError); !ok {
		t.Fatalf("expected RateLimitError, got %v", err)
	}

	err = client.handleErrorResponse(http.StatusBadRequest, []byte(`{"error":{"message":"bad","details":"oops"}}`), "/verify-pin")
	if _, ok := err.(*APIError); !ok {
		t.Fatalf("expected APIError for bad request, got %v", err)
	}

	err = client.handleErrorResponse(http.StatusNotFound, []byte(`{}`), "/unknown")
	if _, ok := err.(*APIError); !ok {
		t.Fatalf("expected APIError for not found, got %v", err)
	}

	err = client.handleErrorResponse(http.StatusRequestTimeout, []byte(`{}`), "/slow")
	if _, ok := err.(*TimeoutError); !ok {
		t.Fatalf("expected TimeoutError, got %v", err)
	}
}

func TestHTTPClientWaitForRateLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "ABCDEFGHIJKLMNOP"
	cfg.RateLimitEnabled = true
	cfg.MaxRequests = 1
	cfg.RateLimitWindow = time.Millisecond * 50

	rateLimiter := NewRateLimiter(cfg.MaxRequests, cfg.RateLimitWindow, cfg.RateLimitEnabled, cfg.DebugMode)
	cacheManager := NewCacheManager(false, cfg.DebugMode, cfg.CacheMaxEntries)
	client := NewHTTPClient(cfg, rateLimiter, cacheManager)

	ctx := context.Background()
	if !client.waitForRateLimit(ctx) {
		t.Fatalf("expected first acquire to succeed")
	}
	if !client.waitForRateLimit(ctx) {
		t.Fatalf("expected second acquire to eventually succeed")
	}
}

func TestHTTPClientInvalidJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}
	client, server := newClientWithServer(t, handler, WithoutCache())
	defer server.Close()

	ctx := context.Background()
	if _, err := client.httpClient.Post(ctx, "/invalid", map[string]string{}); err == nil {
		t.Fatal("expected JSON parse error")
	}
}

func TestHTTPClientAPIFailure(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, apiResponse{
			Success: false,
			Error: &apiErrorResponse{
				Message: "API failed",
			},
		})
	}
	client, server := newClientWithServer(t, handler, WithoutCache())
	defer server.Close()

	ctx := context.Background()
	if _, err := client.httpClient.Post(ctx, "/fail", map[string]string{}); err == nil {
		t.Fatal("expected API error")
	}
}

func TestHTTPClientCalculateBackoff(t *testing.T) {
	cfg := DefaultConfig()
	cfg.APIKey = "ABCDEFGHIJKLMNOP"
	rateLimiter := NewRateLimiter(cfg.MaxRequests, cfg.RateLimitWindow, false, cfg.DebugMode)
	cacheManager := NewCacheManager(false, cfg.DebugMode, cfg.CacheMaxEntries)
	client := NewHTTPClient(cfg, rateLimiter, cacheManager)

	short := client.calculateBackoff(10*time.Millisecond, 0)
	if short <= 0 {
		t.Fatalf("expected positive backoff, got %v", short)
	}

	_ = client.calculateBackoff(time.Hour, 10)
}

func TestHTTPClientContextCancelled(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		writeJSON(t, w, apiResponse{Success: true, Data: map[string]interface{}{"ok": true}})
	}
	client, server := newClientWithServer(t, handler, WithoutCache())
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.httpClient.Post(ctx, "/slow", nil); err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestHTTPClientClientErrorNoRetry(t *testing.T) {
	var attempts int
	handler := func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, apiResponse{
			Success: false,
			Error: &apiErrorResponse{
				Message: "bad input",
			},
		})
	}
	client, server := newClientWithServer(t, handler, WithoutCache())
	defer server.Close()

	if _, err := client.httpClient.Post(context.Background(), "/bad", map[string]string{"pin": "bad"}); err == nil {
		t.Fatal("expected API error for bad request")
	}
	if attempts != 1 {
		t.Fatalf("expected no retries on client error, got %d attempts", attempts)
	}
}

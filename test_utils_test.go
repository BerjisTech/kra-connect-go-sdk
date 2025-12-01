package kra

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type apiResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   *apiErrorResponse      `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
}

type apiErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

const testAPIKey = "abcdefghijklmnop"

// newClientWithServer returns a client wired to the provided HTTP handler.
func newClientWithServer(t testing.TB, handler http.HandlerFunc, extraOpts ...Option) (*Client, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)

	opts := []Option{
		WithAPIKey(strings.Repeat("A", 16)),
		WithBaseURL(server.URL),
		WithoutRateLimit(),
	}
	opts = append(opts, extraOpts...)

	client, err := NewClient(opts...)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	client.httpClient.client = server.Client()
	client.httpClient.client.Timeout = 5 * time.Second

	return client, server
}

// writeJSON writes a JSON response and handles errors for tests.
func writeJSON(t testing.TB, w http.ResponseWriter, v interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	switch val := v.(type) {
	case apiResponse:
		payload := map[string]interface{}{
			"responseCode": "70000",
			"responseDesc": "Successful",
			"status":       "OK",
		}
		if val.Data != nil {
			payload["responseData"] = val.Data
		}
		if !val.Success {
			payload["responseCode"] = "70001"
			payload["status"] = "ERROR"
			if val.Error != nil {
				payload["ErrorCode"] = val.Error.Code
				payload["ErrorMessage"] = val.Error.Message
			} else if val.Message != "" {
				payload["ErrorMessage"] = val.Message
			}
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("failed to encode JSON: %v", err)
		}
	default:
		if err := json.NewEncoder(w).Encode(v); err != nil {
			t.Fatalf("failed to encode JSON: %v", err)
		}
	}
}

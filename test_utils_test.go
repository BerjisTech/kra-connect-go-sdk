package kra

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("failed to encode JSON: %v", err)
	}
}

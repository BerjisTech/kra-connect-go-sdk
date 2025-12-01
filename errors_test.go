package kra

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSDKErrorWrapping(t *testing.T) {
	inner := errors.New("boom")
	err := &SDKError{Message: "outer", Err: inner}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected wrapped error message, got %s", err.Error())
	}
	if !errors.Is(err, inner) {
		t.Fatal("expected errors.Is to unwrap inner error")
	}
}

func TestAPIErrorHelpers(t *testing.T) {
	apiErr := NewAPIError(503, "server down", "/checker/v1/pinbypin", "body")
	if !apiErr.IsServerError() {
		t.Fatal("expected IsServerError to be true")
	}
	if apiErr.IsClientError() {
		t.Fatal("expected IsClientError to be false for 5xx")
	}

	apiErr.StatusCode = 404
	if !apiErr.IsClientError() {
		t.Fatal("expected IsClientError to be true for 4xx")
	}
}

func TestTimeoutAndNetworkErrors(t *testing.T) {
	timeoutErr := NewTimeoutError("/verify", 2*time.Second, 3)
	if timeoutErr.StatusCode != 408 {
		t.Fatalf("unexpected status code %d", timeoutErr.StatusCode)
	}

	netErr := NewNetworkError("/verify", errors.New("dial tcp"))
	if !strings.Contains(netErr.Error(), "Network error") {
		t.Fatalf("unexpected network error message: %s", netErr.Error())
	}

	cacheErr := NewCacheError("set", "pin", "disk full")
	if !strings.Contains(cacheErr.Message, "disk full") {
		t.Fatalf("unexpected cache error message: %s", cacheErr.Message)
	}
}

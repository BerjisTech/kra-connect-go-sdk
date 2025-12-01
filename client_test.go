package kra

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientVerifyPINUsesCache(t *testing.T) {
	var hits int32

	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/verify-pin":
			atomic.AddInt32(&hits, 1)
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"is_valid":          true,
					"taxpayer_name":     "Acme Ltd",
					"status":            "active",
					"taxpayer_type":     "company",
					"registration_date": "2020-01-01",
					"additional_data": map[string]interface{}{
						"source": "cache-test",
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}

	client, server := newClientWithServer(t, handler)
	defer server.Close()

	ctx := context.Background()
	if _, err := client.VerifyPIN(ctx, "P051234567A"); err != nil {
		t.Fatalf("VerifyPIN error = %v", err)
	}
	if _, err := client.VerifyPIN(ctx, "P051234567A"); err != nil {
		t.Fatalf("VerifyPIN second call error = %v", err)
	}

	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("expected 1 network call, got %d", got)
	}
}

func TestClientAllEndpoints(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/verify-pin":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"is_valid":          true,
					"taxpayer_name":     "Acme",
					"status":            "active",
					"taxpayer_type":     "company",
					"registration_date": "2020-02-02",
					"additional_data": map[string]interface{}{
						"note": "full coverage",
					},
				},
			})
		case "/verify-tcc":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"is_valid":         true,
					"expiry_date":      "2025-12-31",
					"issue_date":       "2025-01-01",
					"taxpayer_name":    "Acme",
					"pin_number":       "P051234567A",
					"is_expired":       false,
					"status":           "active",
					"certificate_type": "tax",
					"additional_data": map[string]interface{}{
						"issued_by": "KRA",
					},
				},
			})
		case "/validate-eslip":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"is_valid":          true,
					"taxpayer_pin":      "P051234567A",
					"taxpayer_name":     "Acme",
					"amount":            1000.0,
					"currency":          "KES",
					"payment_date":      "2025-01-01",
					"payment_reference": "ES123",
					"status":            "paid",
					"obligation_type":   "VAT",
					"obligation_period": "202401",
					"additional_data": map[string]interface{}{
						"channel": "mpesa",
					},
				},
			})
		case "/file-nil-return":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"success":                true,
					"pin_number":             "P051234567A",
					"obligation_id":          "OBL123456",
					"period":                 "202401",
					"reference_number":       "REF123",
					"acknowledgement_number": "ACK123",
					"filing_date":            "2025-02-01",
					"status":                 "accepted",
					"message":                "Filed",
					"additional_data": map[string]interface{}{
						"processor": "automation",
					},
				},
			})
		case "/taxpayer/P051234567A":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"pin_number":        "P051234567A",
					"taxpayer_name":     "Acme",
					"taxpayer_type":     "Company",
					"status":            "active",
					"business_name":     "Acme Group",
					"trading_name":      "Acme Trading",
					"postal_address":    "P.O. Box 123",
					"physical_address":  "Nairobi",
					"email_address":     "info@example.com",
					"phone_number":      "+254700000000",
					"registration_date": "2019-01-01",
					"additional_data": map[string]interface{}{
						"segment": "enterprise",
					},
					"obligations": []map[string]interface{}{
						{
							"obligation_id":     "OBL123",
							"obligation_type":   "VAT",
							"description":       "Value Added Tax",
							"status":            "active",
							"registration_date": "2019-01-01",
							"effective_date":    "2019-02-01",
							"end_date":          "2099-12-31",
							"frequency":         "monthly",
							"next_filing_date":  time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02"),
							"is_active":         true,
							"additional_data": map[string]interface{}{
								"notes": "test obligation",
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected endpoint: %s", r.URL.Path)
		}
	}

	client, server := newClientWithServer(t, handler)
	defer server.Close()

	ctx := context.Background()

	if res, err := client.VerifyPIN(ctx, "P051234567A"); err != nil || !res.IsValid {
		t.Fatalf("VerifyPIN() = %v, %v", res, err)
	}

	if res, err := client.VerifyTCC(ctx, "TCC123456"); err != nil || !res.IsCurrentlyValid() {
		t.Fatalf("VerifyTCC() = %v, %v", res, err)
	}

	if res, err := client.ValidateEslip(ctx, "1234567890"); err != nil || !res.IsValid {
		t.Fatalf("ValidateEslip() = %v, %v", res, err)
	}

	nilReq := &NILReturnRequest{
		PINNumber:    "P051234567A",
		ObligationID: "OBL123456",
		Period:       "202401",
	}
	if res, err := client.FileNILReturn(ctx, nilReq); err != nil || !res.IsAccepted() {
		t.Fatalf("FileNILReturn() = %v, %v", res, err)
	}

	if res, err := client.GetTaxpayerDetails(ctx, "P051234567A"); err != nil || res.PINNumber != "P051234567A" {
		t.Fatalf("GetTaxpayerDetails() = %v, %v", res, err)
	}
}

func TestClientBatchOperations(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/verify-pin":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"is_valid":      true,
					"taxpayer_name": "Batch",
					"status":        "active",
					"taxpayer_type": "company",
					"additional_data": map[string]interface{}{
						"batch": true,
					},
				},
			})
		case "/verify-tcc":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"is_valid":         true,
					"is_expired":       false,
					"status":           "active",
					"taxpayer_name":    "Batch",
					"certificate_type": "tax",
				},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}

	client, server := newClientWithServer(t, handler)
	defer server.Close()

	ctx := context.Background()
	pins := []string{"P051234567A", "P051234567B", "P051234567C"}
	results, err := client.VerifyPINsBatch(ctx, pins)
	if err != nil {
		t.Fatalf("VerifyPINsBatch error = %v", err)
	}
	for i, res := range results {
		if res == nil || res.PINNumber != pins[i] {
			t.Fatalf("unexpected result at %d: %+v", i, res)
		}
	}

	tccs := []string{"TCC123456", "TCC123457"}
	if _, err := client.VerifyTCCsBatch(ctx, tccs); err != nil {
		t.Fatalf("VerifyTCCsBatch error = %v", err)
	}
}

func TestClientClearCacheAndClose(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, apiResponse{
			Success: true,
			Data: map[string]interface{}{
				"is_valid":      true,
				"taxpayer_name": "Acme",
				"status":        "active",
				"taxpayer_type": "Company",
			},
		})
	}

	client, server := newClientWithServer(t, handler)
	defer server.Close()

	ctx := context.Background()
	if _, err := client.VerifyPIN(ctx, "P051234567A"); err != nil {
		t.Fatalf("VerifyPIN() error = %v", err)
	}

	if err := client.ClearCache(); err != nil {
		t.Fatalf("ClearCache() error = %v", err)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if _, err := client.VerifyPIN(ctx, "P051234567A"); err == nil {
		t.Fatalf("expected error after Close")
	}

	if err := client.ClearCache(); err == nil {
		t.Fatalf("expected ClearCache to fail after Close")
	}

	if err := client.Close(); err == nil {
		t.Fatalf("expected second Close to fail")
	}
}

func TestClientDebugMode(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, apiResponse{
			Success: true,
			Data: map[string]interface{}{
				"is_valid":      true,
				"taxpayer_name": "Debug",
				"status":        "active",
				"taxpayer_type": "company",
			},
		})
	}
	client, server := newClientWithServer(t, handler, WithDebug(true))
	defer server.Close()

	if _, err := client.VerifyPIN(context.Background(), "P051234567A"); err != nil {
		t.Fatalf("VerifyPIN() error = %v", err)
	}
}

func TestClientValidationErrors(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, apiResponse{Success: true, Data: map[string]interface{}{"is_valid": true}})
	}

	client, server := newClientWithServer(t, handler)
	defer server.Close()

	ctx := context.Background()
	if _, err := client.VerifyPIN(ctx, "INVALID"); err == nil {
		t.Fatalf("expected invalid PIN error")
	}

	if _, err := client.ValidateEslip(ctx, "ABC"); err == nil {
		t.Fatalf("expected invalid eslip error")
	}
}

func TestClientVerifyTCCAPIError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, apiResponse{
			Success: false,
			Error: &apiErrorResponse{
				Code:    "TCC_INVALID",
				Message: "Invalid TCC",
			},
		})
	}

	client, server := newClientWithServer(t, handler)
	defer server.Close()

	ctx := context.Background()
	_, err := client.VerifyTCC(ctx, "TCC123456")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
}

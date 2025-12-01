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
		case "/checker/v1/pinbypin":
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
		case "/checker/v1/pinbypin":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"kraPin":           "P051234567A",
					"isValid":          true,
					"taxpayerName":     "Acme",
					"pinStatus":        "ACTIVE",
					"taxpayerType":     "Company",
					"registrationDate": "2020-02-02",
				},
			})
		case "/v1/kra-tcc/validate":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"isValid":         true,
					"isExpired":       false,
					"expiryDate":      "2025-12-31",
					"issueDate":       "2025-01-01",
					"taxpayerName":    "Acme",
					"kraPin":          "P051234567A",
					"status":          "ACTIVE",
					"certificateType": "tax",
				},
			})
		case "/payment/checker/v1/eslip":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"isValid":          true,
					"taxpayerPin":      "P051234567A",
					"taxpayerName":     "Acme",
					"amount":           1000.0,
					"currency":         "KES",
					"paymentDate":      "2025-01-01",
					"paymentReference": "ES123",
					"status":           "paid",
					"obligationType":   "VAT",
					"obligationPeriod": "202401",
				},
			})
		case "/dtd/return/v1/nil":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"success":               true,
					"referenceNumber":       "REF123",
					"filingDate":            "2025-02-01",
					"acknowledgementNumber": "ACK123",
					"status":                "accepted",
					"message":               "Filed",
				},
			})
		case "/dtd/checker/v1/obligation":
			writeJSON(t, w, apiResponse{
				Success: true,
				Data: map[string]interface{}{
					"obligations": []map[string]interface{}{
						{
							"obligationId":     "OBL123",
							"obligationType":   "VAT",
							"description":      "Value Added Tax",
							"status":           "ACTIVE",
							"registrationDate": "2019-01-01",
							"effectiveDate":    "2019-02-01",
							"endDate":          "2099-12-31",
							"frequency":        "monthly",
							"nextFilingDate":   time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02"),
							"isActive":         true,
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

	if res, err := client.VerifyTCC(ctx, &TCCVerificationRequest{
		KraPIN:    "P051234567A",
		TCCNumber: "TCC123456",
	}); err != nil || !res.IsCurrentlyValid() {
		t.Fatalf("VerifyTCC() = %v, %v", res, err)
	}

	if res, err := client.ValidateEslip(ctx, "1234567890"); err != nil || !res.IsValid {
		t.Fatalf("ValidateEslip() = %v, %v", res, err)
	}

	nilReq := &NILReturnRequest{
		PINNumber:      "P051234567A",
		ObligationCode: 1,
		Month:          1,
		Year:           2024,
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
		case "/checker/v1/pinbypin":
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
		case "/v1/kra-tcc/validate":
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

	tccs := []*TCCVerificationRequest{
		{KraPIN: "P051234567A", TCCNumber: "TCC123456"},
		{KraPIN: "P051234567B", TCCNumber: "TCC123457"},
	}
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
	_, err := client.VerifyTCC(ctx, &TCCVerificationRequest{
		KraPIN:    "P051234567A",
		TCCNumber: "TCC123456",
	})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
}

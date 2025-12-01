package kra

import (
	"context"
	"net/http"
	"testing"
)

func BenchmarkVerifyPIN(b *testing.B) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(b, w, apiResponse{
			Success: true,
			Data: map[string]interface{}{
				"is_valid":      true,
				"taxpayer_name": "Benchmark Co",
				"status":        "active",
				"taxpayer_type": "Company",
			},
		})
	}

	client, server := newClientWithServer(b, handler, WithoutCache())
	defer server.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.VerifyPIN(ctx, "P051234567A"); err != nil {
			b.Fatalf("VerifyPIN error = %v", err)
		}
	}
}

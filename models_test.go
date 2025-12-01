package kra

import (
	"testing"
	"time"
)

func TestPINVerificationResult_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		result *PINVerificationResult
		want   bool
	}{
		{
			name: "valid and active",
			result: &PINVerificationResult{
				IsValid: true,
				Status:  "active",
			},
			want: true,
		},
		{
			name: "valid but not active",
			result: &PINVerificationResult{
				IsValid: true,
				Status:  "inactive",
			},
			want: false,
		},
		{
			name: "not valid",
			result: &PINVerificationResult{
				IsValid: false,
				Status:  "active",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPINVerificationResult_IsCompany(t *testing.T) {
	result := &PINVerificationResult{TaxpayerType: "company"}
	if !result.IsCompany() {
		t.Error("Expected IsCompany() to return true for company type")
	}

	result.TaxpayerType = "individual"
	if result.IsCompany() {
		t.Error("Expected IsCompany() to return false for individual type")
	}
}

func TestPINVerificationResult_IsIndividual(t *testing.T) {
	result := &PINVerificationResult{TaxpayerType: "individual"}
	if !result.IsIndividual() {
		t.Error("Expected IsIndividual() to return true for individual type")
	}

	result.TaxpayerType = "company"
	if result.IsIndividual() {
		t.Error("Expected IsIndividual() to return false for company type")
	}
}

func TestTCCVerificationResult_IsCurrentlyValid(t *testing.T) {
	tests := []struct {
		name   string
		result *TCCVerificationResult
		want   bool
	}{
		{
			name: "valid, not expired, active",
			result: &TCCVerificationResult{
				IsValid:   true,
				IsExpired: false,
				Status:    "active",
			},
			want: true,
		},
		{
			name: "valid but expired",
			result: &TCCVerificationResult{
				IsValid:   true,
				IsExpired: true,
				Status:    "active",
			},
			want: false,
		},
		{
			name: "valid but inactive",
			result: &TCCVerificationResult{
				IsValid:   true,
				IsExpired: false,
				Status:    "inactive",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsCurrentlyValid(); got != tt.want {
				t.Errorf("IsCurrentlyValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTCCVerificationResult_DaysUntilExpiry(t *testing.T) {
	// Test with future date
	futureDate := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")
	result := &TCCVerificationResult{ExpiryDate: futureDate}
	days := result.DaysUntilExpiry()
	if days < 28 || days > 31 {
		t.Errorf("Expected approximately 30 days, got %d", days)
	}

	// Test with past date
	pastDate := time.Now().Add(-30 * 24 * time.Hour).Format("2006-01-02")
	result.ExpiryDate = pastDate
	days = result.DaysUntilExpiry()
	if days >= 0 {
		t.Errorf("Expected negative days for past date, got %d", days)
	}

	// Test with invalid date
	result.ExpiryDate = "invalid-date"
	days = result.DaysUntilExpiry()
	if days != 0 {
		t.Errorf("Expected 0 for invalid date, got %d", days)
	}

	// Test with empty date
	result.ExpiryDate = ""
	days = result.DaysUntilExpiry()
	if days != 0 {
		t.Errorf("Expected 0 for empty date, got %d", days)
	}
}

func TestTCCVerificationResult_IsExpiringSoon(t *testing.T) {
	// Test with date 10 days in future
	futureDate := time.Now().Add(10 * 24 * time.Hour).Format("2006-01-02")
	result := &TCCVerificationResult{ExpiryDate: futureDate}

	// Should be expiring soon within 15 days
	if !result.IsExpiringSoon(15) {
		t.Error("Expected IsExpiringSoon(15) to return true for 10 days out")
	}

	// Should not be expiring soon within 5 days
	if result.IsExpiringSoon(5) {
		t.Error("Expected IsExpiringSoon(5) to return false for 10 days out")
	}
}

func TestTaxpayerDetailsHelpers(t *testing.T) {
	details := &TaxpayerDetails{
		TaxpayerType: "company",
		BusinessName: "BizCo",
		TradingName:  "TradeCo",
		TaxpayerName: "Fallback",
		Status:       "active",
		Obligations: []TaxObligation{
			{ObligationType: "VAT"},
		},
	}

	if !details.IsCompany() {
		t.Fatal("expected IsCompany to return true")
	}
	if details.IsIndividual() {
		t.Fatal("expected IsIndividual to return false")
	}
	if details.GetDisplayName() != "BizCo" {
		t.Fatalf("unexpected display name: %s", details.GetDisplayName())
	}
	if !details.HasObligation("VAT") {
		t.Fatal("expected HasObligation to find VAT")
	}

	details.TaxpayerType = "individual"
	if !details.IsIndividual() {
		t.Fatal("expected IsIndividual after type change")
	}
}

func TestEslipValidationResult_IsPaid(t *testing.T) {
	result := &EslipValidationResult{IsValid: true, Status: "paid"}
	if !result.IsPaid() {
		t.Error("Expected IsPaid() to return true")
	}

	result.Status = "pending"
	if result.IsPaid() {
		t.Error("Expected IsPaid() to return false for pending status")
	}
}

func TestEslipValidationResult_IsPending(t *testing.T) {
	result := &EslipValidationResult{IsValid: true, Status: "pending"}
	if !result.IsPending() {
		t.Error("Expected IsPending() to return true")
	}

	result.Status = "paid"
	if result.IsPending() {
		t.Error("Expected IsPending() to return false for paid status")
	}
}

func TestEslipValidationResult_IsCancelled(t *testing.T) {
	result := &EslipValidationResult{Status: "cancelled"}
	if !result.IsCancelled() {
		t.Error("Expected IsCancelled() to return true")
	}

	result.Status = "paid"
	if result.IsCancelled() {
		t.Error("Expected IsCancelled() to return false for paid status")
	}
}

func TestNILReturnResult_IsAccepted(t *testing.T) {
	result := &NILReturnResult{Success: true, Status: "accepted"}
	if !result.IsAccepted() {
		t.Error("Expected IsAccepted() to return true")
	}

	result.Status = "pending"
	if result.IsAccepted() {
		t.Error("Expected IsAccepted() to return false for pending status")
	}
}

func TestNILReturnResult_IsPending(t *testing.T) {
	result := &NILReturnResult{Success: true, Status: "pending"}
	if !result.IsPending() {
		t.Error("Expected IsPending() to return true")
	}

	result.Status = "accepted"
	if result.IsPending() {
		t.Error("Expected IsPending() to return false for accepted status")
	}
}

func TestNILReturnResult_IsRejected(t *testing.T) {
	// Rejected status
	result := &NILReturnResult{Success: true, Status: "rejected"}
	if !result.IsRejected() {
		t.Error("Expected IsRejected() to return true for rejected status")
	}

	// Failed submission
	result = &NILReturnResult{Success: false}
	if !result.IsRejected() {
		t.Error("Expected IsRejected() to return true for failed submission")
	}

	// Accepted
	result = &NILReturnResult{Success: true, Status: "accepted"}
	if result.IsRejected() {
		t.Error("Expected IsRejected() to return false for accepted status")
	}
}

func TestTaxpayerDetails_IsActive(t *testing.T) {
	details := &TaxpayerDetails{Status: "active"}
	if !details.IsActive() {
		t.Error("Expected IsActive() to return true")
	}

	details.Status = "inactive"
	if details.IsActive() {
		t.Error("Expected IsActive() to return false")
	}
}

func TestTaxpayerDetails_GetDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		details *TaxpayerDetails
		want    string
	}{
		{
			name: "business name available",
			details: &TaxpayerDetails{
				TaxpayerName: "John Doe",
				BusinessName: "Acme Corp",
				TradingName:  "Acme Trading",
			},
			want: "Acme Corp",
		},
		{
			name: "only trading name",
			details: &TaxpayerDetails{
				TaxpayerName: "John Doe",
				TradingName:  "Acme Trading",
			},
			want: "Acme Trading",
		},
		{
			name: "only taxpayer name",
			details: &TaxpayerDetails{
				TaxpayerName: "John Doe",
			},
			want: "John Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.details.GetDisplayName(); got != tt.want {
				t.Errorf("GetDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaxpayerDetails_HasObligation(t *testing.T) {
	details := &TaxpayerDetails{
		Obligations: []TaxObligation{
			{ObligationType: "VAT"},
			{ObligationType: "PAYE"},
			{ObligationType: "INCOME_TAX"},
		},
	}

	if !details.HasObligation("VAT") {
		t.Error("Expected HasObligation('VAT') to return true")
	}

	if details.HasObligation("CUSTOMS") {
		t.Error("Expected HasObligation('CUSTOMS') to return false")
	}
}

func TestTaxObligation_HasEnded(t *testing.T) {
	// Past end date
	pastDate := time.Now().Add(-30 * 24 * time.Hour).Format("2006-01-02")
	obligation := &TaxObligation{EndDate: pastDate}
	if !obligation.HasEnded() {
		t.Error("Expected HasEnded() to return true for past date")
	}

	// Future end date
	futureDate := time.Now().Add(30 * 24 * time.Hour).Format("2006-01-02")
	obligation.EndDate = futureDate
	if obligation.HasEnded() {
		t.Error("Expected HasEnded() to return false for future date")
	}

	// No end date
	obligation.EndDate = ""
	if obligation.HasEnded() {
		t.Error("Expected HasEnded() to return false for no end date")
	}
}

func TestTaxObligation_IsFilingDueSoon(t *testing.T) {
	// Active obligation with filing date 10 days in future
	futureDate := time.Now().Add(10 * 24 * time.Hour).Format("2006-01-02")
	obligation := &TaxObligation{
		IsActive:       true,
		NextFilingDate: futureDate,
	}

	if !obligation.IsFilingDueSoon(15) {
		t.Error("Expected IsFilingDueSoon(15) to return true for 10 days out")
	}

	if obligation.IsFilingDueSoon(5) {
		t.Error("Expected IsFilingDueSoon(5) to return false for 10 days out")
	}

	// Inactive obligation
	obligation.IsActive = false
	if obligation.IsFilingDueSoon(15) {
		t.Error("Expected IsFilingDueSoon() to return false for inactive obligation")
	}
}

func TestTaxObligation_IsFilingOverdue(t *testing.T) {
	// Active obligation with past filing date
	pastDate := time.Now().Add(-10 * 24 * time.Hour).Format("2006-01-02")
	obligation := &TaxObligation{
		IsActive:       true,
		NextFilingDate: pastDate,
	}

	if !obligation.IsFilingOverdue() {
		t.Error("Expected IsFilingOverdue() to return true for past date")
	}

	// Future filing date
	futureDate := time.Now().Add(10 * 24 * time.Hour).Format("2006-01-02")
	obligation.NextFilingDate = futureDate
	if obligation.IsFilingOverdue() {
		t.Error("Expected IsFilingOverdue() to return false for future date")
	}

	// Inactive obligation
	obligation.IsActive = false
	obligation.NextFilingDate = pastDate
	if obligation.IsFilingOverdue() {
		t.Error("Expected IsFilingOverdue() to return false for inactive obligation")
	}
}

package kra

import (
	"time"
)

// PINVerificationResult represents the result of a PIN verification request
type PINVerificationResult struct {
	PINNumber        string                 `json:"pin_number"`
	IsValid          bool                   `json:"is_valid"`
	TaxpayerName     string                 `json:"taxpayer_name,omitempty"`
	Status           string                 `json:"status,omitempty"`
	TaxpayerType     string                 `json:"taxpayer_type,omitempty"`
	RegistrationDate string                 `json:"registration_date,omitempty"`
	AdditionalData   map[string]interface{} `json:"additional_data,omitempty"`
	VerifiedAt       time.Time              `json:"verified_at"`
	Metadata         ResponseMetadata       `json:"metadata"`
	RawData          map[string]interface{} `json:"raw_data,omitempty"`
}

// IsActive returns true if the PIN is valid and active
func (r *PINVerificationResult) IsActive() bool {
	return r.IsValid && r.Status == "active"
}

// IsCompany returns true if the taxpayer is a company
func (r *PINVerificationResult) IsCompany() bool {
	return r.TaxpayerType == "company"
}

// IsIndividual returns true if the taxpayer is an individual
func (r *PINVerificationResult) IsIndividual() bool {
	return r.TaxpayerType == "individual"
}

// TCCVerificationResult represents the result of a TCC verification request
type TCCVerificationResult struct {
	TCCNumber       string                 `json:"tcc_number"`
	IsValid         bool                   `json:"is_valid"`
	TaxpayerName    string                 `json:"taxpayer_name,omitempty"`
	PINNumber       string                 `json:"pin_number,omitempty"`
	IssueDate       string                 `json:"issue_date,omitempty"`
	ExpiryDate      string                 `json:"expiry_date,omitempty"`
	IsExpired       bool                   `json:"is_expired"`
	Status          string                 `json:"status,omitempty"`
	CertificateType string                 `json:"certificate_type,omitempty"`
	AdditionalData  map[string]interface{} `json:"additional_data,omitempty"`
	VerifiedAt      time.Time              `json:"verified_at"`
	Metadata        ResponseMetadata       `json:"metadata"`
	RawData         map[string]interface{} `json:"raw_data,omitempty"`
}

// TCCVerificationRequest represents the payload required for TCC validation
type TCCVerificationRequest struct {
	KraPIN    string `json:"kra_pin"`
	TCCNumber string `json:"tcc_number"`
}

// IsCurrentlyValid returns true if the TCC is valid and not expired
func (r *TCCVerificationResult) IsCurrentlyValid() bool {
	return r.IsValid && !r.IsExpired && r.Status == "active"
}

// DaysUntilExpiry returns the number of days until expiry
func (r *TCCVerificationResult) DaysUntilExpiry() int {
	if r.ExpiryDate == "" {
		return 0
	}

	expiryTime, err := time.Parse("2006-01-02", r.ExpiryDate)
	if err != nil {
		return 0
	}

	days := int(time.Until(expiryTime).Hours() / 24)
	return days
}

// IsExpiringSoon returns true if the TCC expires within the specified days
func (r *TCCVerificationResult) IsExpiringSoon(days int) bool {
	daysUntilExpiry := r.DaysUntilExpiry()
	return daysUntilExpiry >= 0 && daysUntilExpiry <= days
}

// EslipValidationResult represents the result of an e-slip validation request
type EslipValidationResult struct {
	EslipNumber      string                 `json:"eslip_number"`
	IsValid          bool                   `json:"is_valid"`
	TaxpayerPIN      string                 `json:"taxpayer_pin,omitempty"`
	TaxpayerName     string                 `json:"taxpayer_name,omitempty"`
	Amount           float64                `json:"amount,omitempty"`
	Currency         string                 `json:"currency,omitempty"`
	PaymentDate      string                 `json:"payment_date,omitempty"`
	PaymentReference string                 `json:"payment_reference,omitempty"`
	ObligationType   string                 `json:"obligation_type,omitempty"`
	ObligationPeriod string                 `json:"obligation_period,omitempty"`
	Status           string                 `json:"status,omitempty"`
	AdditionalData   map[string]interface{} `json:"additional_data,omitempty"`
	ValidatedAt      time.Time              `json:"validated_at"`
	Metadata         ResponseMetadata       `json:"metadata"`
	RawData          map[string]interface{} `json:"raw_data,omitempty"`
}

// IsPaid returns true if the payment has been confirmed
func (r *EslipValidationResult) IsPaid() bool {
	return r.IsValid && r.Status == "paid"
}

// IsPending returns true if the payment is pending
func (r *EslipValidationResult) IsPending() bool {
	return r.IsValid && r.Status == "pending"
}

// IsCancelled returns true if the payment was cancelled
func (r *EslipValidationResult) IsCancelled() bool {
	return r.Status == "cancelled"
}

// NILReturnRequest represents a NIL return filing request
type NILReturnRequest struct {
	PINNumber      string `json:"pin_number"`
	ObligationCode int    `json:"obligation_code"`
	Month          int    `json:"month"`
	Year           int    `json:"year"`
}

// NILReturnResult represents the result of a NIL return filing
type NILReturnResult struct {
	Success               bool                   `json:"success"`
	PINNumber             string                 `json:"pin_number,omitempty"`
	ObligationID          string                 `json:"obligation_id,omitempty"`
	Period                string                 `json:"period,omitempty"`
	ReferenceNumber       string                 `json:"reference_number,omitempty"`
	FilingDate            string                 `json:"filing_date,omitempty"`
	AcknowledgementNumber string                 `json:"acknowledgement_number,omitempty"`
	Status                string                 `json:"status,omitempty"`
	Message               string                 `json:"message,omitempty"`
	AdditionalData        map[string]interface{} `json:"additional_data,omitempty"`
	FiledAt               time.Time              `json:"filed_at"`
	Metadata              ResponseMetadata       `json:"metadata"`
	RawData               map[string]interface{} `json:"raw_data,omitempty"`
}

// IsAccepted returns true if the filing was accepted
func (r *NILReturnResult) IsAccepted() bool {
	return r.Success && r.Status == "accepted"
}

// IsPending returns true if the filing is pending approval
func (r *NILReturnResult) IsPending() bool {
	return r.Success && r.Status == "pending"
}

// IsRejected returns true if the filing was rejected
func (r *NILReturnResult) IsRejected() bool {
	return !r.Success || r.Status == "rejected"
}

// TaxpayerDetails represents detailed taxpayer information
type TaxpayerDetails struct {
	PINNumber        string                 `json:"pin_number"`
	TaxpayerName     string                 `json:"taxpayer_name,omitempty"`
	TaxpayerType     string                 `json:"taxpayer_type,omitempty"`
	Status           string                 `json:"status,omitempty"`
	RegistrationDate string                 `json:"registration_date,omitempty"`
	BusinessName     string                 `json:"business_name,omitempty"`
	TradingName      string                 `json:"trading_name,omitempty"`
	PostalAddress    string                 `json:"postal_address,omitempty"`
	PhysicalAddress  string                 `json:"physical_address,omitempty"`
	EmailAddress     string                 `json:"email_address,omitempty"`
	PhoneNumber      string                 `json:"phone_number,omitempty"`
	Obligations      []TaxObligation        `json:"obligations,omitempty"`
	AdditionalData   map[string]interface{} `json:"additional_data,omitempty"`
	RetrievedAt      time.Time              `json:"retrieved_at"`
	Metadata         ResponseMetadata       `json:"metadata"`
	RawData          map[string]interface{} `json:"raw_data,omitempty"`
}

// IsActive returns true if the taxpayer is active
func (t *TaxpayerDetails) IsActive() bool {
	return t.Status == "active"
}

// IsCompany returns true if the taxpayer is a company
func (t *TaxpayerDetails) IsCompany() bool {
	return t.TaxpayerType == "company"
}

// IsIndividual returns true if the taxpayer is an individual
func (t *TaxpayerDetails) IsIndividual() bool {
	return t.TaxpayerType == "individual"
}

// GetDisplayName returns the best available name for display
func (t *TaxpayerDetails) GetDisplayName() string {
	if t.BusinessName != "" {
		return t.BusinessName
	}
	if t.TradingName != "" {
		return t.TradingName
	}
	return t.TaxpayerName
}

// HasObligation checks if the taxpayer has a specific obligation type
func (t *TaxpayerDetails) HasObligation(obligationType string) bool {
	for _, ob := range t.Obligations {
		if ob.ObligationType == obligationType {
			return true
		}
	}
	return false
}

// TaxObligation represents a tax obligation
type TaxObligation struct {
	ObligationID     string                 `json:"obligation_id"`
	ObligationType   string                 `json:"obligation_type"`
	Description      string                 `json:"description,omitempty"`
	Status           string                 `json:"status,omitempty"`
	RegistrationDate string                 `json:"registration_date,omitempty"`
	EffectiveDate    string                 `json:"effective_date,omitempty"`
	EndDate          string                 `json:"end_date,omitempty"`
	Frequency        string                 `json:"frequency,omitempty"`
	NextFilingDate   string                 `json:"next_filing_date,omitempty"`
	IsActive         bool                   `json:"is_active"`
	AdditionalData   map[string]interface{} `json:"additional_data,omitempty"`
}

// HasEnded returns true if the obligation has ended
func (o *TaxObligation) HasEnded() bool {
	if o.EndDate == "" {
		return false
	}

	endTime, err := time.Parse("2006-01-02", o.EndDate)
	if err != nil {
		return false
	}

	return time.Now().After(endTime)
}

// IsFilingDueSoon returns true if filing is due within the specified days
func (o *TaxObligation) IsFilingDueSoon(days int) bool {
	if o.NextFilingDate == "" || !o.IsActive {
		return false
	}

	filingTime, err := time.Parse("2006-01-02", o.NextFilingDate)
	if err != nil {
		return false
	}

	daysUntil := int(time.Until(filingTime).Hours() / 24)
	return daysUntil >= 0 && daysUntil <= days
}

// IsFilingOverdue returns true if filing is overdue
func (o *TaxObligation) IsFilingOverdue() bool {
	if o.NextFilingDate == "" || !o.IsActive {
		return false
	}

	filingTime, err := time.Parse("2006-01-02", o.NextFilingDate)
	if err != nil {
		return false
	}

	return time.Now().After(filingTime)
}

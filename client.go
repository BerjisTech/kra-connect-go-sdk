package kra

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Client is the main KRA Connect client
//
// The client is goroutine-safe and can be used concurrently from multiple goroutines.
// It manages authentication, rate limiting, caching, and retry logic automatically.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey(os.Getenv("KRA_API_KEY")),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	result, err := client.VerifyPIN(context.Background(), "P051234567A")
//	if err != nil {
//	    log.Fatal(err)
//	}
type Client struct {
	config       *Config
	httpClient   *HTTPClient
	rateLimiter  *RateLimiter
	cacheManager *CacheManager
	closed       bool
	mu           sync.RWMutex
}

// NewClient creates a new KRA Connect client
//
// The client must be configured with at least an API key using WithAPIKey().
// Other configuration options are optional and have sensible defaults.
//
// Example:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey("your-api-key"),
//	    kra.WithTimeout(30 * time.Second),
//	    kra.WithRetry(3, time.Second, 32*time.Second),
//	    kra.WithCache(true, 1*time.Hour),
//	)
func NewClient(opts ...Option) (*Client, error) {
	// Start with default config
	config := DefaultConfig()

	// Apply options
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create components
	rateLimiter := NewRateLimiter(
		config.MaxRequests,
		config.RateLimitWindow,
		config.RateLimitEnabled,
		config.DebugMode,
	)

	cacheManager := NewCacheManager(config.CacheEnabled, config.DebugMode, config.CacheMaxEntries)

	httpClient := NewHTTPClient(config, rateLimiter, cacheManager)

	return &Client{
		config:       config,
		httpClient:   httpClient,
		rateLimiter:  rateLimiter,
		cacheManager: cacheManager,
	}, nil
}

// VerifyPIN verifies a KRA PIN number
//
// The PIN must be in the format: P followed by 9 digits and a letter (e.g., P051234567A).
// Results are cached according to the configured PIN verification TTL.
//
// Example:
//
//	result, err := client.VerifyPIN(ctx, "P051234567A")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.IsValid {
//	    fmt.Printf("Valid PIN: %s\n", result.TaxpayerName)
//	}
func (c *Client) VerifyPIN(ctx context.Context, pin string) (*PINVerificationResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	// Validate and normalize PIN
	normalizedPIN, err := ValidateAndNormalizePIN(pin)
	if err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := GenerateCacheKey("pin_verification", normalizedPIN)
	if cached, found := c.cacheManager.Get(cacheKey); found {
		if result, ok := cached.(*PINVerificationResult); ok {
			return result, nil
		}
	}

	// Make API request
	apiResp, err := c.httpClient.Post(ctx, "/checker/v1/pinbypin", map[string]string{
		"KRAPIN": normalizedPIN,
	})
	if err != nil {
		return nil, err
	}

	data := apiResp.Data
	result := &PINVerificationResult{
		PINNumber:        normalizedPIN,
		VerifiedAt:       time.Now(),
		Metadata:         apiResp.Meta,
		RawData:          data,
		AdditionalData:   data,
		TaxpayerName:     firstString(data, "taxpayerName", "TaxpayerName", "taxpayer_name"),
		Status:           strings.ToLower(firstString(data, "pinStatus", "status", "TaxpayerStatus")),
		TaxpayerType:     strings.ToLower(firstString(data, "taxpayerType", "TaxpayerType", "taxpayer_type")),
		RegistrationDate: firstString(data, "registrationDate", "RegistrationDate", "registration_date"),
	}

	if pinValue := firstString(data, "kraPin", "KRAPIN", "pin"); pinValue != "" {
		result.PINNumber = pinValue
	}

	if isValid, ok := firstBool(data, "isValid", "IsValid"); ok {
		result.IsValid = isValid
	} else {
		result.IsValid = inferValidityFromStatus(result.Status)
	}

	// Cache result
	c.cacheManager.Set(cacheKey, result, c.config.PINVerificationTTL)

	return result, nil
}

// VerifyTCC verifies a Tax Compliance Certificate
//
// The TCC must be in the format: TCC followed by digits (e.g., TCC123456).
// Results are cached according to the configured TCC verification TTL.
//
// Example:
//
//	req := &kra.TCCVerificationRequest{
//	    KraPIN:    "P051234567A",
//	    TCCNumber: "TCC123456",
//	}
//	result, err := client.VerifyTCC(ctx, req)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.IsCurrentlyValid() {
//	    fmt.Printf("TCC valid until: %s\n", result.ExpiryDate)
//	}
func (c *Client) VerifyTCC(ctx context.Context, req *TCCVerificationRequest) (*TCCVerificationResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, fmt.Errorf("verification request cannot be nil")
	}

	normalizedPIN, err := ValidateAndNormalizePIN(req.KraPIN)
	if err != nil {
		return nil, err
	}

	normalizedTCC, err := ValidateAndNormalizeTCC(req.TCCNumber)
	if err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := GenerateCacheKey("tcc_verification", normalizedPIN+"_"+normalizedTCC)
	if cached, found := c.cacheManager.Get(cacheKey); found {
		if result, ok := cached.(*TCCVerificationResult); ok {
			return result, nil
		}
	}

	// Make API request
	apiResp, err := c.httpClient.Post(ctx, "/v1/kra-tcc/validate", map[string]string{
		"kraPIN":    normalizedPIN,
		"tccNumber": normalizedTCC,
	})
	if err != nil {
		return nil, err
	}

	// Parse response
	result := &TCCVerificationResult{
		TCCNumber:      normalizedTCC,
		PINNumber:      normalizedPIN,
		VerifiedAt:     time.Now(),
		Metadata:       apiResp.Meta,
		RawData:        apiResp.Data,
		AdditionalData: apiResp.Data,
		TaxpayerName:   firstString(apiResp.Data, "taxpayerName", "TaxpayerName", "taxpayer_name"),
		IssueDate:      firstString(apiResp.Data, "issueDate", "IssueDate"),
		ExpiryDate:     firstString(apiResp.Data, "expiryDate", "ExpiryDate"),
		Status:         strings.ToLower(firstString(apiResp.Data, "status", "tccStatus")),
		CertificateType: firstString(apiResp.Data,
			"certificateType",
			"CertificateType"),
	}

	if pin := firstString(apiResp.Data, "kraPin", "TaxpayerPIN", "pin_number"); pin != "" {
		result.PINNumber = pin
	}

	if valid, ok := firstBool(apiResp.Data, "isValid", "IsValid"); ok {
		result.IsValid = valid
	} else {
		result.IsValid = inferValidityFromStatus(result.Status)
	}

	if expired, ok := firstBool(apiResp.Data, "isExpired", "IsExpired"); ok {
		result.IsExpired = expired
	}

	// Cache result
	c.cacheManager.Set(cacheKey, result, c.config.TCCVerificationTTL)

	return result, nil
}

// ValidateEslip validates an electronic payment slip
//
// Results are cached according to the configured e-slip validation TTL.
//
// Example:
//
//	result, err := client.ValidateEslip(ctx, "1234567890")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.IsPaid() {
//	    fmt.Printf("Payment confirmed: %.2f %s\n", result.Amount, result.Currency)
//	}
func (c *Client) ValidateEslip(ctx context.Context, eslipNumber string) (*EslipValidationResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	// Validate e-slip number
	if err := ValidateEslipNumber(eslipNumber); err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := GenerateCacheKey("eslip_validation", eslipNumber)
	if cached, found := c.cacheManager.Get(cacheKey); found {
		if result, ok := cached.(*EslipValidationResult); ok {
			return result, nil
		}
	}

	// Make API request
	apiResp, err := c.httpClient.Post(ctx, "/payment/checker/v1/eslip", map[string]string{
		"EslipNumber": eslipNumber,
	})
	if err != nil {
		return nil, err
	}

	data := apiResp.Data
	result := &EslipValidationResult{
		EslipNumber:  firstString(data, "EslipNumber", "eslipNumber", "eslip", "eslip_number"),
		TaxpayerPIN:  firstString(data, "taxpayerPin", "TaxpayerPIN", "taxpayer_pin"),
		TaxpayerName: firstString(data, "taxpayerName", "TaxpayerName", "taxpayer_name"),
		PaymentDate:  firstString(data, "paymentDate", "PaymentDate"),
		PaymentReference: firstString(
			data,
			"paymentReference",
			"PaymentReference",
			"referenceNumber",
			"payment_reference",
		),
		ObligationType:   firstString(data, "obligationType", "taxType", "obligation_type"),
		ObligationPeriod: firstString(data, "obligationPeriod", "taxPeriod", "obligation_period"),
		Status:           strings.ToLower(firstString(data, "status", "eslipStatus")),
		ValidatedAt:      time.Now(),
		Metadata:         apiResp.Meta,
		RawData:          data,
		AdditionalData:   data,
	}

	if result.EslipNumber == "" {
		result.EslipNumber = eslipNumber
	}

	if amount, ok := firstFloat64(data, "amount", "Amount"); ok {
		result.Amount = amount
	}

	if isValid, ok := firstBool(data, "isValid", "IsValid"); ok {
		result.IsValid = isValid
	} else {
		result.IsValid = inferValidityFromStatus(result.Status)
	}

	if currency := firstString(data, "currency", "Currency"); currency != "" {
		result.Currency = currency
	}

	// Cache result
	c.cacheManager.Set(cacheKey, result, c.config.EslipValidationTTL)

	return result, nil
}

// FileNILReturn files a NIL return for a tax obligation
//
// Example:
//
//	result, err := client.FileNILReturn(ctx, &kra.NILReturnRequest{
//	    PINNumber:      "P051234567A",
//	    ObligationCode: 1,
//	    Month:          1,
//	    Year:           2024,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.IsAccepted() {
//	    fmt.Printf("Reference: %s\n", result.ReferenceNumber)
//	}
func (c *Client) FileNILReturn(ctx context.Context, req *NILReturnRequest) (*NILReturnResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	normalizedPIN, err := ValidateAndNormalizePIN(req.PINNumber)
	if err != nil {
		return nil, err
	}
	if req.ObligationCode <= 0 {
		return nil, NewValidationError("obligation_code", "Obligation code must be positive")
	}
	if req.Month < 1 || req.Month > 12 {
		return nil, NewValidationError("month", "Month must be between 1 and 12")
	}
	if req.Year < 2000 {
		return nil, NewValidationError("year", "Year must be >= 2000")
	}

	payload := map[string]interface{}{
		"TAXPAYERDETAILS": map[string]interface{}{
			"TaxpayerPIN":    normalizedPIN,
			"ObligationCode": req.ObligationCode,
			"Month":          req.Month,
			"Year":           req.Year,
		},
	}

	apiResp, err := c.httpClient.Post(ctx, "/dtd/return/v1/nil", payload)
	if err != nil {
		return nil, err
	}

	data := apiResp.Data
	result := &NILReturnResult{
		PINNumber:             normalizedPIN,
		ObligationID:          fmt.Sprintf("%d", req.ObligationCode),
		Period:                fmt.Sprintf("%04d%02d", req.Year, req.Month),
		FiledAt:               time.Now(),
		Metadata:              apiResp.Meta,
		RawData:               data,
		AdditionalData:        data,
		ReferenceNumber:       firstString(data, "referenceNumber", "RefNumber"),
		FilingDate:            firstString(data, "filingDate", "FilingDate"),
		AcknowledgementNumber: firstString(data, "acknowledgementNumber", "AcknowledgementNumber"),
		Status:                strings.ToLower(firstString(data, "status", "filingStatus")),
		Message:               firstString(data, "message", "responseDesc"),
	}

	if success, ok := firstBool(data, "success", "Success"); ok {
		result.Success = success
	} else {
		result.Success = inferValidityFromStatus(result.Status)
	}

	return result, nil
}

// GetTaxpayerDetails retrieves detailed taxpayer information
//
// Results are cached according to the configured taxpayer details TTL.
//
// Example:
//
//	details, err := client.GetTaxpayerDetails(ctx, "P051234567A")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Name: %s\n", details.GetDisplayName())
//	fmt.Printf("Obligations: %d\n", len(details.Obligations))
func (c *Client) GetTaxpayerDetails(ctx context.Context, pin string) (*TaxpayerDetails, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	// Validate and normalize PIN
	normalizedPIN, err := ValidateAndNormalizePIN(pin)
	if err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := GenerateCacheKey("taxpayer_details", normalizedPIN)
	if cached, found := c.cacheManager.Get(cacheKey); found {
		if details, ok := cached.(*TaxpayerDetails); ok {
			return details, nil
		}
	}

	profileResp, err := c.httpClient.Post(ctx, "/checker/v1/pinbypin", map[string]string{
		"KRAPIN": normalizedPIN,
	})
	if err != nil {
		return nil, err
	}

	obligationResp, err := c.httpClient.Post(ctx, "/dtd/checker/v1/obligation", map[string]string{
		"taxPayerPin": normalizedPIN,
	})
	if err != nil {
		return nil, err
	}

	profile := profileResp.Data
	obligations := parseObligations(obligationResp.Data)

	extra := map[string]interface{}{
		"profile":     profile,
		"obligations": obligationResp.Data,
	}

	details := &TaxpayerDetails{
		PINNumber:        normalizedPIN,
		TaxpayerName:     firstString(profile, "taxpayerName", "TaxpayerName", "taxpayer_name"),
		TaxpayerType:     strings.ToLower(firstString(profile, "taxpayerType", "TaxpayerType", "taxpayer_type")),
		Status:           strings.ToLower(firstString(profile, "pinStatus", "status", "TaxpayerStatus")),
		RegistrationDate: firstString(profile, "registrationDate", "RegistrationDate", "registration_date"),
		BusinessName:     firstString(profile, "businessName", "BusinessName"),
		TradingName:      firstString(profile, "tradingName", "TradingName"),
		PostalAddress:    firstString(profile, "postalAddress", "PostalAddress"),
		PhysicalAddress:  firstString(profile, "physicalAddress", "PhysicalAddress"),
		EmailAddress:     firstString(profile, "emailAddress", "EmailAddress"),
		PhoneNumber:      firstString(profile, "phoneNumber", "PhoneNumber"),
		Obligations:      obligations,
		AdditionalData:   extra,
		RetrievedAt:      time.Now(),
		Metadata:         profileResp.Meta,
		RawData:          profile,
	}

	if details.TaxpayerName == "" {
		details.TaxpayerName = firstString(profile, "legalName", "BusinessName")
	}

	// Cache result
	c.cacheManager.Set(cacheKey, details, c.config.TaxpayerDetailsTTL)

	return details, nil
}

func parseObligations(payload map[string]interface{}) []TaxObligation {
	if payload == nil {
		return nil
	}

	items, ok := payload["obligations"].([]interface{})
	if !ok {
		return nil
	}

	obligations := make([]TaxObligation, 0, len(items))
	for _, item := range items {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		obligation := TaxObligation{
			ObligationID:     firstString(row, "obligationId", "ObligationID", "obligation_id"),
			ObligationType:   firstString(row, "obligationType", "ObligationType", "obligation_type"),
			Description:      firstString(row, "description", "Description"),
			Status:           strings.ToLower(firstString(row, "status", "Status")),
			RegistrationDate: firstString(row, "registrationDate", "RegistrationDate"),
			EffectiveDate:    firstString(row, "effectiveDate", "EffectiveDate"),
			EndDate:          firstString(row, "endDate", "EndDate"),
			Frequency:        firstString(row, "frequency", "Frequency"),
			NextFilingDate:   firstString(row, "nextFilingDate", "NextFilingDate"),
			AdditionalData:   row,
		}
		if isActive, ok := firstBool(row, "isActive", "IsActive"); ok {
			obligation.IsActive = isActive
		} else {
			obligation.IsActive = inferValidityFromStatus(obligation.Status)
		}
		obligations = append(obligations, obligation)
	}

	return obligations
}

func inferValidityFromStatus(status string) bool {
	if status == "" {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return false
	}
	if strings.Contains(s, "invalid") || strings.Contains(s, "inactive") || strings.Contains(s, "expired") || strings.Contains(s, "reject") {
		return false
	}
	return true
}

// VerifyPINsBatch verifies multiple PIN numbers in parallel
//
// This method is more efficient than calling VerifyPIN multiple times
// as it processes requests concurrently with proper goroutine management.
//
// Example:
//
//	pins := []string{"P051234567A", "P051234567B", "P051234567C"}
//	results, err := client.VerifyPINsBatch(ctx, pins)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, result := range results {
//	    fmt.Printf("%s: %v\n", result.PINNumber, result.IsValid)
//	}
func (c *Client) VerifyPINsBatch(ctx context.Context, pins []string) ([]*PINVerificationResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	results := make([]*PINVerificationResult, len(pins))
	errs := make([]error, len(pins))

	var wg sync.WaitGroup
	for i, pin := range pins {
		wg.Add(1)
		go func(index int, p string) {
			defer wg.Done()
			result, err := c.VerifyPIN(ctx, p)
			results[index] = result
			errs[index] = err
		}(i, pin)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errs {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// VerifyTCCsBatch verifies multiple TCC numbers in parallel
//
// Example:
//
//	tccs := []string{"TCC123456", "TCC123457", "TCC123458"}
//	results, err := client.VerifyTCCsBatch(ctx, tccs)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *Client) VerifyTCCsBatch(ctx context.Context, requests []*TCCVerificationRequest) ([]*TCCVerificationResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	results := make([]*TCCVerificationResult, len(requests))
	errs := make([]error, len(requests))

	var wg sync.WaitGroup
	for i, req := range requests {
		wg.Add(1)
		go func(index int, r *TCCVerificationRequest) {
			defer wg.Done()
			result, err := c.VerifyTCC(ctx, r)
			results[index] = result
			errs[index] = err
		}(i, req)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errs {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

// ClearCache clears all cached data
//
// Use this when you want to force fresh data from the API.
func (c *Client) ClearCache() error {
	if err := c.checkClosed(); err != nil {
		return err
	}

	c.cacheManager.Clear()
	return nil
}

// Close closes the client and releases resources
//
// After calling Close, the client cannot be used anymore.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client already closed")
	}

	c.closed = true
	c.cacheManager.Clear()

	return nil
}

// checkClosed checks if the client has been closed
func (c *Client) checkClosed() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	return nil
}

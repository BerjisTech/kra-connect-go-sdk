package kra

import (
	"context"
	"fmt"
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
	responseData, err := c.httpClient.Post(ctx, "/verify-pin", map[string]string{
		"pin": normalizedPIN,
	})
	if err != nil {
		return nil, err
	}

	// Parse response
	result := &PINVerificationResult{
		PINNumber:  normalizedPIN,
		VerifiedAt: time.Now(),
	}

	if isValid, ok := responseData["is_valid"].(bool); ok {
		result.IsValid = isValid
	}
	if taxpayerName, ok := responseData["taxpayer_name"].(string); ok {
		result.TaxpayerName = taxpayerName
	}
	if status, ok := responseData["status"].(string); ok {
		result.Status = status
	}
	if taxpayerType, ok := responseData["taxpayer_type"].(string); ok {
		result.TaxpayerType = taxpayerType
	}
	if registrationDate, ok := responseData["registration_date"].(string); ok {
		result.RegistrationDate = registrationDate
	}
	if additionalData, ok := responseData["additional_data"].(map[string]interface{}); ok {
		result.AdditionalData = additionalData
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
//	result, err := client.VerifyTCC(ctx, "TCC123456")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.IsCurrentlyValid() {
//	    fmt.Printf("TCC valid until: %s\n", result.ExpiryDate)
//	}
func (c *Client) VerifyTCC(ctx context.Context, tcc string) (*TCCVerificationResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	// Validate and normalize TCC
	normalizedTCC, err := ValidateAndNormalizeTCC(tcc)
	if err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := GenerateCacheKey("tcc_verification", normalizedTCC)
	if cached, found := c.cacheManager.Get(cacheKey); found {
		if result, ok := cached.(*TCCVerificationResult); ok {
			return result, nil
		}
	}

	// Make API request
	responseData, err := c.httpClient.Post(ctx, "/verify-tcc", map[string]string{
		"tcc": normalizedTCC,
	})
	if err != nil {
		return nil, err
	}

	// Parse response
	result := &TCCVerificationResult{
		TCCNumber:  normalizedTCC,
		VerifiedAt: time.Now(),
	}

	if isValid, ok := responseData["is_valid"].(bool); ok {
		result.IsValid = isValid
	}
	if taxpayerName, ok := responseData["taxpayer_name"].(string); ok {
		result.TaxpayerName = taxpayerName
	}
	if pinNumber, ok := responseData["pin_number"].(string); ok {
		result.PINNumber = pinNumber
	}
	if issueDate, ok := responseData["issue_date"].(string); ok {
		result.IssueDate = issueDate
	}
	if expiryDate, ok := responseData["expiry_date"].(string); ok {
		result.ExpiryDate = expiryDate
	}
	if isExpired, ok := responseData["is_expired"].(bool); ok {
		result.IsExpired = isExpired
	}
	if status, ok := responseData["status"].(string); ok {
		result.Status = status
	}
	if certificateType, ok := responseData["certificate_type"].(string); ok {
		result.CertificateType = certificateType
	}
	if additionalData, ok := responseData["additional_data"].(map[string]interface{}); ok {
		result.AdditionalData = additionalData
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
	responseData, err := c.httpClient.Post(ctx, "/validate-eslip", map[string]string{
		"eslip_number": eslipNumber,
	})
	if err != nil {
		return nil, err
	}

	// Parse response
	result := &EslipValidationResult{
		EslipNumber: eslipNumber,
		ValidatedAt: time.Now(),
	}

	if isValid, ok := responseData["is_valid"].(bool); ok {
		result.IsValid = isValid
	}
	if taxpayerPIN, ok := responseData["taxpayer_pin"].(string); ok {
		result.TaxpayerPIN = taxpayerPIN
	}
	if taxpayerName, ok := responseData["taxpayer_name"].(string); ok {
		result.TaxpayerName = taxpayerName
	}
	if amount, ok := responseData["amount"].(float64); ok {
		result.Amount = amount
	}
	if currency, ok := responseData["currency"].(string); ok {
		result.Currency = currency
	}
	if paymentDate, ok := responseData["payment_date"].(string); ok {
		result.PaymentDate = paymentDate
	}
	if paymentRef, ok := responseData["payment_reference"].(string); ok {
		result.PaymentReference = paymentRef
	}
	if obligationType, ok := responseData["obligation_type"].(string); ok {
		result.ObligationType = obligationType
	}
	if obligationPeriod, ok := responseData["obligation_period"].(string); ok {
		result.ObligationPeriod = obligationPeriod
	}
	if status, ok := responseData["status"].(string); ok {
		result.Status = status
	}
	if additionalData, ok := responseData["additional_data"].(map[string]interface{}); ok {
		result.AdditionalData = additionalData
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
//	    PINNumber:    "P051234567A",
//	    ObligationID: "OBL123456",
//	    Period:       "202401",
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

	// Validate request
	if _, err := ValidateAndNormalizePIN(req.PINNumber); err != nil {
		return nil, err
	}
	if err := ValidateObligationID(req.ObligationID); err != nil {
		return nil, err
	}
	if err := ValidatePeriod(req.Period); err != nil {
		return nil, err
	}

	// Make API request (no caching for write operations)
	responseData, err := c.httpClient.Post(ctx, "/file-nil-return", map[string]string{
		"pin_number":    req.PINNumber,
		"obligation_id": req.ObligationID,
		"period":        req.Period,
	})
	if err != nil {
		return nil, err
	}

	// Parse response
	result := &NILReturnResult{
		FiledAt: time.Now(),
	}

	if success, ok := responseData["success"].(bool); ok {
		result.Success = success
	}
	if pinNumber, ok := responseData["pin_number"].(string); ok {
		result.PINNumber = pinNumber
	}
	if obligationID, ok := responseData["obligation_id"].(string); ok {
		result.ObligationID = obligationID
	}
	if period, ok := responseData["period"].(string); ok {
		result.Period = period
	}
	if refNumber, ok := responseData["reference_number"].(string); ok {
		result.ReferenceNumber = refNumber
	}
	if filingDate, ok := responseData["filing_date"].(string); ok {
		result.FilingDate = filingDate
	}
	if ackNumber, ok := responseData["acknowledgement_number"].(string); ok {
		result.AcknowledgementNumber = ackNumber
	}
	if status, ok := responseData["status"].(string); ok {
		result.Status = status
	}
	if message, ok := responseData["message"].(string); ok {
		result.Message = message
	}
	if additionalData, ok := responseData["additional_data"].(map[string]interface{}); ok {
		result.AdditionalData = additionalData
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

	// Make API request
	responseData, err := c.httpClient.Get(ctx, fmt.Sprintf("/taxpayer/%s", normalizedPIN))
	if err != nil {
		return nil, err
	}

	// Parse response
	details := &TaxpayerDetails{
		PINNumber:   normalizedPIN,
		RetrievedAt: time.Now(),
	}

	if taxpayerName, ok := responseData["taxpayer_name"].(string); ok {
		details.TaxpayerName = taxpayerName
	}
	if taxpayerType, ok := responseData["taxpayer_type"].(string); ok {
		details.TaxpayerType = taxpayerType
	}
	if status, ok := responseData["status"].(string); ok {
		details.Status = status
	}
	if registrationDate, ok := responseData["registration_date"].(string); ok {
		details.RegistrationDate = registrationDate
	}
	if businessName, ok := responseData["business_name"].(string); ok {
		details.BusinessName = businessName
	}
	if tradingName, ok := responseData["trading_name"].(string); ok {
		details.TradingName = tradingName
	}
	if postalAddress, ok := responseData["postal_address"].(string); ok {
		details.PostalAddress = postalAddress
	}
	if physicalAddress, ok := responseData["physical_address"].(string); ok {
		details.PhysicalAddress = physicalAddress
	}
	if emailAddress, ok := responseData["email_address"].(string); ok {
		details.EmailAddress = emailAddress
	}
	if phoneNumber, ok := responseData["phone_number"].(string); ok {
		details.PhoneNumber = phoneNumber
	}
	if additionalData, ok := responseData["additional_data"].(map[string]interface{}); ok {
		details.AdditionalData = additionalData
	}

	// Parse obligations
	if obligations, ok := responseData["obligations"].([]interface{}); ok {
		for _, ob := range obligations {
			if obMap, ok := ob.(map[string]interface{}); ok {
				obligation := TaxObligation{}
				if id, ok := obMap["obligation_id"].(string); ok {
					obligation.ObligationID = id
				}
				if obType, ok := obMap["obligation_type"].(string); ok {
					obligation.ObligationType = obType
				}
				if desc, ok := obMap["description"].(string); ok {
					obligation.Description = desc
				}
				if status, ok := obMap["status"].(string); ok {
					obligation.Status = status
				}
				if isActive, ok := obMap["is_active"].(bool); ok {
					obligation.IsActive = isActive
				}
				details.Obligations = append(details.Obligations, obligation)
			}
		}
	}

	// Cache result
	c.cacheManager.Set(cacheKey, details, c.config.TaxpayerDetailsTTL)

	return details, nil
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
func (c *Client) VerifyTCCsBatch(ctx context.Context, tccs []string) ([]*TCCVerificationResult, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}

	results := make([]*TCCVerificationResult, len(tccs))
	errs := make([]error, len(tccs))

	var wg sync.WaitGroup
	for i, tcc := range tccs {
		wg.Add(1)
		go func(index int, t string) {
			defer wg.Done()
			result, err := c.VerifyTCC(ctx, t)
			results[index] = result
			errs[index] = err
		}(i, tcc)
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

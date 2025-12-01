# Changelog

All notable changes to the KRA-Connect Go SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.3] - 2025-12-01

### Added
- OAuth client-credentials helpers via `WithClientCredentials` and `WithTokenURL` so CLIs can fetch sandbox tokens without manual environment tweaks.
- TCC verification request payload (`TCCVerificationRequest`) to align with the latest CLI command inputs.
- Documentation updates covering the new configuration knobs and request helpers.

### Changed
- `NILReturnRequest` now includes `ObligationCode`, `Month`, and `Year` fields to match the NIL filing workflow.
- The SDK User-Agent string now stays in sync with the version constant to simplify release management.

## [0.1.1] - 2025-12-01

### Fixed
- Updated module path, documentation, and tooling references to the `BerjisTech` organization.
- Bumped SDK version metadata (constants, User-Agent string) in preparation for the v0.1.1 release.

## [0.1.0] - 2025-01-28

### Added
- Initial release of KRA-Connect Go SDK
- PIN verification with `VerifyPIN()` method
- TCC verification with `VerifyTCC()` method
- E-slip validation with `ValidateEslip()` method
- NIL return filing with `FileNILReturn()` method
- Taxpayer details retrieval with `GetTaxpayerDetails()` method
- Batch operations with `VerifyPINsBatch()` and `VerifyTCCsBatch()` methods
- Context support for cancellation and timeouts
- Automatic retry with exponential backoff
- Response caching with configurable TTL
- Token bucket rate limiting
- Comprehensive error types:
  - `ValidationError` for input validation failures
  - `InvalidPINFormatError` and `InvalidTCCFormatError` for format errors
  - `AuthenticationError` for API authentication failures
  - `RateLimitError` for rate limit exceeded scenarios
  - `TimeoutError` for request timeouts
  - `APIError` for general API errors
  - `NetworkError` for network-related errors
  - `CacheError` for cache operation errors
- Functional options pattern for client configuration
- Goroutine-safe operations
- Debug mode for detailed logging
- Helper methods on result types:
  - `PINVerificationResult.IsActive()`, `IsCompany()`, `IsIndividual()`
  - `TCCVerificationResult.IsCurrentlyValid()`, `DaysUntilExpiry()`, `IsExpiringSoon()`
  - `EslipValidationResult.IsPaid()`, `IsPending()`, `IsCancelled()`
  - `NILReturnResult.IsAccepted()`, `IsPending()`, `IsRejected()`
  - `TaxpayerDetails.IsActive()`, `GetDisplayName()`, `HasObligation()`
  - `TaxObligation.HasEnded()`, `IsFilingDueSoon()`, `IsFilingOverdue()`
- Comprehensive test suite with 80%+ coverage
- Example applications:
  - Basic usage examples
  - Batch operations examples
  - Error handling examples
  - Advanced usage examples
- Complete documentation with GoDoc comments
- Contributing guidelines
- MIT License

### Configuration Options
- `WithAPIKey()` - Set API key for authentication
- `WithBaseURL()` - Configure custom API base URL
- `WithTimeout()` - Set HTTP request timeout
- `WithRetry()` - Configure retry behavior
- `WithRateLimit()` - Enable and configure rate limiting
- `WithoutRateLimit()` - Disable rate limiting
- `WithCache()` - Enable caching with default TTL
- `WithCustomCacheTTLs()` - Set custom TTL for each operation type
- `WithoutCache()` - Disable caching
- `WithDebug()` - Enable debug logging

### Dependencies
- Go 1.21 or higher
- `github.com/stretchr/testify` v1.8.4 (dev)
- `golang.org/x/time` v0.5.0 (for rate limiting)

[Unreleased]: https://github.com/BerjisTech/kra-connect-go-sdk/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/BerjisTech/kra-connect-go-sdk/compare/v0.1.1...v0.1.3
[0.1.1]: https://github.com/BerjisTech/kra-connect-go-sdk/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/BerjisTech/kra-connect-go-sdk/releases/tag/v0.1.0

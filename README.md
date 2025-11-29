# KRA-Connect Go SDK

> Official Go SDK for Kenya Revenue Authority's GavaConnect API

[![Go Version](https://img.shields.io/github/go-mod/go-version/kra-connect/go-sdk?style=flat-square)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/kra-connect/go-sdk.svg)](https://pkg.go.dev/github.com/kra-connect/go-sdk)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](https://opensource.org/licenses/MIT)
[![Tests](https://img.shields.io/github/actions/workflow/status/kra-connect/go-sdk/tests.yml?branch=main&label=tests&style=flat-square)](https://github.com/kra-connect/go-sdk/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kra-connect/go-sdk?style=flat-square)](https://goreportcard.com/report/github.com/kra-connect/go-sdk)
[![Code Coverage](https://img.shields.io/codecov/c/github/kra-connect/go-sdk?style=flat-square)](https://codecov.io/gh/kra-connect/go-sdk)

## Features

- ✅ **PIN Verification** - Verify KRA PIN numbers
- ✅ **TCC Verification** - Check Tax Compliance Certificates
- ✅ **e-Slip Validation** - Validate electronic payment slips
- ✅ **NIL Returns** - File NIL returns programmatically
- ✅ **Taxpayer Details** - Retrieve taxpayer information
- ✅ **Type Safety** - Full Go type safety with generics
- ✅ **Context Support** - Context-aware operations for cancellation and timeouts
- ✅ **Retry Logic** - Automatic retry with exponential backoff
- ✅ **Caching** - Response caching with TTL
- ✅ **Rate Limiting** - Built-in token bucket rate limiter
- ✅ **Idiomatic Go** - Follows Go best practices and conventions
- ✅ **Zero Dependencies** (core) - Only stdlib for core functionality
- ✅ **Goroutine Safe** - Thread-safe operations

## Requirements

- Go 1.21 or higher

## Installation

```bash
go get github.com/kra-connect/go-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    kra "github.com/kra-connect/go-sdk"
)

func main() {
    // Create client
    client, err := kra.NewClient(
        kra.WithAPIKey(os.Getenv("KRA_API_KEY")),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Verify PIN
    ctx := context.Background()
    result, err := client.VerifyPIN(ctx, "P051234567A")
    if err != nil {
        log.Fatal(err)
    }

    if result.IsValid {
        fmt.Printf("Valid PIN: %s\n", result.TaxpayerName)
        fmt.Printf("Status: %s\n", result.Status)
    }
}
```

## Usage Examples

### PIN Verification

```go
result, err := client.VerifyPIN(ctx, "P051234567A")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Valid: %v\n", result.IsValid)
fmt.Printf("Name: %s\n", result.TaxpayerName)
fmt.Printf("Type: %s\n", result.TaxpayerType)
fmt.Printf("Status: %s\n", result.Status)
```

### TCC Verification

```go
result, err := client.VerifyTCC(ctx, "TCC123456")
if err != nil {
    log.Fatal(err)
}

if result.IsCurrentlyValid() {
    fmt.Printf("TCC valid until: %s\n", result.ExpiryDate)
    days := result.DaysUntilExpiry()
    fmt.Printf("Days remaining: %d\n", days)
}
```

### E-slip Validation

```go
result, err := client.ValidateEslip(ctx, "1234567890")
if err != nil {
    log.Fatal(err)
}

if result.IsPaid() {
    fmt.Printf("Payment confirmed: %s %.2f\n", result.Currency, result.Amount)
}
```

### NIL Return Filing

```go
result, err := client.FileNILReturn(ctx, &kra.NILReturnRequest{
    PINNumber:    "P051234567A",
    ObligationID: "OBL123456",
    Period:       "202401",
})
if err != nil {
    log.Fatal(err)
}

if result.IsAccepted() {
    fmt.Printf("Reference: %s\n", result.ReferenceNumber)
}
```

### Batch Operations

```go
pins := []string{"P051234567A", "P051234567B", "P051234567C"}

results, err := client.VerifyPINsBatch(ctx, pins)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("%s: %v\n", result.PINNumber, result.IsValid)
}
```

### Configuration

```go
client, err := kra.NewClient(
    kra.WithAPIKey(os.Getenv("KRA_API_KEY")),
    kra.WithBaseURL("https://api.kra.go.ke/gavaconnect/v1"),
    kra.WithTimeout(30 * time.Second),
    kra.WithRetry(3, time.Second, 32*time.Second),
    kra.WithCache(true, 1*time.Hour),
    kra.WithRateLimit(100, time.Minute),
    kra.WithDebug(true),
)
```

### Context and Cancellation

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := client.VerifyPIN(ctx, "P051234567A")

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(5 * time.Second)
    cancel()
}()

result, err := client.VerifyPIN(ctx, "P051234567A")
```

### Error Handling

```go
result, err := client.VerifyPIN(ctx, "P051234567A")
if err != nil {
    switch e := err.(type) {
    case *kra.ValidationError:
        fmt.Printf("Validation error: %s\n", e.Message)
    case *kra.AuthenticationError:
        fmt.Printf("Authentication failed: %s\n", e.Message)
    case *kra.RateLimitError:
        fmt.Printf("Rate limited. Retry after: %v\n", e.RetryAfter)
    case *kra.APIError:
        fmt.Printf("API error: %s (status: %d)\n", e.Message, e.StatusCode)
    default:
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

## API Reference

### Client

```go
type Client struct {
    // contains filtered or unexported fields
}

func NewClient(opts ...Option) (*Client, error)
func (c *Client) Close() error
func (c *Client) VerifyPIN(ctx context.Context, pin string) (*PINVerificationResult, error)
func (c *Client) VerifyTCC(ctx context.Context, tcc string) (*TCCVerificationResult, error)
func (c *Client) ValidateEslip(ctx context.Context, eslip string) (*EslipValidationResult, error)
func (c *Client) FileNILReturn(ctx context.Context, req *NILReturnRequest) (*NILReturnResult, error)
func (c *Client) GetTaxpayerDetails(ctx context.Context, pin string) (*TaxpayerDetails, error)
func (c *Client) VerifyPINsBatch(ctx context.Context, pins []string) ([]*PINVerificationResult, error)
func (c *Client) VerifyTCCsBatch(ctx context.Context, tccs []string) ([]*TCCVerificationResult, error)
```

### Configuration Options

```go
func WithAPIKey(key string) Option
func WithBaseURL(url string) Option
func WithTimeout(timeout time.Duration) Option
func WithRetry(maxRetries int, initialDelay, maxDelay time.Duration) Option
func WithCache(enabled bool, ttl time.Duration) Option
func WithRateLimit(maxRequests int, window time.Duration) Option
func WithDebug(debug bool) Option
func WithHTTPClient(client *http.Client) Option
```

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for contribution guidelines.

## License

MIT License - see [LICENSE](../../LICENSE) for details.

## Support

- Issues: https://github.com/kra-connect/go-sdk/issues
- Documentation: https://pkg.go.dev/github.com/kra-connect/go-sdk
- Examples: [examples/](examples/)

## Related SDKs

- [Python SDK](https://github.com/kra-connect/python-sdk)
- [Node.js SDK](https://github.com/kra-connect/node-sdk)
- [PHP SDK](https://github.com/kra-connect/php-sdk)
- [Flutter SDK](https://github.com/kra-connect/flutter-sdk)

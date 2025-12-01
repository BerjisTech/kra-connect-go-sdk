// Package kra provides a Go SDK for the Kenya Revenue Authority's GavaConnect API.
//
// The SDK provides a simple, idiomatic Go interface for interacting with KRA services
// including PIN verification, TCC verification, e-slip validation, NIL return filing,
// and taxpayer details retrieval.
//
// # Features
//
//   - PIN Verification - Verify KRA PIN numbers
//   - TCC Verification - Check Tax Compliance Certificates
//   - E-slip Validation - Validate electronic payment slips
//   - NIL Returns - File NIL returns programmatically
//   - Taxpayer Details - Retrieve taxpayer information
//   - Type Safety - Full Go type safety
//   - Context Support - Context-aware operations for cancellation and timeouts
//   - Retry Logic - Automatic retry with exponential backoff
//   - Caching - Response caching with TTL
//   - Rate Limiting - Built-in token bucket rate limiter
//   - Goroutine Safe - Thread-safe operations
//   - Zero Dependencies - Only stdlib for core functionality
//
// # Quick Start
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "log"
//	    "os"
//
//	    kra "github.com/BerjisTech/kra-connect-go-sdk"
//	)
//
//	func main() {
//	    // Create client
//	    client, err := kra.NewClient(
//	        kra.WithAPIKey(os.Getenv("KRA_API_KEY")),
//	    )
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    defer client.Close()
//
//	    // Verify PIN
//	    ctx := context.Background()
//	    result, err := client.VerifyPIN(ctx, "P051234567A")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    if result.IsValid {
//	        fmt.Printf("Valid PIN: %s\n", result.TaxpayerName)
//	    }
//	}
//
// # Configuration
//
// The client can be configured using functional options:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey(os.Getenv("KRA_API_KEY")),
//	    kra.WithBaseURL("https://api.kra.go.ke/gavaconnect/v1"),
//	    kra.WithTimeout(30 * time.Second),
//	    kra.WithRetry(3, time.Second, 32*time.Second),
//	    kra.WithCache(true, 1*time.Hour),
//	    kra.WithRateLimit(100, time.Minute),
//	    kra.WithDebug(true),
//	)
//
// # Context and Cancellation
//
// All API methods accept a context parameter for cancellation and timeouts:
//
//	// With timeout
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	result, err := client.VerifyPIN(ctx, "P051234567A")
//
//	// With cancellation
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    time.Sleep(5 * time.Second)
//	    cancel()
//	}()
//	result, err := client.VerifyPIN(ctx, "P051234567A")
//
// # Error Handling
//
// The SDK provides specific error types for different scenarios:
//
//	result, err := client.VerifyPIN(ctx, "P051234567A")
//	if err != nil {
//	    switch e := err.(type) {
//	    case *kra.ValidationError:
//	        fmt.Printf("Validation error: %s\n", e.Message)
//	    case *kra.AuthenticationError:
//	        fmt.Printf("Authentication failed: %s\n", e.Message)
//	    case *kra.RateLimitError:
//	        fmt.Printf("Rate limited. Retry after: %v\n", e.RetryAfter)
//	    case *kra.APIError:
//	        fmt.Printf("API error: %s (status: %d)\n", e.Message, e.StatusCode)
//	    default:
//	        fmt.Printf("Error: %v\n", err)
//	    }
//	    return
//	}
//
// # Batch Operations
//
// Process multiple requests concurrently:
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
//
// # Caching
//
// Results are cached automatically with configurable TTLs:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey(apiKey),
//	    kra.WithCustomCacheTTLs(
//	        2*time.Hour,    // PIN verification
//	        1*time.Hour,    // TCC verification
//	        30*time.Minute, // E-slip validation
//	        4*time.Hour,    // Taxpayer details
//	        48*time.Hour,   // NIL return
//	    ),
//	)
//
// # Rate Limiting
//
// The SDK includes built-in rate limiting using a token bucket algorithm:
//
//	client, err := kra.NewClient(
//	    kra.WithAPIKey(apiKey),
//	    kra.WithRateLimit(100, time.Minute), // 100 requests per minute
//	)
//
// # Thread Safety
//
// The client is goroutine-safe and can be used concurrently from multiple goroutines:
//
//	var wg sync.WaitGroup
//	for _, pin := range pins {
//	    wg.Add(1)
//	    go func(p string) {
//	        defer wg.Done()
//	        result, err := client.VerifyPIN(ctx, p)
//	        // Handle result...
//	    }(pin)
//	}
//	wg.Wait()
//
// For more information and examples, see https://github.com/BerjisTech/kra-connect-go-sdk
package kra

const (
	// Version is the current version of the SDK
	Version = "0.1.3"

	// DefaultBaseURL is the default KRA API base URL
	DefaultBaseURL = "https://api.kra.go.ke/gavaconnect/v1"
)

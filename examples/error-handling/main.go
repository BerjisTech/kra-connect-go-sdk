//go:build ignore

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	kra "github.com/BerjisTech/kra-connect-go-sdk"
)

func main() {
	// Example 1: Handle Missing API Key
	fmt.Println("=== Example 1: Missing API Key ===")
	_, err := kra.NewClient()
	if err != nil {
		fmt.Printf("Expected error: %v\n\n", err)
	}

	// Example 2: Handle Invalid Configuration
	fmt.Println("=== Example 2: Invalid Configuration ===")
	_, err = kra.NewClient(
		kra.WithAPIKey("valid-api-key-here"),
		kra.WithTimeout(-5*time.Second), // Invalid timeout
	)
	if err != nil {
		fmt.Printf("Configuration error: %v\n\n", err)
	}

	// Get API key from environment for remaining examples
	apiKey := os.Getenv("KRA_API_KEY")
	if apiKey == "" {
		log.Fatal("KRA_API_KEY environment variable is required")
	}

	// Create client
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithDebug(true),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 3: Handle Validation Errors
	fmt.Println("=== Example 3: Validation Errors ===")
	handleValidationErrors(ctx, client)

	// Example 4: Handle Authentication Errors
	fmt.Println("=== Example 4: Authentication Errors ===")
	handleAuthenticationErrors()

	// Example 5: Handle Rate Limit Errors
	fmt.Println("=== Example 5: Rate Limit Errors ===")
	handleRateLimitErrors(ctx, client)

	// Example 6: Handle Timeout Errors
	fmt.Println("=== Example 6: Timeout Errors ===")
	handleTimeoutErrors(apiKey)

	// Example 7: Handle Network Errors
	fmt.Println("=== Example 7: Network Errors ===")
	handleNetworkErrors(apiKey)

	// Example 8: Using errors.Is and errors.As
	fmt.Println("=== Example 8: Using errors.Is and errors.As ===")
	demonstrateErrorInspection(ctx, client)
}

func handleValidationErrors(ctx context.Context, client *kra.Client) {
	invalidPINs := []string{
		"",                  // Empty
		"INVALID",           // Invalid format
		"P12345",            // Too short
		"P0512345678901234", // Too long
	}

	for _, pin := range invalidPINs {
		_, err := client.VerifyPIN(ctx, pin)
		if err != nil {
			// Type assertion to get specific error details
			if validationErr, ok := err.(*kra.ValidationError); ok {
				fmt.Printf("Validation Error - Field: %s, Message: %s\n",
					validationErr.Field,
					validationErr.Message,
				)
			} else {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
	fmt.Println()
}

func handleAuthenticationErrors() {
	// Create client with invalid API key
	client, err := kra.NewClient(
		kra.WithAPIKey("invalid-api-key-1234567890"),
	)
	if err != nil {
		fmt.Printf("Client creation error: %v\n\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	_, err = client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		if authErr, ok := err.(*kra.AuthenticationError); ok {
			fmt.Printf("Authentication Error: %s (Status: %d)\n",
				authErr.Message,
				authErr.StatusCode,
			)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
	fmt.Println()
}

func handleRateLimitErrors(ctx context.Context, client *kra.Client) {
	// Create a client with very low rate limit
	restrictedClient, err := kra.NewClient(
		kra.WithAPIKey(os.Getenv("KRA_API_KEY")),
		kra.WithRateLimit(2, 10*time.Second), // Only 2 requests per 10 seconds
	)
	if err != nil {
		fmt.Printf("Client creation error: %v\n\n", err)
		return
	}
	defer restrictedClient.Close()

	// Make several requests quickly
	for i := 1; i <= 5; i++ {
		_, err := restrictedClient.VerifyPIN(ctx, fmt.Sprintf("P%09dA", i))
		if err != nil {
			if rateLimitErr, ok := err.(*kra.RateLimitError); ok {
				fmt.Printf("Rate Limit Exceeded: Retry after %v\n",
					rateLimitErr.RetryAfter,
				)
				fmt.Printf("Limit: %d requests per %v\n",
					rateLimitErr.Limit,
					rateLimitErr.Window,
				)
				break
			}
		}
		fmt.Printf("Request %d succeeded\n", i)
	}
	fmt.Println()
}

func handleTimeoutErrors(apiKey string) {
	// Create client with very short timeout
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithTimeout(1*time.Millisecond), // Unrealistically short
	)
	if err != nil {
		fmt.Printf("Client creation error: %v\n\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	_, err = client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		if timeoutErr, ok := err.(*kra.TimeoutError); ok {
			fmt.Printf("Timeout Error: %s\n", timeoutErr.Message)
			fmt.Printf("Endpoint: %s, Timeout: %v, Attempt: %d\n",
				timeoutErr.Endpoint,
				timeoutErr.Timeout,
				timeoutErr.AttemptNumber,
			)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
	fmt.Println()
}

func handleNetworkErrors(apiKey string) {
	// Create client with invalid base URL
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithBaseURL("https://invalid-url-that-does-not-exist.local"),
	)
	if err != nil {
		fmt.Printf("Client creation error: %v\n\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	_, err = client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		if networkErr, ok := err.(*kra.NetworkError); ok {
			fmt.Printf("Network Error: %s\n", networkErr.Message)
			fmt.Printf("Endpoint: %s\n", networkErr.Endpoint)
			if networkErr.Err != nil {
				fmt.Printf("Underlying error: %v\n", networkErr.Err)
			}
		} else {
			fmt.Printf("Error: %v\n", err)
		}
	}
	fmt.Println()
}

func demonstrateErrorInspection(ctx context.Context, client *kra.Client) {
	// Try with invalid PIN
	_, err := client.VerifyPIN(ctx, "INVALID-PIN")
	if err != nil {
		// Method 1: Type switch
		fmt.Println("Method 1: Type switch")
		switch e := err.(type) {
		case *kra.ValidationError:
			fmt.Printf("  Validation error on field '%s': %s\n", e.Field, e.Message)
		case *kra.AuthenticationError:
			fmt.Printf("  Authentication failed: %s\n", e.Message)
		case *kra.RateLimitError:
			fmt.Printf("  Rate limited. Retry after: %v\n", e.RetryAfter)
		case *kra.TimeoutError:
			fmt.Printf("  Request timed out: %s\n", e.Message)
		case *kra.NetworkError:
			fmt.Printf("  Network error: %s\n", e.Message)
		case *kra.APIError:
			fmt.Printf("  API error: %s (Status: %d)\n", e.Message, e.StatusCode)
		default:
			fmt.Printf("  Unknown error: %v\n", err)
		}
		fmt.Println()

		// Method 2: errors.As
		fmt.Println("Method 2: errors.As")
		var validationErr *kra.ValidationError
		if errors.As(err, &validationErr) {
			fmt.Printf("  Found validation error: %s\n", validationErr.Message)
			fmt.Printf("  Field: %s\n", validationErr.Field)
			fmt.Printf("  Details: %v\n", validationErr.Details)
		}
		fmt.Println()

		// Method 3: Check error hierarchy
		fmt.Println("Method 3: Error hierarchy")
		var kraErr *kra.SDKError
		if errors.As(err, &kraErr) {
			fmt.Printf("  KRA Error: %s\n", kraErr.Message)
			fmt.Printf("  Status Code: %d\n", kraErr.StatusCode)
			fmt.Printf("  Details: %v\n", kraErr.Details)
			if kraErr.Err != nil {
				fmt.Printf("  Wrapped error: %v\n", kraErr.Err)
			}
		}
		fmt.Println()
	}
}

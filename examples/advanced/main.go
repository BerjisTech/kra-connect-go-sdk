package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	kra "github.com/kra-connect/go-sdk"
)

func main() {
	apiKey := os.Getenv("KRA_API_KEY")
	if apiKey == "" {
		log.Fatal("KRA_API_KEY environment variable is required")
	}

	// Example 1: Custom Configuration
	fmt.Println("=== Example 1: Custom Configuration ===")
	customConfigExample(apiKey)

	// Example 2: Context with Timeout
	fmt.Println("=== Example 2: Context with Timeout ===")
	contextTimeoutExample(apiKey)

	// Example 3: Context with Cancellation
	fmt.Println("=== Example 3: Context with Cancellation ===")
	contextCancellationExample(apiKey)

	// Example 4: Concurrent Operations
	fmt.Println("=== Example 4: Concurrent Operations ===")
	concurrentExample(apiKey)

	// Example 5: Cache Management
	fmt.Println("=== Example 5: Cache Management ===")
	cacheManagementExample(apiKey)

	// Example 6: Custom Cache TTLs
	fmt.Println("=== Example 6: Custom Cache TTLs ===")
	customCacheTTLExample(apiKey)
}

func customConfigExample(apiKey string) {
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithBaseURL("https://api.kra.go.ke/gavaconnect/v1"),
		kra.WithTimeout(60*time.Second),
		kra.WithRetry(5, 2*time.Second, 64*time.Second),
		kra.WithRateLimit(200, 1*time.Minute),
		kra.WithCache(true, 2*time.Hour),
		kra.WithDebug(true),
	)
	if err != nil {
		log.Printf("Failed to create client: %v\n\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	result, err := client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		log.Printf("Error: %v\n\n", err)
		return
	}

	fmt.Printf("PIN %s is %v\n\n", result.PINNumber, result.IsValid)
}

func contextTimeoutExample(apiKey string) {
	client, err := kra.NewClient(kra.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Failed to create client: %v\n\n", err)
		return
	}
	defer client.Close()

	// Create context with 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Request timed out after 5 seconds")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println()
		return
	}

	fmt.Printf("Verification completed within timeout: %v\n\n", result.IsValid)
}

func contextCancellationExample(apiKey string) {
	client, err := kra.NewClient(kra.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Failed to create client: %v\n\n", err)
		return
	}
	defer client.Close()

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 2 seconds
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("Cancelling request...")
		cancel()
	}()

	result, err := client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		if ctx.Err() == context.Canceled {
			fmt.Println("Request was cancelled")
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println()
		return
	}

	fmt.Printf("Verification completed: %v\n\n", result.IsValid)
}

func concurrentExample(apiKey string) {
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithRateLimit(50, 1*time.Minute),
	)
	if err != nil {
		log.Printf("Failed to create client: %v\n\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Simulate concurrent requests from multiple users
	pins := []string{
		"P051234567A",
		"P051234567B",
		"P051234567C",
		"P051234567D",
		"P051234567E",
	}

	tccs := []string{
		"TCC123456",
		"TCC123457",
		"TCC123458",
	}

	var wg sync.WaitGroup
	results := make(chan string, len(pins)+len(tccs))

	// Verify PINs concurrently
	for _, pin := range pins {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			result, err := client.VerifyPIN(ctx, p)
			if err != nil {
				results <- fmt.Sprintf("PIN %s: ERROR - %v", p, err)
				return
			}
			results <- fmt.Sprintf("PIN %s: %v (%s)", p, result.IsValid, result.TaxpayerName)
		}(pin)
	}

	// Verify TCCs concurrently
	for _, tcc := range tccs {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			result, err := client.VerifyTCC(ctx, t)
			if err != nil {
				results <- fmt.Sprintf("TCC %s: ERROR - %v", t, err)
				return
			}
			status := "Invalid"
			if result.IsCurrentlyValid() {
				status = "Valid"
			}
			results <- fmt.Sprintf("TCC %s: %s", t, status)
		}(tcc)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Print results
	for result := range results {
		fmt.Println(result)
	}
	fmt.Println()
}

func cacheManagementExample(apiKey string) {
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithCache(true, 1*time.Hour),
	)
	if err != nil {
		log.Printf("Failed to create client: %v\n\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// First request - will hit the API
	fmt.Println("First request (API call)...")
	start := time.Now()
	result1, err := client.VerifyPIN(ctx, "P051234567A")
	duration1 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v\n\n", err)
		return
	}
	fmt.Printf("Result: %v (took %v)\n", result1.IsValid, duration1)

	// Second request - will use cache
	fmt.Println("Second request (cached)...")
	start = time.Now()
	result2, err := client.VerifyPIN(ctx, "P051234567A")
	duration2 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v\n\n", err)
		return
	}
	fmt.Printf("Result: %v (took %v)\n", result2.IsValid, duration2)

	fmt.Printf("Cache speedup: %.2fx faster\n", float64(duration1)/float64(duration2))

	// Clear cache
	fmt.Println("\nClearing cache...")
	client.ClearCache()

	// Third request - will hit the API again
	fmt.Println("Third request after cache clear (API call)...")
	start = time.Now()
	result3, err := client.VerifyPIN(ctx, "P051234567A")
	duration3 := time.Since(start)
	if err != nil {
		log.Printf("Error: %v\n\n", err)
		return
	}
	fmt.Printf("Result: %v (took %v)\n\n", result3.IsValid, duration3)
}

func customCacheTTLExample(apiKey string) {
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithCustomCacheTTLs(
			4*time.Hour,    // PIN verification - 4 hours
			2*time.Hour,    // TCC verification - 2 hours
			30*time.Minute, // E-slip validation - 30 minutes
			8*time.Hour,    // Taxpayer details - 8 hours
			48*time.Hour,   // NIL return - 48 hours
		),
	)
	if err != nil {
		log.Printf("Failed to create client: %v\n\n", err)
		return
	}
	defer client.Close()

	ctx := context.Background()

	// PIN verification will be cached for 4 hours
	pinResult, err := client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("PIN result (cached for 4 hours): %v\n", pinResult.IsValid)
	}

	// TCC verification will be cached for 2 hours
	tccResult, err := client.VerifyTCC(ctx, "TCC123456")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("TCC result (cached for 2 hours): %v\n", tccResult.IsValid)
	}

	// E-slip validation will be cached for 30 minutes
	eslipResult, err := client.ValidateEslip(ctx, "1234567890")
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("E-slip result (cached for 30 minutes): %v\n", eslipResult.IsValid)
	}

	fmt.Println("\nEach operation type has its own cache TTL based on data volatility")
	fmt.Println()
}

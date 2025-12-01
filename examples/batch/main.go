//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	kra "github.com/BerjisTech/kra-connect-go-sdk"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("KRA_API_KEY")
	if apiKey == "" {
		log.Fatal("KRA_API_KEY environment variable is required")
	}

	// Create client
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
		kra.WithRateLimit(50, 1*time.Minute), // Adjust rate limit for batch operations
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 1: Batch PIN Verification
	fmt.Println("=== Batch PIN Verification ===")
	pins := []string{
		"P051234567A",
		"P051234567B",
		"P051234567C",
		"P051234567D",
		"P051234567E",
	}

	start := time.Now()
	pinResults, err := client.VerifyPINsBatch(ctx, pins)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Batch PIN verification failed: %v", err)
	} else {
		fmt.Printf("Verified %d PINs in %v\n\n", len(pinResults), duration)
		for i, result := range pinResults {
			if result != nil {
				fmt.Printf("%d. %s: %v (%s)\n",
					i+1,
					result.PINNumber,
					result.IsValid,
					result.TaxpayerName,
				)
			}
		}
		fmt.Println()
	}

	// Example 2: Batch TCC Verification
	fmt.Println("=== Batch TCC Verification ===")
	tccs := []string{
		"TCC123456",
		"TCC123457",
		"TCC123458",
		"TCC123459",
		"TCC123460",
	}

	start = time.Now()
	tccResults, err := client.VerifyTCCsBatch(ctx, tccs)
	duration = time.Since(start)

	if err != nil {
		log.Printf("Batch TCC verification failed: %v", err)
	} else {
		fmt.Printf("Verified %d TCCs in %v\n\n", len(tccResults), duration)
		for i, result := range tccResults {
			if result != nil {
				status := "Invalid"
				if result.IsCurrentlyValid() {
					status = fmt.Sprintf("Valid (expires in %d days)", result.DaysUntilExpiry())
				} else if result.IsValid && result.IsExpired {
					status = "Expired"
				}

				fmt.Printf("%d. %s: %s\n",
					i+1,
					result.TCCNumber,
					status,
				)
			}
		}
		fmt.Println()
	}

	// Example 3: Process Large Dataset with Progress Tracking
	fmt.Println("=== Process Large Dataset ===")
	largePINList := generatePINList(100) // Generate 100 test PINs

	processed := 0
	validCount := 0
	invalidCount := 0

	// Process in batches of 10
	batchSize := 10
	totalBatches := (len(largePINList) + batchSize - 1) / batchSize

	start = time.Now()
	for i := 0; i < len(largePINList); i += batchSize {
		end := i + batchSize
		if end > len(largePINList) {
			end = len(largePINList)
		}

		batch := largePINList[i:end]
		results, err := client.VerifyPINsBatch(ctx, batch)
		if err != nil {
			log.Printf("Batch %d failed: %v", (i/batchSize)+1, err)
			continue
		}

		for _, result := range results {
			if result != nil {
				processed++
				if result.IsValid {
					validCount++
				} else {
					invalidCount++
				}
			}
		}

		// Show progress
		currentBatch := (i / batchSize) + 1
		progress := float64(currentBatch) / float64(totalBatches) * 100
		fmt.Printf("Progress: %.1f%% (%d/%d batches)\n", progress, currentBatch, totalBatches)
	}

	duration = time.Since(start)
	fmt.Printf("\nProcessed %d PINs in %v\n", processed, duration)
	fmt.Printf("Valid: %d, Invalid: %d\n", validCount, invalidCount)
	fmt.Printf("Average time per PIN: %v\n", duration/time.Duration(processed))
}

// generatePINList generates a list of test PIN numbers
func generatePINList(count int) []string {
	pins := make([]string, count)
	for i := 0; i < count; i++ {
		pins[i] = fmt.Sprintf("P%09dA", i)
	}
	return pins
}

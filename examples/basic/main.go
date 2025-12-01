package main

import (
	"context"
	"fmt"
	"log"
	"os"

	kra "github.com/BerjisTech/kra-connect-go-sdk"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("KRA_API_KEY")
	if apiKey == "" {
		log.Fatal("KRA_API_KEY environment variable is required")
	}

	// Create client with default configuration
	client, err := kra.NewClient(
		kra.WithAPIKey(apiKey),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Example 1: Verify PIN
	fmt.Println("=== PIN Verification ===")
	pinResult, err := client.VerifyPIN(ctx, "P051234567A")
	if err != nil {
		log.Printf("PIN verification failed: %v", err)
	} else {
		fmt.Printf("PIN: %s\n", pinResult.PINNumber)
		fmt.Printf("Valid: %v\n", pinResult.IsValid)
		fmt.Printf("Name: %s\n", pinResult.TaxpayerName)
		fmt.Printf("Type: %s\n", pinResult.TaxpayerType)
		fmt.Printf("Status: %s\n", pinResult.Status)
		fmt.Printf("Active: %v\n", pinResult.IsActive())
		fmt.Println()
	}

	// Example 2: Verify TCC
	fmt.Println("=== TCC Verification ===")
	tccResult, err := client.VerifyTCC(ctx, "TCC123456")
	if err != nil {
		log.Printf("TCC verification failed: %v", err)
	} else {
		fmt.Printf("TCC: %s\n", tccResult.TCCNumber)
		fmt.Printf("Valid: %v\n", tccResult.IsValid)
		fmt.Printf("Name: %s\n", tccResult.TaxpayerName)
		fmt.Printf("PIN: %s\n", tccResult.PINNumber)
		fmt.Printf("Issue Date: %s\n", tccResult.IssueDate)
		fmt.Printf("Expiry Date: %s\n", tccResult.ExpiryDate)
		fmt.Printf("Currently Valid: %v\n", tccResult.IsCurrentlyValid())
		fmt.Printf("Days Until Expiry: %d\n", tccResult.DaysUntilExpiry())
		fmt.Println()
	}

	// Example 3: Validate E-slip
	fmt.Println("=== E-slip Validation ===")
	eslipResult, err := client.ValidateEslip(ctx, "1234567890")
	if err != nil {
		log.Printf("E-slip validation failed: %v", err)
	} else {
		fmt.Printf("E-slip: %s\n", eslipResult.EslipNumber)
		fmt.Printf("Valid: %v\n", eslipResult.IsValid)
		fmt.Printf("Amount: %.2f %s\n", eslipResult.Amount, eslipResult.Currency)
		fmt.Printf("Payment Date: %s\n", eslipResult.PaymentDate)
		fmt.Printf("Status: %s\n", eslipResult.Status)
		fmt.Printf("Paid: %v\n", eslipResult.IsPaid())
		fmt.Println()
	}

	// Example 4: File NIL Return
	fmt.Println("=== NIL Return Filing ===")
	nilResult, err := client.FileNILReturn(ctx, &kra.NILReturnRequest{
		PINNumber:    "P051234567A",
		ObligationID: "OBL123456",
		Period:       "202401",
	})
	if err != nil {
		log.Printf("NIL return filing failed: %v", err)
	} else {
		fmt.Printf("Success: %v\n", nilResult.Success)
		fmt.Printf("Reference: %s\n", nilResult.ReferenceNumber)
		fmt.Printf("Status: %s\n", nilResult.Status)
		fmt.Printf("Message: %s\n", nilResult.Message)
		fmt.Printf("Accepted: %v\n", nilResult.IsAccepted())
		fmt.Println()
	}

	// Example 5: Get Taxpayer Details
	fmt.Println("=== Taxpayer Details ===")
	details, err := client.GetTaxpayerDetails(ctx, "P051234567A")
	if err != nil {
		log.Printf("Failed to get taxpayer details: %v", err)
	} else {
		fmt.Printf("PIN: %s\n", details.PINNumber)
		fmt.Printf("Name: %s\n", details.GetDisplayName())
		fmt.Printf("Type: %s\n", details.TaxpayerType)
		fmt.Printf("Status: %s\n", details.Status)
		fmt.Printf("Email: %s\n", details.EmailAddress)
		fmt.Printf("Phone: %s\n", details.PhoneNumber)
		fmt.Printf("Obligations: %d\n", len(details.Obligations))

		if len(details.Obligations) > 0 {
			fmt.Println("\nTax Obligations:")
			for _, ob := range details.Obligations {
				fmt.Printf("  - %s (%s)\n", ob.Description, ob.ObligationType)
				fmt.Printf("    Status: %s, Active: %v\n", ob.Status, ob.IsActive)
				if ob.NextFilingDate != "" {
					fmt.Printf("    Next Filing: %s\n", ob.NextFilingDate)
				}
			}
		}
	}
}

// Package main demonstrates basic usage of the Go Invoice Ninja SDK.
//
// This example shows how to:
// - Create a client with configuration options
// - List payments with pagination
// - Handle errors properly
//
// Run with: go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

func main() {
	// Get API token from environment variable
	token := os.Getenv("INVOICE_NINJA_TOKEN")
	if token == "" {
		log.Fatal("INVOICE_NINJA_TOKEN environment variable is required")
	}

	// Create a new client with options
	client := invoiceninja.NewClient(token,
		invoiceninja.WithTimeout(30*time.Second),
		// Uncomment for self-hosted instances:
		// invoiceninja.WithBaseURL("https://your-instance.com"),
	)

	ctx := context.Background()

	// Example 1: List payments
	fmt.Println("=== Listing Payments ===")
	payments, err := client.Payments.List(ctx, &invoiceninja.PaymentListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		// Handle specific error types
		if apiErr, ok := err.(*invoiceninja.APIError); ok {
			if apiErr.IsUnauthorized() {
				log.Fatal("Invalid API token")
			}
			if apiErr.IsRateLimited() {
				log.Fatal("Rate limited, try again later")
			}
			log.Fatalf("API error: %v", apiErr)
		}
		log.Fatalf("Error listing payments: %v", err)
	}

	fmt.Printf("Found %d payments\n", len(payments.Data))
	for _, p := range payments.Data {
		fmt.Printf("  - Payment %s: $%.2f (Client: %s)\n", p.Number, p.Amount, p.ClientID)
	}

	// Example 2: List invoices
	fmt.Println("\n=== Listing Invoices ===")
	invoices, err := client.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		log.Fatalf("Error listing invoices: %v", err)
	}

	fmt.Printf("Found %d invoices\n", len(invoices.Data))
	for _, inv := range invoices.Data {
		fmt.Printf("  - Invoice %s: $%.2f (Status: %s)\n", inv.Number, inv.Amount, inv.StatusID)
	}

	// Example 3: List clients
	fmt.Println("\n=== Listing Clients ===")
	clients, err := client.Clients.List(ctx, &invoiceninja.ClientListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		log.Fatalf("Error listing clients: %v", err)
	}

	fmt.Printf("Found %d clients\n", len(clients.Data))
	for _, c := range clients.Data {
		fmt.Printf("  - Client %s: %s\n", c.Number, c.Name)
	}

	fmt.Println("\n✅ Basic example completed successfully!")
}

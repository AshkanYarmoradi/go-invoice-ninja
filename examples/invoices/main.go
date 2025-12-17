// Package main demonstrates creating and managing invoices.
//
// This example shows how to:
// - Create an invoice with line items
// - Apply discounts and taxes
// - Update invoice status
// - Download invoice PDF
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
	token := os.Getenv("INVOICE_NINJA_TOKEN")
	if token == "" {
		log.Fatal("INVOICE_NINJA_TOKEN environment variable is required")
	}

	// For this example, you also need a client ID
	clientID := os.Getenv("INVOICE_NINJA_CLIENT_ID")
	if clientID == "" {
		log.Fatal("INVOICE_NINJA_CLIENT_ID environment variable is required")
	}

	client := invoiceninja.NewClient(token)
	ctx := context.Background()

	// Create an invoice with line items
	fmt.Println("=== Creating Invoice ===")
	invoice, err := client.Invoices.Create(ctx, &invoiceninja.Invoice{
		ClientID: clientID,
		Date:     time.Now().Format("2006-01-02"),
		DueDate:  time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
		LineItems: []invoiceninja.LineItem{
			{
				ProductKey: "Consulting",
				Notes:      "Professional consulting services",
				Quantity:   10,
				Cost:       150.00,
			},
			{
				ProductKey: "Development",
				Notes:      "Custom software development",
				Quantity:   20,
				Cost:       125.00,
			},
			{
				ProductKey: "Support",
				Notes:      "Technical support hours",
				Quantity:   5,
				Cost:       75.00,
			},
		},
		PublicNotes: "Thank you for your business!",
		Terms:       "Payment due within 30 days",
		Footer:      "Please make checks payable to Acme Corp",
	})
	if err != nil {
		log.Fatalf("Error creating invoice: %v", err)
	}

	fmt.Printf("Created invoice: %s\n", invoice.Number)
	fmt.Printf("  Amount: $%.2f\n", invoice.Amount)
	fmt.Printf("  Balance: $%.2f\n", invoice.Balance)

	// Get the invoice details
	fmt.Println("\n=== Getting Invoice Details ===")
	inv, err := client.Invoices.Get(ctx, invoice.ID)
	if err != nil {
		log.Fatalf("Error getting invoice: %v", err)
	}

	fmt.Printf("Invoice %s details:\n", inv.Number)
	fmt.Printf("  Client ID: %s\n", inv.ClientID)
	fmt.Printf("  Date: %s\n", inv.Date)
	fmt.Printf("  Due Date: %s\n", inv.DueDate)
	fmt.Printf("  Status: %s\n", inv.StatusID)
	fmt.Printf("  Line Items: %d\n", len(inv.LineItems))

	// Update the invoice
	fmt.Println("\n=== Updating Invoice ===")
	inv.PublicNotes = "Thank you for choosing our services!"
	updated, err := client.Invoices.Update(ctx, inv.ID, inv)
	if err != nil {
		log.Fatalf("Error updating invoice: %v", err)
	}
	fmt.Printf("Updated invoice notes: %s\n", updated.PublicNotes)

	fmt.Println("\n✅ Invoice example completed successfully!")
}

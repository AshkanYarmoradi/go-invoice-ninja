// Package main demonstrates webhook handling.
//
// This example shows how to:
// - Set up a webhook endpoint
// - Register event handlers
// - Handle different webhook events
//
// Run with: go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

func main() {
	webhookSecret := os.Getenv("INVOICE_NINJA_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("Warning: INVOICE_NINJA_WEBHOOK_SECRET not set, signature verification disabled")
		webhookSecret = "" // Empty string disables signature verification
	}

	// Create a webhook handler
	webhookHandler := invoiceninja.NewWebhookHandler(webhookSecret)

	// Register handlers for different event types using convenience methods
	webhookHandler.OnPaymentCreated(handlePaymentCreated)
	webhookHandler.OnInvoiceCreated(handleInvoiceCreated)
	webhookHandler.OnClientCreated(handleClientCreated)

	// You can also use the generic On method for any event type
	webhookHandler.On("invoice.paid", func(event *invoiceninja.WebhookEvent) error {
		log.Printf("Invoice paid event received")
		prettyJSON, _ := json.MarshalIndent(json.RawMessage(event.Data), "", "  ")
		log.Printf("Event data:\n%s", prettyJSON)
		return nil
	})

	// Use the built-in HTTP handler
	http.HandleFunc("/webhook", webhookHandler.HandleRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting webhook server on port %s...\n", port)
	fmt.Println("Send webhooks to: http://localhost:" + port + "/webhook")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handlePaymentCreated(event *invoiceninja.WebhookEvent) error {
	log.Printf("Payment created event received")

	// Parse the payment data using the helper method
	payment, err := event.ParsePayment()
	if err != nil {
		log.Printf("Could not parse payment data: %v", err)
		// Still log the raw data for debugging
		prettyJSON, _ := json.MarshalIndent(json.RawMessage(event.Data), "", "  ")
		log.Printf("Raw event data:\n%s", prettyJSON)
		return nil // Don't return error to acknowledge receipt
	}

	log.Printf("Payment ID: %s", payment.ID)
	log.Printf("Payment Amount: $%.2f", payment.Amount)
	log.Printf("Payment Number: %s", payment.Number)

	// Process the payment...
	// e.g., update your database, send notifications, etc.
	return nil
}

func handleInvoiceCreated(event *invoiceninja.WebhookEvent) error {
	log.Printf("Invoice created event received")

	// Parse the invoice data using the helper method
	invoice, err := event.ParseInvoice()
	if err != nil {
		log.Printf("Could not parse invoice data: %v", err)
		prettyJSON, _ := json.MarshalIndent(json.RawMessage(event.Data), "", "  ")
		log.Printf("Raw event data:\n%s", prettyJSON)
		return nil
	}

	log.Printf("Invoice ID: %s", invoice.ID)
	log.Printf("Invoice Number: %s", invoice.Number)
	log.Printf("Invoice Amount: $%.2f", invoice.Amount)

	// Process the invoice...
	return nil
}

func handleClientCreated(event *invoiceninja.WebhookEvent) error {
	log.Printf("Client created event received")

	// Parse the client data using the helper method
	client, err := event.ParseClient()
	if err != nil {
		log.Printf("Could not parse client data: %v", err)
		prettyJSON, _ := json.MarshalIndent(json.RawMessage(event.Data), "", "  ")
		log.Printf("Raw event data:\n%s", prettyJSON)
		return nil
	}

	log.Printf("Client ID: %s", client.ID)
	log.Printf("Client Name: %s", client.Name)

	// Process the new client...
	return nil
}

// Package main demonstrates webhook handling.
//
// This example shows how to:
// - Set up a webhook endpoint
// - Protect it with a shared secret header
// - Register event handlers
// - Handle different webhook events
//
// In Invoice Ninja (Settings > Account Management > Integrations > API Webhooks), create one
// webhook per event. Name the event in the target URL, for example
// https://your-server.com/webhook?event=payment.created, and add an X-Webhook-Secret header
// with the value of INVOICE_NINJA_WEBHOOK_SECRET.
//
// Run with: go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

func main() {
	webhookSecret := os.Getenv("INVOICE_NINJA_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Println("Warning: INVOICE_NINJA_WEBHOOK_SECRET not set, requests will not be authenticated")
	}

	// Create a webhook handler. With a secret, requests must send it in the X-Webhook-Secret header.
	webhookHandler := invoiceninja.NewWebhookHandler(webhookSecret)

	// Register handlers for different event types using convenience methods
	webhookHandler.OnPaymentCreated(handlePaymentCreated)
	webhookHandler.OnInvoiceCreated(handleInvoiceCreated)
	webhookHandler.OnClientCreated(handleClientCreated)

	// You can also use the generic On method with any event name you put in the target URL
	webhookHandler.On("invoice.updated", func(event *invoiceninja.WebhookEvent) error {
		log.Printf("Invoice updated event received")
		logEventData(event)
		return nil
	})

	// WebhookHandler implements http.Handler
	mux := http.NewServeMux()
	mux.Handle("/webhook", webhookHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	fmt.Printf("Starting webhook server on port %s...\n", port)
	fmt.Println("Send webhooks to: http://localhost:" + port + "/webhook?event=payment.created")
	log.Fatal(server.ListenAndServe())
}

func handlePaymentCreated(event *invoiceninja.WebhookEvent) error {
	log.Printf("Payment created event received")

	// Parse the payment data using the helper method
	payment, err := event.ParsePayment()
	if err != nil {
		log.Printf("Could not parse payment data: %v", err)
		// Still log the raw data for debugging
		logEventData(event)
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
		logEventData(event)
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
		logEventData(event)
		return nil
	}

	log.Printf("Client ID: %s", client.ID)
	log.Printf("Client Name: %s", client.Name)

	// Process the new client...
	return nil
}

// logEventData logs the raw event data for debugging.
func logEventData(event *invoiceninja.WebhookEvent) {
	prettyJSON, err := json.MarshalIndent(event.Data, "", "  ")
	if err != nil {
		log.Printf("Event data: %s", event.Data)
		return
	}
	log.Printf("Event data:\n%s", prettyJSON)
}

// Package main demonstrates webhook handling.
//
// This example shows how to:
// - Set up a webhook endpoint
// - Verify webhook signatures
// - Handle different webhook events
//
// Run with: go run main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

func main() {
	webhookSecret := os.Getenv("INVOICE_NINJA_WEBHOOK_SECRET")
	if webhookSecret == "" {
		log.Fatal("INVOICE_NINJA_WEBHOOK_SECRET environment variable is required")
	}

	// Create a webhook handler
	webhookHandler := invoiceninja.NewWebhookHandler(webhookSecret)

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request", http.StatusBadRequest)
			return
		}

		// Get the signature from the header
		signature := r.Header.Get("X-Ninja-Signature")
		if signature == "" {
			log.Println("Missing webhook signature")
			http.Error(w, "Missing signature", http.StatusUnauthorized)
			return
		}

		// Verify the webhook signature
		if !webhookHandler.VerifySignature(body, signature) {
			log.Println("Invalid webhook signature")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		// Parse the webhook event
		event, err := webhookHandler.ParseEvent(body)
		if err != nil {
			log.Printf("Error parsing webhook event: %v", err)
			http.Error(w, "Error parsing event", http.StatusBadRequest)
			return
		}

		// Handle different event types
		switch event.EventType {
		case invoiceninja.WebhookEventPaymentCreated:
			handlePaymentCreated(event)

		case invoiceninja.WebhookEventInvoicePaid:
			handleInvoicePaid(event)

		case invoiceninja.WebhookEventClientCreated:
			handleClientCreated(event)

		default:
			log.Printf("Received unhandled webhook event: %s", event.EventType)
		}

		// Respond with success
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Webhook received")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting webhook server on port %s...\n", port)
	fmt.Println("Send webhooks to: http://localhost:" + port + "/webhook")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handlePaymentCreated(event *invoiceninja.WebhookEvent) {
	log.Printf("Payment created event received")

	// Extract payment data from the event
	paymentData, ok := event.Data["payment"].(map[string]interface{})
	if !ok {
		log.Println("Could not parse payment data")
		return
	}

	prettyJSON, _ := json.MarshalIndent(paymentData, "", "  ")
	log.Printf("Payment data:\n%s", prettyJSON)

	// Process the payment...
	// e.g., update your database, send notifications, etc.
}

func handleInvoicePaid(event *invoiceninja.WebhookEvent) {
	log.Printf("Invoice paid event received")

	invoiceData, ok := event.Data["invoice"].(map[string]interface{})
	if !ok {
		log.Println("Could not parse invoice data")
		return
	}

	prettyJSON, _ := json.MarshalIndent(invoiceData, "", "  ")
	log.Printf("Invoice data:\n%s", prettyJSON)

	// Process the paid invoice...
}

func handleClientCreated(event *invoiceninja.WebhookEvent) {
	log.Printf("Client created event received")

	clientData, ok := event.Data["client"].(map[string]interface{})
	if !ok {
		log.Println("Could not parse client data")
		return
	}

	prettyJSON, _ := json.MarshalIndent(clientData, "", "  ")
	log.Printf("Client data:\n%s", prettyJSON)

	// Process the new client...
}

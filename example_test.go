// Example demonstrates basic usage of the Invoice Ninja SDK.
//
// To run this example:
//
//	go run example_test.go
package invoiceninja_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

func Example_basicUsage() {
	// Create a new client with your API token
	client := invoiceninja.NewClient("your-api-token")

	// For self-hosted instances, specify your base URL:
	// client := invoiceninja.NewClient("your-api-token",
	//     invoiceninja.WithBaseURL("https://your-instance.com"))

	ctx := context.Background()

	// List payments with pagination
	payments, err := client.Payments.List(ctx, &invoiceninja.PaymentListOptions{
		PerPage: 10,
		Page:    1,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d payments\n", len(payments.Data))
	for _, p := range payments.Data {
		fmt.Printf("  - %s: $%.2f\n", p.Number, p.Amount)
	}
}

func Example_createPayment() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Create a payment for an invoice
	payment, err := client.Payments.Create(ctx, &invoiceninja.PaymentRequest{
		ClientID: "client-hashed-id",
		Amount:   250.00,
		Date:     "2024-01-15",
		Invoices: []invoiceninja.PaymentInvoice{
			{
				InvoiceID: "invoice-hashed-id",
				Amount:    250.00,
			},
		},
		TransactionRef: "TXN-12345",
		PrivateNotes:   "Payment received via bank transfer",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created payment: %s\n", payment.ID)
}

func Example_createInvoice() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Create an invoice with line items
	invoice, err := client.Invoices.Create(ctx, &invoiceninja.Invoice{
		ClientID: "client-hashed-id",
		Date:     "2024-01-15",
		DueDate:  "2024-02-15",
		LineItems: []invoiceninja.LineItem{
			{
				ProductKey: "Consulting Services",
				Notes:      "January 2024 consulting work",
				Quantity:   10,
				Cost:       150.00,
			},
			{
				ProductKey: "Support Hours",
				Notes:      "Technical support",
				Quantity:   5,
				Cost:       75.00,
			},
		},
		PublicNotes: "Thank you for your business!",
		Terms:       "Net 30",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created invoice: %s (Amount: $%.2f)\n", invoice.Number, invoice.Amount)
}

func Example_createClient() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Create a new client with contact
	newClient, err := client.Clients.Create(ctx, &invoiceninja.INClient{
		Name:       "Acme Corporation",
		Website:    "https://acme.com",
		Phone:      "+1 555-1234",
		Address1:   "123 Main Street",
		City:       "San Francisco",
		State:      "CA",
		PostalCode: "94102",
		CountryID:  "840", // USA
		Contacts: []invoiceninja.ClientContact{
			{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@acme.com",
				Phone:     "+1 555-5678",
				IsPrimary: true,
			},
			{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane.smith@acme.com",
				IsPrimary: false,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Created client: %s (ID: %s)\n", newClient.Name, newClient.ID)
}

func Example_refundPayment() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Process a partial refund
	payment, err := client.Payments.Refund(ctx, &invoiceninja.RefundRequest{
		ID:            "payment-hashed-id",
		Amount:        50.00,
		GatewayRefund: true, // Process refund through payment gateway
		SendEmail:     true, // Send refund notification email
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Refunded $%.2f from payment %s\n", payment.Refunded, payment.Number)
}

func Example_errorHandling() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	payment, err := client.Payments.Get(ctx, "non-existent-id")
	if err != nil {
		if apiErr, ok := invoiceninja.IsAPIError(err); ok {
			switch {
			case apiErr.IsNotFound():
				fmt.Println("Payment not found")
			case apiErr.IsUnauthorized():
				fmt.Println("Invalid or expired API token")
			case apiErr.IsForbidden():
				fmt.Println("You don't have permission to access this resource")
			case apiErr.IsValidationError():
				fmt.Println("Validation errors:")
				for field, errors := range apiErr.Errors {
					fmt.Printf("  %s: %v\n", field, errors)
				}
			case apiErr.IsRateLimited():
				fmt.Println("Rate limit exceeded, please wait before retrying")
			case apiErr.IsServerError():
				fmt.Println("Server error, please try again later")
			default:
				fmt.Printf("API error (status %d): %s\n", apiErr.StatusCode, apiErr.Message)
			}
		} else {
			fmt.Printf("Network or other error: %v\n", err)
		}
		return
	}

	fmt.Printf("Payment: %s\n", payment.Number)
}

func Example_genericRequest() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Access any endpoint not covered by specialized methods
	// Example: Get activities
	var activities json.RawMessage
	err := client.Request(ctx, "GET", "/api/v1/activities", nil, &activities)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Activities response: %s\n", string(activities))

	// Example: With query parameters
	query := url.Values{}
	query.Set("per_page", "50")
	query.Set("page", "1")

	var products json.RawMessage
	err = client.RequestWithQuery(ctx, "GET", "/api/v1/products", query, nil, &products)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Products response: %s\n", string(products))
}

func Example_bulkOperations() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Archive multiple payments at once
	paymentIDs := []string{"payment-id-1", "payment-id-2", "payment-id-3"}
	archivedPayments, err := client.Payments.Bulk(ctx, "archive", paymentIDs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Archived %d payments\n", len(archivedPayments))

	// Mark multiple invoices as sent
	invoiceIDs := []string{"invoice-id-1", "invoice-id-2"}
	sentInvoices, err := client.Invoices.Bulk(ctx, "mark_sent", invoiceIDs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Marked %d invoices as sent\n", len(sentInvoices))
}

func Example_pagination() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Iterate through all pages of clients
	page := 1
	perPage := 20

	for {
		clients, err := client.Clients.List(ctx, &invoiceninja.ClientListOptions{
			PerPage: perPage,
			Page:    page,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Page %d: %d clients\n", page, len(clients.Data))

		for _, c := range clients.Data {
			fmt.Printf("  - %s (Balance: $%.2f)\n", c.Name, c.Balance)
		}

		// Check if there are more pages
		if page >= clients.Meta.Pagination.TotalPages {
			break
		}
		page++
	}
}

func Example_filtering() {
	client := invoiceninja.NewClient("your-api-token")
	ctx := context.Background()

	// Filter payments by client and date range
	payments, err := client.Payments.List(ctx, &invoiceninja.PaymentListOptions{
		ClientID:  "client-hashed-id",
		CreatedAt: "2024-01-01",
		Status:    "active",
		Sort:      "amount|desc",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d payments for client\n", len(payments.Data))

	// Filter clients by balance
	clients, err := client.Clients.List(ctx, &invoiceninja.ClientListOptions{
		Balance: "gt:1000", // Balance greater than $1000
		Sort:    "balance|desc",
		Include: "contacts", // Include contact information
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d clients with balance > $1000\n", len(clients.Data))
}

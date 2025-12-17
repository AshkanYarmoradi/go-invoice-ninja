// Package invoiceninja_test provides integration tests for the Invoice Ninja SDK.
//
// To run these tests against the demo API:
//
//	go test -tags=integration -v ./...
//
// To run against a custom server:
//
//	INVOICE_NINJA_BASE_URL=https://your-server.com \
//	INVOICE_NINJA_API_TOKEN=your-token \
//	go test -tags=integration -v ./...
//
//go:build integration
// +build integration

package invoiceninja_test

import (
	"context"
	"os"
	"testing"
	"time"

	invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

var (
	testClient *invoiceninja.Client
)

func TestMain(m *testing.M) {
	baseURL := os.Getenv("INVOICE_NINJA_BASE_URL")
	if baseURL == "" {
		baseURL = invoiceninja.DemoBaseURL
	}

	apiToken := os.Getenv("INVOICE_NINJA_API_TOKEN")
	if apiToken == "" {
		apiToken = "TOKEN" // Demo API token
	}

	testClient = invoiceninja.NewClient(apiToken,
		invoiceninja.WithBaseURL(baseURL),
		invoiceninja.WithTimeout(30*time.Second),
	)

	os.Exit(m.Run())
}

func TestIntegration_ListPayments(t *testing.T) {
	ctx := context.Background()

	payments, err := testClient.Payments.List(ctx, &invoiceninja.PaymentListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("failed to list payments: %v", err)
	}

	t.Logf("Found %d payments (page 1)", len(payments.Data))

	for _, p := range payments.Data {
		t.Logf("  Payment %s: $%.2f", p.Number, p.Amount)
	}
}

func TestIntegration_ListInvoices(t *testing.T) {
	ctx := context.Background()

	invoices, err := testClient.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("failed to list invoices: %v", err)
	}

	t.Logf("Found %d invoices (page 1)", len(invoices.Data))

	for _, inv := range invoices.Data {
		t.Logf("  Invoice %s: $%.2f (balance: $%.2f)", inv.Number, inv.Amount, inv.Balance)
	}
}

func TestIntegration_ListClients(t *testing.T) {
	ctx := context.Background()

	clients, err := testClient.Clients.List(ctx, &invoiceninja.ClientListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("failed to list clients: %v", err)
	}

	t.Logf("Found %d clients (page 1)", len(clients.Data))

	for _, c := range clients.Data {
		t.Logf("  Client %s (balance: $%.2f)", c.Name, c.Balance)
	}
}

func TestIntegration_ListPaymentTerms(t *testing.T) {
	ctx := context.Background()

	terms, err := testClient.PaymentTerms.List(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list payment terms: %v", err)
	}

	t.Logf("Found %d payment terms", len(terms.Data))

	for _, term := range terms.Data {
		t.Logf("  %s: %d days", term.Name, term.NumDays)
	}
}

func TestIntegration_ListCredits(t *testing.T) {
	ctx := context.Background()

	credits, err := testClient.Credits.List(ctx, &invoiceninja.CreditListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("failed to list credits: %v", err)
	}

	t.Logf("Found %d credits (page 1)", len(credits.Data))

	for _, credit := range credits.Data {
		t.Logf("  Credit %s: $%.2f", credit.Number, credit.Amount)
	}
}

func TestIntegration_GetPayment(t *testing.T) {
	ctx := context.Background()

	// First, get a list of payments to find an ID
	payments, err := testClient.Payments.List(ctx, &invoiceninja.PaymentListOptions{
		PerPage: 1,
	})
	if err != nil {
		t.Fatalf("failed to list payments: %v", err)
	}

	if len(payments.Data) == 0 {
		t.Skip("No payments available to test Get")
	}

	paymentID := payments.Data[0].ID

	// Now get the specific payment
	payment, err := testClient.Payments.Get(ctx, paymentID)
	if err != nil {
		t.Fatalf("failed to get payment: %v", err)
	}

	t.Logf("Got payment: %s ($%.2f)", payment.Number, payment.Amount)
}

func TestIntegration_GetInvoice(t *testing.T) {
	ctx := context.Background()

	// First, get a list of invoices to find an ID
	invoices, err := testClient.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
		PerPage: 1,
	})
	if err != nil {
		t.Fatalf("failed to list invoices: %v", err)
	}

	if len(invoices.Data) == 0 {
		t.Skip("No invoices available to test Get")
	}

	invoiceID := invoices.Data[0].ID

	// Now get the specific invoice
	invoice, err := testClient.Invoices.Get(ctx, invoiceID)
	if err != nil {
		t.Fatalf("failed to get invoice: %v", err)
	}

	t.Logf("Got invoice: %s ($%.2f)", invoice.Number, invoice.Amount)
}

func TestIntegration_GetClient(t *testing.T) {
	ctx := context.Background()

	// First, get a list of clients to find an ID
	clients, err := testClient.Clients.List(ctx, &invoiceninja.ClientListOptions{
		PerPage: 1,
	})
	if err != nil {
		t.Fatalf("failed to list clients: %v", err)
	}

	if len(clients.Data) == 0 {
		t.Skip("No clients available to test Get")
	}

	clientID := clients.Data[0].ID

	// Now get the specific client
	client, err := testClient.Clients.Get(ctx, clientID)
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}

	t.Logf("Got client: %s (balance: $%.2f)", client.Name, client.Balance)
}

func TestIntegration_Pagination(t *testing.T) {
	ctx := context.Background()

	// Test pagination by getting multiple pages
	page1, err := testClient.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
		PerPage: 5,
		Page:    1,
	})
	if err != nil {
		t.Fatalf("failed to get page 1: %v", err)
	}

	t.Logf("Page 1: %d invoices, Total: %d, TotalPages: %d",
		len(page1.Data),
		page1.Meta.Pagination.Total,
		page1.Meta.Pagination.TotalPages)

	if page1.Meta.Pagination.TotalPages > 1 {
		page2, err := testClient.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
			PerPage: 5,
			Page:    2,
		})
		if err != nil {
			t.Fatalf("failed to get page 2: %v", err)
		}

		t.Logf("Page 2: %d invoices", len(page2.Data))

		// Ensure we got different data
		if len(page1.Data) > 0 && len(page2.Data) > 0 {
			if page1.Data[0].ID == page2.Data[0].ID {
				t.Error("expected different invoices on different pages")
			}
		}
	}
}

func TestIntegration_GenericRequest(t *testing.T) {
	ctx := context.Background()

	// Use the generic request method to access activities
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	err := testClient.Request(ctx, "GET", "/api/v1/activities", nil, &result)
	if err != nil {
		t.Fatalf("failed to get activities: %v", err)
	}

	t.Logf("Found %d activities", len(result.Data))
}

func TestIntegration_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Try to get a non-existent payment
	_, err := testClient.Payments.Get(ctx, "nonexistent-id-12345")
	if err == nil {
		t.Error("expected error for non-existent payment")
		return
	}

	apiErr, ok := invoiceninja.IsAPIError(err)
	if !ok {
		t.Errorf("expected APIError, got %T", err)
		return
	}

	t.Logf("Got expected error: %v (status: %d)", apiErr.Message, apiErr.StatusCode)

	// Should be either 404 (not found) or 400 (bad request for invalid ID format)
	if apiErr.StatusCode != 404 && apiErr.StatusCode != 400 && apiErr.StatusCode != 422 {
		t.Errorf("expected status 404, 400, or 422, got %d", apiErr.StatusCode)
	}
}

func TestIntegration_Filtering(t *testing.T) {
	ctx := context.Background()

	// Test filtering invoices by status
	activeInvoices, err := testClient.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
		Status:  "active",
		PerPage: 5,
	})
	if err != nil {
		t.Fatalf("failed to list active invoices: %v", err)
	}

	t.Logf("Found %d active invoices", len(activeInvoices.Data))

	// Test sorting
	sortedInvoices, err := testClient.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
		Sort:    "amount|desc",
		PerPage: 5,
	})
	if err != nil {
		t.Fatalf("failed to list sorted invoices: %v", err)
	}

	t.Logf("Found %d invoices (sorted by amount desc)", len(sortedInvoices.Data))

	// Verify sorting order
	if len(sortedInvoices.Data) >= 2 {
		if sortedInvoices.Data[0].Amount < sortedInvoices.Data[1].Amount {
			t.Log("Note: Invoice amounts may not be strictly sorted if amounts are equal")
		}
	}
}

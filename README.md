# Go Invoice Ninja SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/AshkanYarmoradi/go-invoice-ninja.svg)](https://pkg.go.dev/github.com/AshkanYarmoradi/go-invoice-ninja)
[![Go Report Card](https://goreportcard.com/badge/github.com/AshkanYarmoradi/go-invoice-ninja)](https://goreportcard.com/report/github.com/AshkanYarmoradi/go-invoice-ninja)
[![CI](https://github.com/AshkanYarmoradi/go-invoice-ninja/actions/workflows/ci.yml/badge.svg)](https://github.com/AshkanYarmoradi/go-invoice-ninja/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/AshkanYarmoradi/go-invoice-ninja/branch/main/graph/badge.svg)](https://codecov.io/gh/AshkanYarmoradi/go-invoice-ninja)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A professional, idiomatic Go SDK for the [Invoice Ninja](https://invoiceninja.com) API. This SDK provides a clean interface for interacting with Invoice Ninja's comprehensive invoicing and payment platform.

## ✨ Features

- 🔐 **Secure Authentication** - Token-based API authentication
- 💳 **Payment Management** - Full CRUD operations with refund support
- 📄 **Invoice Operations** - Create, send, and manage invoices
- 👥 **Client Management** - Client CRUD with merge capabilities
- 💰 **Credits & Payment Terms** - Complete credit and terms management
- 📥 **File Operations** - Download PDFs and upload documents
- 🔔 **Webhook Handling** - Built-in handler with signature verification
- ⚡ **Rate Limiting** - Client-side limiting with automatic retry
- 🔄 **Retry Logic** - Exponential backoff for transient failures
- 🌐 **Self-hosted Support** - Works with cloud and self-hosted instances
- ✅ **Fully Tested** - 90+ tests with comprehensive coverage

## 📦 Installation

```bash
go get github.com/AshkanYarmoradi/go-invoice-ninja
```

## 📖 Documentation

- [Getting Started](docs/getting-started.md)
- [Authentication](docs/authentication.md)
- [Error Handling](docs/error-handling.md)
- [API Reference](docs/api-reference.md)

## 🏗️ Project Structure

```
go-invoice-ninja/
├── .github/workflows/     # CI/CD pipelines
├── docs/                  # Detailed documentation
├── examples/              # Runnable examples
│   ├── basic/            # Basic usage
│   ├── invoices/         # Invoice operations
│   └── webhooks/         # Webhook handling
├── testdata/              # Test fixtures
│
├── client.go             # Main client
├── clients.go            # Clients service
├── credits.go            # Credits service
├── errors.go             # Error types
├── files.go              # File operations
├── invoices.go           # Invoices service
├── models.go             # Data models
├── payments.go           # Payments service
├── payment_terms.go      # Payment terms
├── retry.go              # Retry & rate limiting
├── webhooks.go           # Webhook handling
│
├── CHANGELOG.md          # Version history
├── CONTRIBUTING.md       # Contribution guide
├── LICENSE               # MIT License
├── Makefile              # Build automation
└── README.md             # This file
```

## 🚀 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

func main() {
    // Create a new client
    client := invoiceninja.NewClient("your-api-token")
    
    // For self-hosted instances:
    // client := invoiceninja.NewClient("your-api-token", 
    //     invoiceninja.WithBaseURL("https://your-instance.com"))

    ctx := context.Background()

    // List payments
    payments, err := client.Payments.List(ctx, &invoiceninja.PaymentListOptions{
        PerPage: 10,
        Page:    1,
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, payment := range payments.Data {
        fmt.Printf("Payment %s: $%.2f\n", payment.Number, payment.Amount)
    }
}
```

## 🔑 Authentication

All API requests require an API token. You can obtain your token from:
**Settings > Account Management > Integrations > API tokens**

```go
client := invoiceninja.NewClient("your-api-token")
```

## ⚙️ Configuration Options

```go
// Custom HTTP client
client := invoiceninja.NewClient("token",
    invoiceninja.WithHTTPClient(customHTTPClient))

// Custom base URL (for self-hosted)
client := invoiceninja.NewClient("token",
    invoiceninja.WithBaseURL("https://your-instance.com"))

// Custom timeout
client := invoiceninja.NewClient("token",
    invoiceninja.WithTimeout(60 * time.Second))
```

## 💳 Payments

### List Payments

```go
payments, err := client.Payments.List(ctx, &invoiceninja.PaymentListOptions{
    PerPage:  20,
    Page:     1,
    ClientID: "client-hash-id",
    Status:   "active",
    Sort:     "amount|desc",
})
```

### Get Payment

```go
payment, err := client.Payments.Get(ctx, "payment-hash-id")
```

### Create Payment

```go
payment, err := client.Payments.Create(ctx, &invoiceninja.PaymentRequest{
    ClientID: "client-hash-id",
    Amount:   100.00,
    Date:     "2024-01-15",
    Invoices: []invoiceninja.PaymentInvoice{
        {InvoiceID: "invoice-hash-id", Amount: 100.00},
    },
})
```

### Update Payment

```go
payment, err := client.Payments.Update(ctx, "payment-hash-id", &invoiceninja.PaymentRequest{
    PrivateNotes: "Updated notes",
})
```

### Delete Payment

```go
err := client.Payments.Delete(ctx, "payment-hash-id")
```

### Refund Payment

```go
payment, err := client.Payments.Refund(ctx, &invoiceninja.RefundRequest{
    ID:            "payment-hash-id",
    Amount:        50.00,
    GatewayRefund: true,
})
```

### Bulk Actions

```go
// Archive multiple payments
payments, err := client.Payments.Bulk(ctx, "archive", []string{"id1", "id2"})

// Single item convenience methods
payment, err := client.Payments.Archive(ctx, "payment-hash-id")
payment, err := client.Payments.Restore(ctx, "payment-hash-id")
```

## Invoices

### List Invoices

```go
invoices, err := client.Invoices.List(ctx, &invoiceninja.InvoiceListOptions{
    PerPage:  20,
    ClientID: "client-hash-id",
})
```

### Get Invoice

```go
invoice, err := client.Invoices.Get(ctx, "invoice-hash-id")
```

### Create Invoice

```go
invoice, err := client.Invoices.Create(ctx, &invoiceninja.Invoice{
    ClientID: "client-hash-id",
    LineItems: []invoiceninja.LineItem{
        {ProductKey: "Product A", Quantity: 2, Cost: 50.00},
    },
})
```

### Invoice Actions

```go
// Mark as paid
invoice, err := client.Invoices.MarkPaid(ctx, "invoice-hash-id")

// Mark as sent
invoice, err := client.Invoices.MarkSent(ctx, "invoice-hash-id")

// Send via email
invoice, err := client.Invoices.Email(ctx, "invoice-hash-id")
```

## Clients

### List Clients

```go
clients, err := client.Clients.List(ctx, &invoiceninja.ClientListOptions{
    PerPage: 20,
    Balance: "gt:1000",  // Balance greater than 1000
    Include: "contacts,documents",
})
```

### Create Client

```go
newClient, err := client.Clients.Create(ctx, &invoiceninja.INClient{
    Name: "Acme Corporation",
    Contacts: []invoiceninja.ClientContact{
        {
            FirstName: "John",
            LastName:  "Doe",
            Email:     "john@acme.com",
            IsPrimary: true,
        },
    },
})
```

### Merge Clients

```go
mergedClient, err := client.Clients.Merge(ctx, "primary-id", "mergeable-id")
```

## Payment Terms

```go
// List payment terms
terms, err := client.PaymentTerms.List(ctx, nil)

// Create a payment term
term, err := client.PaymentTerms.Create(ctx, &invoiceninja.PaymentTerm{
    Name:    "Net 45",
    NumDays: 45,
})

// Get, Update, Delete
term, err := client.PaymentTerms.Get(ctx, "term-id")
term, err := client.PaymentTerms.Update(ctx, "term-id", &invoiceninja.PaymentTerm{Name: "Net 60"})
err := client.PaymentTerms.Delete(ctx, "term-id")
```

## Credits

```go
// List credits
credits, err := client.Credits.List(ctx, &invoiceninja.CreditListOptions{
    ClientID: "client-hash-id",
    PerPage:  20,
})

// Create a credit
credit, err := client.Credits.Create(ctx, &invoiceninja.Credit{
    ClientID: "client-hash-id",
    LineItems: []invoiceninja.LineItem{
        {ProductKey: "Credit", Quantity: 1, Cost: 100.00},
    },
})

// Credit actions
credit, err := client.Credits.MarkSent(ctx, "credit-id")
credit, err := client.Credits.Email(ctx, "credit-id")
```

## File Downloads

```go
// Download invoice PDF
pdf, err := client.Downloads.DownloadInvoicePDF(ctx, "invitation-key")

// Download delivery note
pdf, err := client.Downloads.DownloadInvoiceDeliveryNote(ctx, "invoice-id")

// Download credit PDF
pdf, err := client.Downloads.DownloadCreditPDF(ctx, "invitation-key")

// Save to file
os.WriteFile("invoice.pdf", pdf, 0644)
```

## File Uploads

```go
// Upload document to invoice
err := client.Uploads.UploadInvoiceDocument(ctx, "invoice-id", "/path/to/file.pdf")

// Upload to other entities
err := client.Uploads.UploadPaymentDocument(ctx, "payment-id", "/path/to/file.pdf")
err := client.Uploads.UploadClientDocument(ctx, "client-id", "/path/to/file.pdf")
err := client.Uploads.UploadCreditDocument(ctx, "credit-id", "/path/to/file.pdf")

// Upload from io.Reader
reader := bytes.NewReader(pdfContent)
err := client.Uploads.UploadDocumentFromReader(ctx, "invoices", "invoice-id", "document.pdf", reader)
```

## Webhooks

Handle incoming webhooks from Invoice Ninja:

```go
// Create a webhook handler
handler := invoiceninja.NewWebhookHandler("your-webhook-secret")

// Register event handlers
handler.OnPaymentCreated(func(event *invoiceninja.WebhookEvent) error {
    payment, err := event.ParsePayment()
    if err != nil {
        return err
    }
    fmt.Printf("New payment: %s ($%.2f)\n", payment.Number, payment.Amount)
    return nil
})

handler.OnInvoiceCreated(func(event *invoiceninja.WebhookEvent) error {
    invoice, err := event.ParseInvoice()
    if err != nil {
        return err
    }
    fmt.Printf("New invoice: %s\n", invoice.Number)
    return nil
})

// Use as HTTP handler
http.Handle("/webhook", handler)
http.ListenAndServe(":8080", nil)
```

Supported webhook events:
- `OnInvoiceCreated`, `OnInvoiceUpdated`, `OnInvoiceDeleted`
- `OnPaymentCreated`, `OnPaymentUpdated`, `OnPaymentDeleted`
- `OnClientCreated`, `OnClientUpdated`
- `OnCreditCreated`, `OnQuoteCreated`

## Rate Limiting & Retry

For production use, use the rate-limited client with automatic retries:

```go
// Create a rate-limited client
client := invoiceninja.NewRateLimitedClient("your-api-token",
    invoiceninja.WithBaseURL("https://your-instance.com"))

// Configure rate limit (requests per second)
client.SetRateLimit(10)

// Configure retry behavior
client.SetRetryConfig(&invoiceninja.RetryConfig{
    MaxRetries:         3,
    InitialBackoff:     1 * time.Second,
    MaxBackoff:         30 * time.Second,
    BackoffMultiplier:  2.0,
    RetryOnStatusCodes: []int{429, 500, 502, 503, 504},
    Jitter:             true,
})
```

## Generic Requests

For API endpoints not covered by specialized methods, use the generic request:

```go
// GET request
var activities json.RawMessage
err := client.Request(ctx, "GET", "/api/v1/activities", nil, &activities)

// POST request with body
body := map[string]interface{}{
    "name": "New Product",
    "cost": 99.99,
}
var result json.RawMessage
err := client.Request(ctx, "POST", "/api/v1/products", body, &result)

// With query parameters
query := url.Values{}
query.Set("per_page", "50")
err := client.RequestWithQuery(ctx, "GET", "/api/v1/products", query, nil, &result)
```

## Error Handling

The SDK provides typed errors with helper methods:

```go
payment, err := client.Payments.Get(ctx, "invalid-id")
if err != nil {
    if apiErr, ok := invoiceninja.IsAPIError(err); ok {
        if apiErr.IsNotFound() {
            fmt.Println("Payment not found")
        } else if apiErr.IsUnauthorized() {
            fmt.Println("Invalid API token")
        } else if apiErr.IsValidationError() {
            fmt.Printf("Validation errors: %v\n", apiErr.Errors)
        } else if apiErr.IsRateLimited() {
            fmt.Println("Rate limit exceeded, please wait")
        }
    }
    log.Fatal(err)
}
```

## API Reference

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 400 | Bad Request |
| 401 | Unauthorized - Invalid API token |
| 403 | Forbidden - No permission |
| 404 | Not Found |
| 422 | Validation Error |
| 429 | Rate Limited |
| 5xx | Server Error |

## 🧪 Testing

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Run with coverage
make coverage

# Run linter
make lint
```

## 🔗 Integration Tests

Run integration tests against a live Invoice Ninja server:

```bash
# Run against demo server
go test -tags=integration -v ./...

# Run against custom server
INVOICE_NINJA_BASE_URL=https://your-server.com \
INVOICE_NINJA_API_TOKEN=your-token \
go test -tags=integration -v ./...
```

## 📚 Examples

Check out the [examples](examples/) directory for complete working examples:

- [Basic Usage](examples/basic/) - Getting started
- [Invoice Management](examples/invoices/) - Creating and managing invoices
- [Webhook Handling](examples/webhooks/) - Setting up webhook handlers

## 📋 API Reference

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 400 | Bad Request |
| 401 | Unauthorized - Invalid API token |
| 403 | Forbidden - No permission |
| 404 | Not Found |
| 422 | Validation Error |
| 429 | Rate Limited |
| 5xx | Server Error |

## 📄 License

This SDK is released under the [MIT License](LICENSE).

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`make test`)
5. Run the linter (`make lint`)
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📞 Support

- [GitHub Issues](https://github.com/AshkanYarmoradi/go-invoice-ninja/issues)
- [Invoice Ninja Documentation](https://invoiceninja.github.io/)
- [Invoice Ninja API Reference](https://api-docs.invoicing.co/)


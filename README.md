# Go Invoice Ninja SDK

A professional Go SDK for the [Invoice Ninja](https://invoiceninja.com) API. This SDK provides a clean, idiomatic Go interface for interacting with Invoice Ninja's payment-related functionality.

## Features

- **Payment-focused**: Specialized support for payments, invoices, and clients
- **Generic requests**: Access any API endpoint not covered by specialized methods
- **Self-hosted support**: Works with both cloud (invoicing.co) and self-hosted instances
- **Comprehensive error handling**: Typed errors with helper methods
- **Context support**: All operations support Go's context for cancellation and timeouts
- **Fully tested**: Comprehensive test coverage

## Installation

```bash
go get github.com/invoiceninja/go-invoice-ninja/invoiceninja
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/invoiceninja/go-invoice-ninja/invoiceninja"
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

## Authentication

All API requests require an API token. You can obtain your token from:
**Settings > Account Management > Integrations > API tokens**

```go
client := invoiceninja.NewClient("your-api-token")
```

## Configuration Options

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

## Payments

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

## License

This SDK is released under the MIT License.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test -v ./...`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Adding or updating tests
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

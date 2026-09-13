# API Reference

This document provides a detailed reference for all available SDK methods.

## Client

### Creating a Client

```go
client := invoiceninja.NewClient(apiToken string, opts ...ClientOption)

// With rate limiting (10 requests per second) and DefaultRetryConfig() enabled
client := invoiceninja.NewRateLimitedClient(apiToken string, opts ...ClientOption)
```

### Options

| Option | Description |
|--------|-------------|
| `WithBaseURL(url)` | Set custom base URL |
| `WithHTTPClient(client)` | Use custom HTTP client |
| `WithTimeout(duration)` | Set request timeout (a client passed to `WithHTTPClient` is not modified) |
| `WithRateLimiter(limiter)` | Limit requests per second, e.g. `NewRateLimiter(10)` |
| `WithRetryConfig(config)` | Retry failed requests, e.g. `DefaultRetryConfig()` (see [Error Handling](error-handling.md#retry-configuration)) |

---

## Payments Service

### List Payments

```go
payments, err := client.Payments.List(ctx, &PaymentListOptions{
    PerPage:   int,    // Items per page (default: 20)
    Page:      int,    // Page number
    ClientID:  string, // Filter by client
    Status:    string, // Filter by status
    Sort:      string, // Sort field (e.g., "amount|desc")
    CreatedAt: int,    // Filter by created timestamp
    UpdatedAt: int,    // Filter by updated timestamp
    IsDeleted: bool,   // Include deleted
})
```

### Get Payment

```go
payment, err := client.Payments.Get(ctx, paymentID string)
```

### Create Payment

```go
payment, err := client.Payments.Create(ctx, &PaymentRequest{
    ClientID:       string,           // Required
    Amount:         float64,          // Required
    Date:           string,           // Payment date
    TypeID:         string,           // Payment type
    TransactionRef: string,           // Reference number
    PrivateNotes:   string,           // Internal notes
    Invoices:       []PaymentInvoice, // Applied invoices
    Credits:        []PaymentCredit,  // Applied credits
})
```

### Update Payment

```go
payment, err := client.Payments.Update(ctx, paymentID string, &PaymentRequest{...})
```

### Delete Payment

```go
err := client.Payments.Delete(ctx, paymentID string)
```

### Refund Payment

```go
payment, err := client.Payments.Refund(ctx, &RefundRequest{
    ID:       string,  // Payment ID
    Amount:   float64, // Refund amount
    Invoices: []RefundInvoice,
    Date:     string,
})
```

### Bulk Actions

```go
err := client.Payments.Bulk(ctx, &BulkActionRequest{
    Action: string,   // "archive", "restore", "delete"
    IDs:    []string, // Payment IDs
})
```

---

## Invoices Service

### List Invoices

```go
invoices, err := client.Invoices.List(ctx, &InvoiceListOptions{
    PerPage:  int,
    Page:     int,
    ClientID: string,
    Status:   string,
    Sort:     string,
})
```

### Get Invoice

```go
invoice, err := client.Invoices.Get(ctx, invoiceID string)
```

### Create Invoice

```go
invoice, err := client.Invoices.Create(ctx, &Invoice{
    ClientID:    string,     // Required
    Date:        string,     // Invoice date
    DueDate:     string,     // Due date
    LineItems:   []LineItem, // Invoice items
    PublicNotes: string,     // Client-visible notes
    Terms:       string,     // Payment terms
    Footer:      string,     // Footer text
    Discount:    float64,    // Discount amount
    TaxName1:    string,     // Tax name
    TaxRate1:    float64,    // Tax rate
})
```

### Update Invoice

```go
invoice, err := client.Invoices.Update(ctx, invoiceID string, &Invoice{...})
```

### Delete Invoice

```go
err := client.Invoices.Delete(ctx, invoiceID string)
```

### Download PDF

```go
pdfBytes, err := client.Downloads.Invoice(ctx, invitationKey string)
```

### Bulk Actions

```go
err := client.Invoices.Bulk(ctx, &BulkActionRequest{
    Action: string,   // "archive", "restore", "delete", "mark_sent", "mark_paid"
    IDs:    []string,
})
```

---

## Clients Service

### List Clients

```go
clients, err := client.Clients.List(ctx, &ClientListOptions{
    PerPage: int,
    Page:    int,
    Status:  string,
    Sort:    string,
})
```

### Get Client

```go
c, err := client.Clients.Get(ctx, clientID string)
```

### Create Client

```go
c, err := client.Clients.Create(ctx, &Client{
    Name:           string, // Required
    DisplayName:    string,
    Address1:       string,
    Address2:       string,
    City:           string,
    State:          string,
    PostalCode:     string,
    CountryID:      string,
    Phone:          string,
    Website:        string,
    PrivateNotes:   string,
    PublicNotes:    string,
    VATNumber:      string,
    IDNumber:       string,
    Contacts:       []ClientContact,
})
```

### Update Client

```go
c, err := client.Clients.Update(ctx, clientID string, &Client{...})
```

### Delete Client

```go
err := client.Clients.Delete(ctx, clientID string)
```

### Merge Clients

```go
c, err := client.Clients.Merge(ctx, targetClientID, sourceClientID string)
```

---

## Credits Service

### List Credits

```go
credits, err := client.Credits.List(ctx, &CreditListOptions{...})
```

### Get Credit

```go
credit, err := client.Credits.Get(ctx, creditID string)
```

### Create Credit

```go
credit, err := client.Credits.Create(ctx, &Credit{
    ClientID:  string,
    Amount:    float64,
    Date:      string,
    LineItems: []LineItem,
})
```

### Update Credit

```go
credit, err := client.Credits.Update(ctx, creditID string, &Credit{...})
```

### Delete Credit

```go
err := client.Credits.Delete(ctx, creditID string)
```

---

## Payment Terms Service

### List Payment Terms

```go
terms, err := client.PaymentTerms.List(ctx, &PaymentTermListOptions{...})
```

### Get Payment Term

```go
term, err := client.PaymentTerms.Get(ctx, termID string)
```

### Create Payment Term

```go
term, err := client.PaymentTerms.Create(ctx, &PaymentTerm{
    Name:    string, // e.g., "Net 30"
    NumDays: int,    // e.g., 30
})
```

### Update Payment Term

```go
term, err := client.PaymentTerms.Update(ctx, termID string, &PaymentTerm{...})
```

### Delete Payment Term

```go
err := client.PaymentTerms.Delete(ctx, termID string)
```

---

## Webhooks

Webhooks are created in Invoice Ninja under Settings > Account Management > Integrations > API Webhooks, one webhook per event. The SDK doesn't manage those webhooks; it provides a handler for the requests Invoice Ninja sends.

Invoice Ninja sends only the entity (for example the payment) as JSON, without the event name or a signature. For each webhook:
- Put the event name in the target URL, e.g. `https://example.com/webhook?event=payment.created`, or add an `X-Webhook-Event` header
- Add an `X-Webhook-Secret` header with the secret you pass to `NewWebhookHandler`

### Webhook Handler

```go
handler := invoiceninja.NewWebhookHandler(secret string)

// Register a handler for any event name
handler.On(eventType string, func(event *invoiceninja.WebhookEvent) error { ... })

// Or use the helpers: OnInvoiceCreated, OnInvoiceUpdated, OnInvoiceDeleted,
// OnPaymentCreated, OnPaymentUpdated, OnPaymentDeleted, OnClientCreated,
// OnClientUpdated, OnCreditCreated, OnQuoteCreated
handler.OnPaymentCreated(func(event *invoiceninja.WebhookEvent) error { ... })

// WebhookHandler implements http.Handler
http.Handle("/webhook", handler)
```

How requests are handled:
- Only `POST` and `PUT` are accepted; other methods get `405`.
- If a secret is set, the `X-Webhook-Secret` header must match it. For custom senders, a hex-encoded HMAC-SHA256 signature of the body in `X-Ninja-Signature` is accepted instead. Otherwise the response is `401`.
- The event name is read from the `X-Webhook-Event` header, then the `event` query parameter. A JSON body of the form `{"event_type": "...", "data": {...}}` is also accepted. Without an event name the response is `400`.
- Bodies larger than `MaxWebhookBodyBytes` (10 MB) get `413`.
- Events without a registered handler get `200`. If a handler returns an error, the response is `500` without the error details.

### Parsing Event Data

```go
invoice, err := event.ParseInvoice() // *Invoice
payment, err := event.ParsePayment() // *Payment
client, err := event.ParseClient()   // *INClient
credit, err := event.ParseCredit()   // *Credit
```

---

## Generic Requests

For endpoints not covered by specialized methods:

```go
var result json.RawMessage
err := client.Request(ctx, method, path string, body, result interface{})
```

Example:

```go
var activities []map[string]interface{}
err := client.Request(ctx, "GET", "/api/v1/activities", nil, &activities)
```

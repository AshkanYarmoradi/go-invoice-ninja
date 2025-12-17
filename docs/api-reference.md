# API Reference

This document provides a detailed reference for all available SDK methods.

## Client

### Creating a Client

```go
client := invoiceninja.NewClient(apiToken string, opts ...Option)
```

### Options

| Option | Description |
|--------|-------------|
| `WithBaseURL(url)` | Set custom base URL |
| `WithHTTPClient(client)` | Use custom HTTP client |
| `WithTimeout(duration)` | Set request timeout |
| `WithRateLimiter(limiter)` | Enable rate limiting |
| `WithRetryConfig(config)` | Configure retry behavior |

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

## Webhooks Service

### List Webhooks

```go
webhooks, err := client.Webhooks.List(ctx, &WebhookListOptions{...})
```

### Create Webhook

```go
webhook, err := client.Webhooks.Create(ctx, &Webhook{
    TargetURL:  string, // Your webhook endpoint
    EventID:    string, // Event to subscribe to
    Format:     string, // "JSON"
})
```

### Delete Webhook

```go
err := client.Webhooks.Delete(ctx, webhookID string)
```

### Webhook Handler

```go
handler := invoiceninja.NewWebhookHandler(secret string)

// Verify signature
valid := handler.VerifySignature(body []byte, signature string)

// Parse event
event, err := handler.ParseEvent(body []byte)
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

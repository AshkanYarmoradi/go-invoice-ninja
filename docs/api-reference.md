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
    Filter:    string, // Search across amount, date and custom values
    Number:    string, // Filter by payment number
    ClientID:  string, // Filter by client
    Status:    string, // Filter by status
    CreatedAt: string, // Filter by creation date
    UpdatedAt: string, // Filter by update date
    IsDeleted: *bool,  // Filter by deleted status
    VendorID:  string, // Filter by vendor
    Sort:      string, // Sort order (e.g., "id|desc", "number|asc")
    Include:   string, // Related entities to include
})
```

Zero-value fields, including a nil `IsDeleted`, are not sent. Pass `nil` instead of an options struct to use the API defaults. `List` returns `*ListResponse[Payment]`, with the payments in `Data` and pagination details in `Meta.Pagination`. The `List` methods of the other services work the same way.

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

### Create Payment with Email Receipt

```go
payment, err := client.Payments.CreateWithEmailReceipt(ctx, &PaymentRequest{...}, sendEmail bool)
```

Same as `Create`, but also sets the `email_receipt` query parameter to `sendEmail`.

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
    ID:            string,           // Payment ID
    Amount:        float64,          // Refund amount
    Invoices:      []PaymentInvoice, // Invoices to refund
    Date:          string,           // Refund date
    GatewayRefund: bool,             // Refund through the payment gateway
    SendEmail:     bool,             // Send a refund notification email
})
```

### Archive and Restore

```go
payment, err := client.Payments.Archive(ctx, paymentID string)
payment, err := client.Payments.Restore(ctx, paymentID string)
```

Each sends a single-ID bulk action and returns the first payment in the response, or an error if the response contains none.

### Bulk Actions

```go
payments, err := client.Payments.Bulk(ctx, action string, ids []string) // []Payment
```

`action` is sent as-is, for example `"archive"`, `"restore"` or `"delete"`.

### Get Blank Payment

```go
payment, err := client.Payments.GetBlank(ctx)
```

Returns a new payment populated with the server's default values.

---

## Invoices Service

### List Invoices

```go
invoices, err := client.Invoices.List(ctx, &InvoiceListOptions{
    PerPage:   int,    // Items per page (default: 20)
    Page:      int,    // Page number
    Filter:    string, // Search across multiple fields
    ClientID:  string, // Filter by client
    Status:    string, // Filter by status
    CreatedAt: string, // Filter by creation date
    UpdatedAt: string, // Filter by update date
    IsDeleted: *bool,  // Filter by deleted status
    Sort:      string, // Sort order (e.g., "id|desc", "number|asc")
    Include:   string, // Related entities to include
})
```

### Get Invoice

```go
invoice, err := client.Invoices.Get(ctx, invoiceID string)
```

### Create Invoice

```go
invoice, err := client.Invoices.Create(ctx, &Invoice{
    ClientID:           string,     // Required
    Date:               string,     // Invoice date
    DueDate:            string,     // Due date
    LineItems:          []LineItem, // Invoice items
    PublicNotes:        string,     // Client-visible notes
    Terms:              string,     // Payment terms
    Footer:             string,     // Footer text
    Discount:           float64,    // Discount amount
    TaxName1:           string,     // Tax name
    TaxRate1:           float64,    // Tax rate
    UsesInclusiveTaxes: bool,       // Line item prices already include taxes
})
```

`UsesInclusiveTaxes` is left out of the request when it is `false` (like other `bool` fields), so an update can't switch an invoice back to exclusive taxes.

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
pdfBytes, err := client.Invoices.Download(ctx, invitationKey string)
```

`Download` calls `client.Downloads.DownloadInvoicePDF(ctx, invitationKey)`. It takes an invitation key, not an invoice ID. See [Downloads Service](#downloads-service).

### Invoice Actions

```go
invoice, err := client.Invoices.Archive(ctx, invoiceID string)
invoice, err := client.Invoices.Restore(ctx, invoiceID string)
invoice, err := client.Invoices.MarkPaid(ctx, invoiceID string)
invoice, err := client.Invoices.MarkSent(ctx, invoiceID string)
invoice, err := client.Invoices.Email(ctx, invoiceID string)
```

Each sends a single-ID bulk action (`archive`, `restore`, `mark_paid`, `mark_sent` or `email`) and returns the first invoice in the response, or an error if the response contains none.

### Bulk Actions

```go
invoices, err := client.Invoices.Bulk(ctx, action string, ids []string) // []Invoice
```

`action` is sent as-is, for example `"archive"`, `"restore"`, `"delete"`, `"mark_sent"`, `"mark_paid"` or `"email"`.

### Get Blank Invoice

```go
invoice, err := client.Invoices.GetBlank(ctx)
```

---

## Clients Service

### List Clients

```go
clients, err := client.Clients.List(ctx, &ClientListOptions{
    PerPage:   int,    // Items per page (default: 20)
    Page:      int,    // Page number
    Filter:    string, // Search across multiple fields
    Balance:   string, // Filter by balance (e.g., "gt:1000", "lt:500")
    Status:    string, // Filter by status
    CreatedAt: string, // Filter by creation date
    UpdatedAt: string, // Filter by update date
    IsDeleted: *bool,  // Filter by deleted status
    Sort:      string, // Sort order (e.g., "name|desc", "balance|asc")
    Include:   string, // Related entities (contacts, documents, activities)
})
```

### Get Client

```go
c, err := client.Clients.Get(ctx, clientID string)
```

### Create Client

```go
c, err := client.Clients.Create(ctx, &INClient{
    Name:           string, // Required
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
    VatNumber:      string,
    IDNumber:       string,
    Contacts:       []ClientContact,
})
```

A `ClientContact` has `ID`, `FirstName`, `LastName`, `Email`, `Phone`, `IsPrimary`, `ContactKey` and `CustomValue1` to `CustomValue4`.

### Update Client

```go
c, err := client.Clients.Update(ctx, clientID string, &INClient{...})
```

### Delete Client

```go
err := client.Clients.Delete(ctx, clientID string)
```

`Delete` is a soft delete. To permanently remove a client and all of their records, use `Purge`:

```go
err := client.Clients.Purge(ctx, clientID string)
```

### Archive and Restore

```go
c, err := client.Clients.Archive(ctx, clientID string)
c, err := client.Clients.Restore(ctx, clientID string)
```

### Merge Clients

```go
c, err := client.Clients.Merge(ctx, primaryID, mergeableID string)
```

Sends `POST /api/v1/clients/{primaryID}/{mergeableID}/merge` and returns the client from the response.

### Bulk Actions

```go
clients, err := client.Clients.Bulk(ctx, action string, ids []string) // []INClient
```

### Get Blank Client

```go
c, err := client.Clients.GetBlank(ctx)
```

### Client Portal Links

```go
resp, err := client.Clients.SwitchToClientPortal(ctx, clientID, contactID string) // *ClientPortalSwitchResponse
portalURL := resp.URL

// Same as SwitchToClientPortal(ctx, clientID, ""):
resp, err := client.Clients.GetClientPortalURL(ctx, clientID string)
```

`SwitchToClientPortal` does not call a portal endpoint. It fetches the client with `Get`, picks a contact, and builds the URL `{baseURL}/client/key_login/{contact_key}` from the client's base URL and the contact's `ContactKey`:

- If `contactID` is set, it uses the contact with that ID.
- If `contactID` is empty, it uses the primary contact, falling back to the first contact.

It returns an error when:

- The client can't be fetched. This error wraps the underlying error with `%w`, so check it with `invoiceninja.IsAPIError(err)` rather than a type assertion.
- The client has no contacts.
- `contactID` is set and no contact has that ID.
- The chosen contact has no `contact_key` (portal access may be disabled).

Anyone with the URL can open the client portal without a password, so treat it like a credential.

### Statements

`GetStatement(ctx, req *StatementRequest) ([]byte, error)` is not implemented and always returns an error.

---

## Credits Service

### List Credits

```go
credits, err := client.Credits.List(ctx, &CreditListOptions{
    PerPage:   int,    // Items per page
    Page:      int,    // Page number
    Filter:    string, // Search filter
    ClientID:  string, // Filter by client
    Status:    string, // Filter by status
    CreatedAt: string, // Filter by creation date
    UpdatedAt: string, // Filter by update date
    IsDeleted: *bool,  // Filter by deleted status
    Sort:      string, // Sort order
    Include:   string, // Related entities to include
})
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

The `Credit` type has these fields:

| Type | Fields |
|------|--------|
| `string` | `ID`, `UserID`, `AssignedUserID`, `ClientID`, `StatusID`, `InvoiceID`, `Number`, `PONumber`, `Terms`, `PublicNotes`, `PrivateNotes`, `Footer`, `CustomValue1`, `CustomValue2`, `CustomValue3`, `CustomValue4`, `TaxName1`, `TaxName2`, `TaxName3`, `Date`, `LastSentDate`, `NextSendDate`, `PartialDueDate`, `DueDate` |
| `float64` | `TaxRate1`, `TaxRate2`, `TaxRate3`, `TotalTaxes`, `Amount`, `Balance`, `PaidToDate`, `Discount`, `Partial` |
| `bool` | `IsAmountDiscount`, `IsDeleted`, `UsesInclusiveTaxes` |
| `[]LineItem` | `LineItems` |
| `int64` | `UpdatedAt`, `ArchivedAt`, `CreatedAt` |

### Update Credit

```go
credit, err := client.Credits.Update(ctx, creditID string, &Credit{...})
```

### Delete Credit

```go
err := client.Credits.Delete(ctx, creditID string)
```

### Credit Actions

```go
credit, err := client.Credits.Archive(ctx, creditID string)
credit, err := client.Credits.Restore(ctx, creditID string)
credit, err := client.Credits.MarkSent(ctx, creditID string)
credit, err := client.Credits.Email(ctx, creditID string)
```

Each sends a single-ID bulk action (`archive`, `restore`, `mark_sent` or `email`) and returns the first credit in the response, or an error if the response contains none.

### Bulk Actions

```go
credits, err := client.Credits.Bulk(ctx, action string, ids []string) // []Credit
```

### Get Blank Credit

```go
credit, err := client.Credits.GetBlank(ctx)
```

---

## Downloads Service

Each method reads the whole response into memory and returns the raw bytes. Requests are sent with `Accept: application/pdf`, and a response with status 400 or above is returned as an `*APIError`.

| Method | Request |
|--------|---------|
| `DownloadInvoicePDF(ctx, invitationKey string) ([]byte, error)` | `GET /api/v1/invoice/{invitationKey}/download` |
| `DownloadInvoiceDeliveryNote(ctx, invoiceID string) ([]byte, error)` | `GET /api/v1/invoices/{invoiceID}/delivery_note` |
| `DownloadCreditPDF(ctx, invitationKey string) ([]byte, error)` | `GET /api/v1/credit/{invitationKey}/download` |
| `DownloadQuotePDF(ctx, invitationKey string) ([]byte, error)` | `GET /api/v1/quote/{invitationKey}/download` |

```go
pdfBytes, err := client.Downloads.DownloadInvoicePDF(ctx, invitationKey)
if err != nil {
    return err
}
if err := os.WriteFile("invoice.pdf", pdfBytes, 0o600); err != nil {
    return err
}
```

---

## Uploads Service

Each method sends the file as multipart form data in a `POST` request, with the file in the `documents[]` field and `_method=PUT`, and returns only an error. The file content is buffered in memory before the request is sent, and a response with status 400 or above is returned as an `*APIError`.

| Method | Request |
|--------|---------|
| `UploadDocument(ctx, entityType, entityID, filePath string) error` | `POST /api/v1/{entityType}/{entityID}/upload` |
| `UploadInvoiceDocument(ctx, invoiceID, filePath string) error` | `POST /api/v1/invoices/{invoiceID}/upload` |
| `UploadPaymentDocument(ctx, paymentID, filePath string) error` | `POST /api/v1/payments/{paymentID}/upload` |
| `UploadClientDocument(ctx, clientID, filePath string) error` | `POST /api/v1/clients/{clientID}/upload` |
| `UploadCreditDocument(ctx, creditID, filePath string) error` | `POST /api/v1/credits/{creditID}/upload` |
| `UploadDocumentFromReader(ctx, entityType, entityID, filename string, reader io.Reader) error` | `POST /api/v1/{entityType}/{entityID}/upload` |

`entityType` is the plural path segment, such as `"invoices"` or `"clients"`. The methods that take `filePath` open the file and use its base name as the upload filename.

```go
err := client.Uploads.UploadInvoiceDocument(ctx, invoiceID, "receipt.pdf")

// Upload from memory or any other io.Reader:
err = client.Uploads.UploadDocumentFromReader(ctx, "invoices", invoiceID, "receipt.pdf", bytes.NewReader(data))
```

---

## Payment Terms Service

### List Payment Terms

```go
terms, err := client.PaymentTerms.List(ctx, &PaymentTermListOptions{
    PerPage: int,    // Items per page
    Page:    int,    // Page number
    Include: string, // Related entities to include
})
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

### Archive and Restore

```go
term, err := client.PaymentTerms.Archive(ctx, termID string)
term, err := client.PaymentTerms.Restore(ctx, termID string)
```

### Bulk Actions

```go
terms, err := client.PaymentTerms.Bulk(ctx, action string, ids []string) // []PaymentTerm
```

### Get Blank Payment Term

```go
term, err := client.PaymentTerms.GetBlank(ctx)
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
Request(ctx, method, path string, body, result interface{}) error
RequestWithQuery(ctx, method, path string, query url.Values, body, result interface{}) error
```

`path` is appended to the base URL, so API paths must include the `/api/v1` prefix. A non-nil `body` is sent as JSON, a non-empty response body is decoded into `result` when `result` is not `nil`, and a response with status 400 or above is returned as an `*APIError`.

Example:

```go
var activities struct {
    Data []map[string]interface{} `json:"data"`
}
err := client.Request(ctx, "GET", "/api/v1/activities", nil, &activities)

query := url.Values{}
query.Set("per_page", "50")

var products json.RawMessage
err = client.RequestWithQuery(ctx, "GET", "/api/v1/products", query, nil, &products)
```

### Base URL

```go
client.SetBaseURL(baseURL string)
```

Changes the base URL of an existing client, for example to point it at a self-hosted instance. A trailing `/` is removed.

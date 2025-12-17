# Basic Example

This example demonstrates basic usage of the Go Invoice Ninja SDK.

## Features Demonstrated

- Creating a client with configuration options
- Listing payments with pagination
- Listing invoices
- Listing clients
- Proper error handling

## Prerequisites

1. An Invoice Ninja account (cloud or self-hosted)
2. An API token from Settings > Account Management > Integrations > API tokens

## Running the Example

```bash
# Set your API token
export INVOICE_NINJA_TOKEN="your-api-token-here"

# Run the example
go run main.go
```

## Expected Output

```
=== Listing Payments ===
Found 5 payments
  - Payment 0001: $250.00 (Client: abc123)
  - Payment 0002: $150.00 (Client: def456)
  ...

=== Listing Invoices ===
Found 5 invoices
  - Invoice INV-0001: $500.00 (Status: 4)
  ...

=== Listing Clients ===
Found 5 clients
  - Client 0001: Acme Corporation
  ...

✅ Basic example completed successfully!
```

## Self-Hosted Instances

For self-hosted Invoice Ninja instances, uncomment the `WithBaseURL` option:

```go
client := invoiceninja.NewClient(token,
    invoiceninja.WithBaseURL("https://your-instance.com"),
)
```

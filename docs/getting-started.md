# Getting Started

This guide will help you get started with the Go Invoice Ninja SDK.

## Installation

```bash
go get github.com/AshkanYarmoradi/go-invoice-ninja
```

## Prerequisites

Before using this SDK, you need:

1. **An Invoice Ninja Account**
   - Cloud: Sign up at [invoiceninja.com](https://invoiceninja.com)
   - Self-hosted: [Installation Guide](https://invoiceninja.github.io/docs/self-host-installation/)

2. **An API Token**
   - Go to Settings > Account Management > Integrations > API tokens
   - Create a new token and save it securely

## Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    invoiceninja "github.com/AshkanYarmoradi/go-invoice-ninja"
)

func main() {
    // Create a client
    client := invoiceninja.NewClient("your-api-token")
    ctx := context.Background()

    // List payments
    payments, err := client.Payments.List(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, p := range payments.Data {
        fmt.Printf("Payment: %s - $%.2f\n", p.Number, p.Amount)
    }
}
```

## Configuration Options

### Custom Base URL (Self-Hosted)

```go
client := invoiceninja.NewClient("token",
    invoiceninja.WithBaseURL("https://your-instance.com"))
```

### Custom Timeout

```go
client := invoiceninja.NewClient("token",
    invoiceninja.WithTimeout(60 * time.Second))
```

### Custom HTTP Client

```go
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:    100,
        IdleConnTimeout: 90 * time.Second,
    },
}

client := invoiceninja.NewClient("token",
    invoiceninja.WithHTTPClient(httpClient))
```

### Rate Limiting

```go
client := invoiceninja.NewClient("token",
    invoiceninja.WithRateLimiter(invoiceninja.NewRateLimiter(10))) // 10 req/sec
```

## Available Services

The SDK provides the following services:

| Service | Description |
|---------|-------------|
| `client.Payments` | Payment operations |
| `client.Invoices` | Invoice management |
| `client.Clients` | Client management |
| `client.Credits` | Credit operations |
| `client.PaymentTerms` | Payment terms |
| `client.Webhooks` | Webhook management |
| `client.Downloads` | File downloads |

## Next Steps

- [Authentication Guide](authentication.md)
- [Error Handling](error-handling.md)
- [API Reference](api-reference.md)
- [Examples](/examples/)

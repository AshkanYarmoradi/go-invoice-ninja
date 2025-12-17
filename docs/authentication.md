# Authentication

This guide covers authentication with the Invoice Ninja API.

## API Tokens

The Invoice Ninja API uses token-based authentication. All API requests must include your API token in the `X-API-TOKEN` header.

### Obtaining an API Token

1. Log in to your Invoice Ninja account
2. Navigate to Settings > Account Management > Integrations
3. Click on "API Tokens"
4. Create a new token
5. Copy the token immediately (it won't be shown again)

### Using Your Token

```go
client := invoiceninja.NewClient("your-api-token")
```

The SDK automatically includes the token in all requests.

## Security Best Practices

### Environment Variables

Never hardcode your API token. Use environment variables:

```go
import "os"

token := os.Getenv("INVOICE_NINJA_TOKEN")
if token == "" {
    log.Fatal("INVOICE_NINJA_TOKEN is required")
}

client := invoiceninja.NewClient(token)
```

### Token Permissions

When creating API tokens in Invoice Ninja, consider:

- **Create separate tokens** for different applications
- **Use descriptive names** to identify token usage
- **Rotate tokens periodically**
- **Revoke unused tokens**

### Secure Storage

- Store tokens in secure secret management systems
- Never commit tokens to version control
- Use `.env` files only for local development
- Encrypt tokens at rest in production

## Self-Hosted Authentication

For self-hosted instances, you may also need to provide an API secret:

```go
client := invoiceninja.NewClient("your-api-token",
    invoiceninja.WithBaseURL("https://your-instance.com"),
    invoiceninja.WithAPISecret("your-api-secret"), // if required
)
```

## Handling Authentication Errors

```go
payments, err := client.Payments.List(ctx, nil)
if err != nil {
    if apiErr, ok := err.(*invoiceninja.APIError); ok {
        if apiErr.IsUnauthorized() {
            log.Fatal("Invalid API token - please check your credentials")
        }
        if apiErr.IsForbidden() {
            log.Fatal("Access denied - check token permissions")
        }
    }
    log.Fatal(err)
}
```

## Token Rotation

To rotate your API token:

1. Generate a new token in Invoice Ninja
2. Update your application configuration
3. Deploy the new configuration
4. Revoke the old token

```go
// Example: Token rotation with graceful fallback
func createClient() *invoiceninja.Client {
    token := os.Getenv("INVOICE_NINJA_TOKEN")
    
    // Optional: Support for rotating tokens
    fallbackToken := os.Getenv("INVOICE_NINJA_TOKEN_FALLBACK")
    
    client := invoiceninja.NewClient(token)
    
    // Test the connection
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err := client.Payments.List(ctx, &invoiceninja.PaymentListOptions{PerPage: 1})
    if err != nil {
        if apiErr, ok := err.(*invoiceninja.APIError); ok && apiErr.IsUnauthorized() {
            if fallbackToken != "" {
                log.Println("Primary token failed, trying fallback")
                return invoiceninja.NewClient(fallbackToken)
            }
        }
        log.Fatalf("Authentication failed: %v", err)
    }
    
    return client
}
```

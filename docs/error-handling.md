# Error Handling

This guide covers error handling strategies when using the Go Invoice Ninja SDK.

## Error Types

### APIError

All API errors are returned as `*invoiceninja.APIError`:

```go
type APIError struct {
    StatusCode int                 // HTTP status code
    Message    string              // Error message
    Errors     map[string][]string // Field-specific validation errors
}
```

### Checking Error Types

```go
payments, err := client.Payments.List(ctx, nil)
if err != nil {
    if apiErr, ok := err.(*invoiceninja.APIError); ok {
        // Handle API-specific error
        fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
    } else {
        // Handle other errors (network, etc.)
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Error Helper Methods

The `APIError` type provides helper methods for common error types:

```go
if apiErr, ok := err.(*invoiceninja.APIError); ok {
    switch {
    case apiErr.IsNotFound():
        // 404 - Resource not found
        log.Printf("Resource not found")
        
    case apiErr.IsUnauthorized():
        // 401 - Invalid or missing API token
        log.Fatal("Please check your API token")
        
    case apiErr.IsForbidden():
        // 403 - Insufficient permissions
        log.Fatal("Access denied")
        
    case apiErr.IsValidationError():
        // 422 - Validation failed
        for field, errors := range apiErr.Errors {
            log.Printf("Field %s: %v", field, errors)
        }
        
    case apiErr.IsRateLimited():
        // 429 - Too many requests
        log.Println("Rate limited, waiting...")
        time.Sleep(time.Minute)
        
    case apiErr.IsServerError():
        // 5xx - Server error
        log.Println("Server error, please try again later")
    }
}
```

## Handling Validation Errors

Validation errors (422) include field-specific error messages:

```go
payment, err := client.Payments.Create(ctx, &invoiceninja.PaymentRequest{
    Amount: -100, // Invalid!
})
if err != nil {
    if apiErr, ok := err.(*invoiceninja.APIError); ok && apiErr.IsValidationError() {
        for field, errors := range apiErr.Errors {
            for _, e := range errors {
                fmt.Printf("Validation error on %s: %s\n", field, e)
            }
        }
    }
}
```

## Retry Configuration

The SDK includes automatic retry for transient errors:

```go
retryConfig := &invoiceninja.RetryConfig{
    MaxRetries:         3,
    InitialBackoff:     time.Second,
    MaxBackoff:         30 * time.Second,
    BackoffMultiplier:  2.0,
    RetryOnStatusCodes: []int{429, 500, 502, 503, 504},
    Jitter:             true,
}

client := invoiceninja.NewClient("token",
    invoiceninja.WithRetryConfig(retryConfig))
```

### Default Retry Behavior

By default, the SDK retries on:
- **429** - Rate limit exceeded
- **500** - Internal server error
- **502** - Bad gateway
- **503** - Service unavailable
- **504** - Gateway timeout

With exponential backoff: 1s → 2s → 4s (with jitter)

## Rate Limiting

### Server-Side Rate Limits

Invoice Ninja enforces rate limits. When exceeded, you'll receive a 429 error:

```go
if apiErr.IsRateLimited() {
    // Check for Retry-After header
    retryAfter := apiErr.Headers.Get("Retry-After")
    if retryAfter != "" {
        duration, _ := strconv.Atoi(retryAfter)
        time.Sleep(time.Duration(duration) * time.Second)
    }
}
```

### Client-Side Rate Limiting

Prevent hitting rate limits with client-side limiting:

```go
// Limit to 10 requests per second
rateLimiter := invoiceninja.NewRateLimiter(10)

client := invoiceninja.NewClient("token",
    invoiceninja.WithRateLimiter(rateLimiter))
```

## Context and Timeouts

Always use context for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

payments, err := client.Payments.List(ctx, nil)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Request timed out")
    }
    if errors.Is(err, context.Canceled) {
        log.Println("Request was canceled")
    }
}
```

## Best Practices

### 1. Always Check Errors

```go
// Bad
payments, _ := client.Payments.List(ctx, nil)

// Good
payments, err := client.Payments.List(ctx, nil)
if err != nil {
    return fmt.Errorf("listing payments: %w", err)
}
```

### 2. Use Error Wrapping

```go
invoice, err := client.Invoices.Get(ctx, invoiceID)
if err != nil {
    return nil, fmt.Errorf("fetching invoice %s: %w", invoiceID, err)
}
```

### 3. Log Contextual Information

```go
if apiErr, ok := err.(*invoiceninja.APIError); ok {
    log.Printf("API error: status=%d message=%s endpoint=%s",
        apiErr.StatusCode,
        apiErr.Message,
        "/api/v1/payments",
    )
}
```

### 4. Implement Circuit Breakers for Critical Paths

For production systems, consider implementing circuit breakers:

```go
// Pseudo-code with a circuit breaker library
breaker := circuitbreaker.New(circuitbreaker.Config{
    MaxFailures: 3,
    Timeout:     time.Minute,
})

result, err := breaker.Execute(func() (interface{}, error) {
    return client.Payments.List(ctx, nil)
})
```

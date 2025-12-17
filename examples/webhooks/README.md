# Webhooks Example

This example demonstrates how to handle webhooks from Invoice Ninja.

## Features Demonstrated

- Setting up a webhook HTTP endpoint
- Verifying webhook signatures (HMAC-SHA256)
- Parsing webhook events
- Handling different event types

## Prerequisites

1. An Invoice Ninja account (cloud or self-hosted)
2. A publicly accessible URL (for Invoice Ninja to send webhooks)
3. A webhook secret configured in Invoice Ninja

## Setting Up Webhooks in Invoice Ninja

1. Go to Settings > Webhooks
2. Create a new webhook
3. Set the Target URL to your server's webhook endpoint
4. Copy the Secret and set it as `INVOICE_NINJA_WEBHOOK_SECRET`
5. Select the events you want to receive

## Running the Example

```bash
# Set your webhook secret
export INVOICE_NINJA_WEBHOOK_SECRET="your-webhook-secret-here"

# Optional: Set a custom port
export PORT=8080

# Run the server
go run main.go
```

## Testing Locally

For local development, you can use a tool like [ngrok](https://ngrok.com/) to expose your local server:

```bash
# In one terminal, run the webhook server
go run main.go

# In another terminal, expose it with ngrok
ngrok http 8080
```

Then use the ngrok URL as your webhook endpoint in Invoice Ninja.

## Supported Events

This example handles the following events:

- `payment.created` - When a new payment is recorded
- `invoice.paid` - When an invoice is fully paid
- `client.created` - When a new client is created

## Security

Always verify webhook signatures to ensure the request came from Invoice Ninja:

```go
if !webhookHandler.VerifySignature(body, signature) {
    // Reject the request
}
```

# Webhooks Example

This example demonstrates how to handle webhooks from Invoice Ninja.

## Features Demonstrated

- Setting up a webhook HTTP endpoint
- Authenticating requests with a shared secret header
- Routing events to handlers
- Parsing the payment, invoice or client sent in a webhook

## How Invoice Ninja Sends Webhooks

Each webhook in Invoice Ninja is set up for one event. When the event happens, Invoice Ninja sends the entity (for example the payment) as JSON to the webhook's target URL. The request contains neither the event name nor a signature, so:

- **The event name** comes from the target URL (`?event=payment.created`) or from an `X-Webhook-Event` header that you add to the webhook.
- **Authentication** uses a shared secret: add an `X-Webhook-Secret` header with your secret to the webhook, and pass the same secret to `NewWebhookHandler`.

## Prerequisites

1. An Invoice Ninja account (cloud or self-hosted)
2. A publicly accessible URL (for Invoice Ninja to send webhooks)
3. A random secret, for example from `openssl rand -hex 32`

## Setting Up Webhooks in Invoice Ninja

1. Go to Settings > Account Management > Integrations > API Webhooks
2. Create a webhook for each event you want to receive
3. Select the event, and set the Target URL to your endpoint with the event name, for example `https://your-server.com/webhook?event=payment.created`
4. Keep the method as POST (PUT also works)
5. Add a header named `X-Webhook-Secret` with your secret as the value

## Event Names

Use these event names so the built-in helpers match:

| Invoice Ninja event | Event name | Helper |
|---------------------|------------|--------|
| Create Invoice | `invoice.created` | `OnInvoiceCreated` |
| Update Invoice | `invoice.updated` | `OnInvoiceUpdated` |
| Delete Invoice | `invoice.deleted` | `OnInvoiceDeleted` |
| Create Payment | `payment.created` | `OnPaymentCreated` |
| Update Payment | `payment.updated` | `OnPaymentUpdated` |
| Delete Payment | `payment.deleted` | `OnPaymentDeleted` |
| Create Client | `client.created` | `OnClientCreated` |
| Update Client | `client.updated` | `OnClientUpdated` |
| Create Credit | `credit.created` | `OnCreditCreated` |
| Create Quote | `quote.created` | `OnQuoteCreated` |

For other events, choose a name and register it with `webhookHandler.On("name", ...)`.

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

You can also send a test request with curl:

```bash
curl -X POST "http://localhost:8080/webhook?event=payment.created" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: $INVOICE_NINJA_WEBHOOK_SECRET" \
  -d '{"id":"pay123","number":"0001","amount":100}'
```

## Supported Events

This example handles the following events:

- `payment.created` - When a new payment is recorded
- `invoice.created` - When a new invoice is created
- `invoice.updated` - When an invoice is updated
- `client.created` - When a new client is created

## Security

- Always set a secret in production. Without one, anyone who knows the URL can send events.
- Use HTTPS, so the secret header is encrypted in transit.
- The handler compares secrets in constant time and rejects request bodies larger than 10 MB.

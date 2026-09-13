package invoiceninja

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Invoice Ninja webhook requests contain only the entity that changed (for example an
// invoice or a payment) as JSON. They carry neither the event name nor a signature, so
// the event name and a shared secret come from the target URL and headers configured
// for each webhook in Invoice Ninja.
const (
	// WebhookEventHeader is the request header that names the event, such as "payment.created".
	WebhookEventHeader = "X-Webhook-Event"

	// WebhookEventQueryParam is the query parameter that names the event when it is part of
	// the webhook's target URL, as in https://example.com/webhook?event=payment.created.
	WebhookEventQueryParam = "event"

	// WebhookSecretHeader is the request header that must contain the secret passed to NewWebhookHandler.
	WebhookSecretHeader = "X-Webhook-Secret" //nolint:gosec // G101 false positive: this is a header name, not a credential.

	// MaxWebhookBodyBytes is the largest request body WebhookHandler accepts (10 MB).
	MaxWebhookBodyBytes = 10 << 20
)

// Headers that may carry a hex-encoded HMAC-SHA256 signature of the body from custom senders.
const (
	signatureHeader       = "X-Ninja-Signature"
	legacySignatureHeader = "X-Invoice-Ninja-Signature"
)

// WebhookEvent represents a webhook event received from Invoice Ninja.
type WebhookEvent struct {
	// EventType is the event name, such as "invoice.created" or "payment.created".
	EventType string `json:"event_type"`

	// Data contains the entity JSON, such as an invoice or a payment.
	Data json.RawMessage `json:"data"`
}

// WebhookHandler handles incoming webhook requests from Invoice Ninja.
type WebhookHandler struct {
	// secret is the shared secret used to authenticate requests.
	secret string

	// handlers maps event types to handler functions.
	handlers map[string]WebhookEventHandler
}

// WebhookEventHandler is a function that handles a specific webhook event.
type WebhookEventHandler func(event *WebhookEvent) error

// NewWebhookHandler creates a new webhook handler.
//
// If secret is not empty, every request must include it in the X-Webhook-Secret header.
// Invoice Ninja does not sign webhook requests, so add this header to each webhook in
// Invoice Ninja (Settings > Account Management > Integrations > API Webhooks). For custom
// senders, a hex-encoded HMAC-SHA256 signature of the body in the X-Ninja-Signature
// header is accepted instead.
func NewWebhookHandler(secret string) *WebhookHandler {
	return &WebhookHandler{
		secret:   secret,
		handlers: make(map[string]WebhookEventHandler),
	}
}

// On registers a handler for an event name, such as "invoice.created".
func (h *WebhookHandler) On(eventType string, handler WebhookEventHandler) {
	h.handlers[eventType] = handler
}

// OnInvoiceCreated registers a handler for invoice.created events.
func (h *WebhookHandler) OnInvoiceCreated(handler WebhookEventHandler) {
	h.On("invoice.created", handler)
}

// OnInvoiceUpdated registers a handler for invoice.updated events.
func (h *WebhookHandler) OnInvoiceUpdated(handler WebhookEventHandler) {
	h.On("invoice.updated", handler)
}

// OnInvoiceDeleted registers a handler for invoice.deleted events.
func (h *WebhookHandler) OnInvoiceDeleted(handler WebhookEventHandler) {
	h.On("invoice.deleted", handler)
}

// OnPaymentCreated registers a handler for payment.created events.
func (h *WebhookHandler) OnPaymentCreated(handler WebhookEventHandler) {
	h.On("payment.created", handler)
}

// OnPaymentUpdated registers a handler for payment.updated events.
func (h *WebhookHandler) OnPaymentUpdated(handler WebhookEventHandler) {
	h.On("payment.updated", handler)
}

// OnPaymentDeleted registers a handler for payment.deleted events.
func (h *WebhookHandler) OnPaymentDeleted(handler WebhookEventHandler) {
	h.On("payment.deleted", handler)
}

// OnClientCreated registers a handler for client.created events.
func (h *WebhookHandler) OnClientCreated(handler WebhookEventHandler) {
	h.On("client.created", handler)
}

// OnClientUpdated registers a handler for client.updated events.
func (h *WebhookHandler) OnClientUpdated(handler WebhookEventHandler) {
	h.On("client.updated", handler)
}

// OnCreditCreated registers a handler for credit.created events.
func (h *WebhookHandler) OnCreditCreated(handler WebhookEventHandler) {
	h.On("credit.created", handler)
}

// OnQuoteCreated registers a handler for quote.created events.
func (h *WebhookHandler) OnQuoteCreated(handler WebhookEventHandler) {
	h.On("quote.created", handler)
}

// HandleRequest processes an incoming webhook HTTP request.
//
// The event name is read from the X-Webhook-Event header, then from the "event" query
// parameter. A JSON body of the form {"event_type": "...", "data": {...}} is also
// accepted. The response is 200 when the event was handled or no handler is registered
// for it.
func (h *WebhookHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		w.Header().Set("Allow", "POST, PUT")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check the secret header before reading the body
	secretVerified := false
	if h.secret != "" {
		if secret := r.Header.Get(WebhookSecretHeader); secret != "" {
			if !secretsEqual(secret, h.secret) {
				http.Error(w, "Invalid secret", http.StatusUnauthorized)
				return
			}
			secretVerified = true
		} else if requestSignature(r) == "" {
			http.Error(w, "Missing secret", http.StatusUnauthorized)
			return
		}
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxWebhookBodyBytes))
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Verify the signature of custom senders that don't send the secret header
	if h.secret != "" && !secretVerified && !h.verifySignature(body, requestSignature(r)) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	event, err := parseWebhookEvent(r, body)
	if err != nil {
		http.Error(w, "Failed to parse webhook payload", http.StatusBadRequest)
		return
	}

	if event.EventType == "" {
		http.Error(w, fmt.Sprintf("Missing event name: set the %s header or the %q query parameter",
			WebhookEventHeader, WebhookEventQueryParam), http.StatusBadRequest)
		return
	}

	// Find and execute the handler
	handler, ok := h.handlers[event.EventType]
	if !ok {
		// No handler registered for this event type, acknowledge receipt
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := handler(event); err != nil {
		// Don't send internal error details back to the caller
		http.Error(w, "Webhook handler failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// parseWebhookEvent builds a WebhookEvent from the request and its body.
func parseWebhookEvent(r *http.Request, body []byte) (*WebhookEvent, error) {
	var envelope WebhookEvent
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}

	event := &WebhookEvent{
		EventType: r.Header.Get(WebhookEventHeader),
		Data:      json.RawMessage(body),
	}
	if event.EventType == "" {
		event.EventType = r.URL.Query().Get(WebhookEventQueryParam)
	}

	// Envelope format used by custom senders: {"event_type": "...", "data": {...}}
	if envelope.EventType != "" {
		if event.EventType == "" {
			event.EventType = envelope.EventType
		}
		event.Data = envelope.Data
	}

	return event, nil
}

// requestSignature returns the HMAC signature sent with the request, if any.
func requestSignature(r *http.Request) string {
	if signature := r.Header.Get(signatureHeader); signature != "" {
		return signature
	}
	return r.Header.Get(legacySignatureHeader)
}

// secretsEqual compares two secrets in constant time.
func secretsEqual(a, b string) bool {
	hashA := sha256.Sum256([]byte(a))
	hashB := sha256.Sum256([]byte(b))
	return hmac.Equal(hashA[:], hashB[:])
}

// verifySignature verifies a hex-encoded HMAC-SHA256 signature of the payload.
func (h *WebhookHandler) verifySignature(payload []byte, signature string) bool {
	if signature == "" {
		return false
	}

	// Remove "sha256=" prefix if present
	signature = strings.TrimPrefix(signature, "sha256=")

	mac := hmac.New(sha256.New, []byte(h.secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// ParseInvoice parses the webhook data as an Invoice.
func (e *WebhookEvent) ParseInvoice() (*Invoice, error) {
	var invoice Invoice
	if err := json.Unmarshal(e.Data, &invoice); err != nil {
		return nil, fmt.Errorf("failed to parse invoice data: %w", err)
	}
	return &invoice, nil
}

// ParsePayment parses the webhook data as a Payment.
func (e *WebhookEvent) ParsePayment() (*Payment, error) {
	var payment Payment
	if err := json.Unmarshal(e.Data, &payment); err != nil {
		return nil, fmt.Errorf("failed to parse payment data: %w", err)
	}
	return &payment, nil
}

// ParseClient parses the webhook data as a Client.
func (e *WebhookEvent) ParseClient() (*INClient, error) {
	var client INClient
	if err := json.Unmarshal(e.Data, &client); err != nil {
		return nil, fmt.Errorf("failed to parse client data: %w", err)
	}
	return &client, nil
}

// ParseCredit parses the webhook data as a Credit.
func (e *WebhookEvent) ParseCredit() (*Credit, error) {
	var credit Credit
	if err := json.Unmarshal(e.Data, &credit); err != nil {
		return nil, fmt.Errorf("failed to parse credit data: %w", err)
	}
	return &credit, nil
}

// ServeHTTP implements http.Handler interface.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.HandleRequest(w, r)
}

package invoiceninja

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// WebhookEvent represents an Invoice Ninja webhook event.
type WebhookEvent struct {
	// EventType is the type of event (e.g., "invoice.created", "payment.created").
	EventType string `json:"event_type"`

	// Data contains the event payload.
	Data json.RawMessage `json:"data"`
}

// WebhookHandler handles incoming webhook requests from Invoice Ninja.
type WebhookHandler struct {
	// secret is the webhook signing secret for signature verification.
	secret string

	// handlers maps event types to handler functions.
	handlers map[string]WebhookEventHandler
}

// WebhookEventHandler is a function that handles a specific webhook event.
type WebhookEventHandler func(event *WebhookEvent) error

// NewWebhookHandler creates a new webhook handler.
// If secret is provided, signature verification will be enforced.
func NewWebhookHandler(secret string) *WebhookHandler {
	return &WebhookHandler{
		secret:   secret,
		handlers: make(map[string]WebhookEventHandler),
	}
}

// On registers a handler for a specific event type.
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
func (h *WebhookHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature if secret is configured
	if h.secret != "" {
		signature := r.Header.Get("X-Ninja-Signature")
		if signature == "" {
			signature = r.Header.Get("X-Invoice-Ninja-Signature")
		}

		if !h.verifySignature(body, signature) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Failed to parse webhook payload", http.StatusBadRequest)
		return
	}

	// Find and execute the handler
	handler, ok := h.handlers[event.EventType]
	if !ok {
		// No handler registered for this event type, acknowledge receipt
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := handler(&event); err != nil {
		http.Error(w, fmt.Sprintf("Handler error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// verifySignature verifies the webhook signature.
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

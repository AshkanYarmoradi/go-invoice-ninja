package invoiceninja

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// invoiceNinjaPayment is a payment webhook body as Invoice Ninja sends it: only the entity.
const invoiceNinjaPayment = `{"id":"pay123","amount":100.5,"client_id":"client123","number":"0001"}`

func TestWebhookHandler(t *testing.T) {
	handler := NewWebhookHandler("")

	var receivedEvent *WebhookEvent
	handler.OnPaymentCreated(func(event *WebhookEvent) error {
		receivedEvent = event
		return nil
	})

	payload := map[string]interface{}{
		"event_type": "payment.created",
		"data": map[string]interface{}{
			"id":     "pay123",
			"amount": 100.00,
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if receivedEvent == nil {
		t.Fatal("expected event to be received")
	}

	if receivedEvent.EventType != "payment.created" {
		t.Errorf("expected event type 'payment.created', got '%s'", receivedEvent.EventType)
	}
}

func TestWebhookHandlerInvoiceNinjaPayload(t *testing.T) {
	handler := NewWebhookHandler("")

	var receivedPayment *Payment
	handler.OnPaymentCreated(func(event *WebhookEvent) error {
		payment, err := event.ParsePayment()
		if err != nil {
			return err
		}
		receivedPayment = payment
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(invoiceNinjaPayment))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(WebhookEventHeader, "payment.created")

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if receivedPayment == nil {
		t.Fatal("expected payment handler to be called")
	}

	if receivedPayment.ID != "pay123" || receivedPayment.Amount != 100.5 {
		t.Errorf("expected payment pay123 with amount 100.5, got %s with amount %f", receivedPayment.ID, receivedPayment.Amount)
	}
}

func TestWebhookHandlerEventFromQueryParameter(t *testing.T) {
	handler := NewWebhookHandler("")

	var receivedInvoice *Invoice
	handler.OnInvoiceUpdated(func(event *WebhookEvent) error {
		invoice, err := event.ParseInvoice()
		if err != nil {
			return err
		}
		receivedInvoice = invoice
		return nil
	})

	body := `{"id":"inv123","number":"INV001","amount":500,"entity_type":"invoice"}`
	req := httptest.NewRequest(http.MethodPost, "/webhook?event=invoice.updated", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if receivedInvoice == nil {
		t.Fatal("expected invoice handler to be called")
	}

	if receivedInvoice.Number != "INV001" {
		t.Errorf("expected invoice number 'INV001', got '%s'", receivedInvoice.Number)
	}
}

func TestWebhookHandlerAcceptsPut(t *testing.T) {
	handler := NewWebhookHandler("")

	called := false
	handler.OnPaymentCreated(func(event *WebhookEvent) error {
		called = true
		return nil
	})

	req := httptest.NewRequest(http.MethodPut, "/webhook?event=payment.created", strings.NewReader(invoiceNinjaPayment))

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !called {
		t.Error("expected handler to be called for a PUT request")
	}
}

func TestWebhookHandlerMissingEventName(t *testing.T) {
	handler := NewWebhookHandler("")

	called := false
	handler.OnPaymentCreated(func(event *WebhookEvent) error {
		called = true
		return nil
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(invoiceNinjaPayment))

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 without an event name, got %d", w.Code)
	}

	if called {
		t.Error("expected handler not to be called without an event name")
	}
}

func TestWebhookHandlerInvalidJSON(t *testing.T) {
	handler := NewWebhookHandler("")

	req := httptest.NewRequest(http.MethodPost, "/webhook?event=payment.created", strings.NewReader("not json"))

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid JSON, got %d", w.Code)
	}
}

func TestWebhookHandlerSecretHeader(t *testing.T) {
	tests := []struct {
		name           string
		secret         string
		expectedStatus int
		expectCalled   bool
	}{
		{
			name:           "valid secret",
			secret:         "test-secret",
			expectedStatus: http.StatusOK,
			expectCalled:   true,
		},
		{
			name:           "wrong secret",
			secret:         "wrong-secret",
			expectedStatus: http.StatusUnauthorized,
			expectCalled:   false,
		},
		{
			name:           "missing secret",
			secret:         "",
			expectedStatus: http.StatusUnauthorized,
			expectCalled:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewWebhookHandler("test-secret")

			called := false
			handler.OnPaymentCreated(func(event *WebhookEvent) error {
				called = true
				return nil
			})

			req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(invoiceNinjaPayment))
			req.Header.Set(WebhookEventHeader, "payment.created")
			if tt.secret != "" {
				req.Header.Set(WebhookSecretHeader, tt.secret)
			}

			w := httptest.NewRecorder()
			handler.HandleRequest(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if called != tt.expectCalled {
				t.Errorf("expected handler called = %v, got %v", tt.expectCalled, called)
			}
		})
	}
}

func TestWebhookHandlerWithSignature(t *testing.T) {
	secret := "test-secret"
	handler := NewWebhookHandler(secret)

	handler.OnInvoiceCreated(func(event *WebhookEvent) error {
		return nil
	})

	payload := []byte(`{"event_type":"invoice.created","data":{"id":"inv123"}}`)

	// Test with invalid signature
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Ninja-Signature", "invalid-signature")

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 for invalid signature, got %d", w.Code)
	}
}

func TestWebhookHandlerWithValidSignature(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"event_type":"invoice.created","data":{"id":"inv123"}}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	for _, headerValue := range []string{signature, "sha256=" + signature} {
		handler := NewWebhookHandler(secret)

		var receivedEvent *WebhookEvent
		handler.OnInvoiceCreated(func(event *WebhookEvent) error {
			receivedEvent = event
			return nil
		})

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Ninja-Signature", headerValue)

		w := httptest.NewRecorder()
		handler.HandleRequest(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("signature %q: expected status 200, got %d", headerValue, w.Code)
		}

		if receivedEvent == nil {
			t.Fatalf("signature %q: expected handler to be called", headerValue)
		}

		invoice, err := receivedEvent.ParseInvoice()
		if err != nil {
			t.Fatalf("failed to parse invoice: %v", err)
		}
		if invoice.ID != "inv123" {
			t.Errorf("expected invoice ID 'inv123', got '%s'", invoice.ID)
		}
	}
}

func TestWebhookHandlerMethodNotAllowed(t *testing.T) {
	handler := NewWebhookHandler("")

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
	}
}

func TestWebhookHandlerBodyTooLarge(t *testing.T) {
	handler := NewWebhookHandler("")

	body := bytes.Repeat([]byte("a"), MaxWebhookBodyBytes+1)
	req := httptest.NewRequest(http.MethodPost, "/webhook?event=payment.created", bytes.NewReader(body))

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 413, got %d", w.Code)
	}
}

func TestWebhookHandlerErrorIsNotExposed(t *testing.T) {
	handler := NewWebhookHandler("")

	handler.OnPaymentCreated(func(event *WebhookEvent) error {
		return errors.New("database password is hunter2")
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook?event=payment.created", strings.NewReader(invoiceNinjaPayment))

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	if strings.Contains(w.Body.String(), "hunter2") {
		t.Errorf("expected handler error details not to be sent, got %q", w.Body.String())
	}
}

func TestWebhookHandlerUnregisteredEvent(t *testing.T) {
	handler := NewWebhookHandler("")

	payload := []byte(`{"event_type":"unknown.event","data":{}}`)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	// Should still return 200 for unregistered events
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for unregistered event, got %d", w.Code)
	}
}

func TestWebhookEventParsers(t *testing.T) {
	// Test ParseInvoice
	invoiceEvent := &WebhookEvent{
		EventType: "invoice.created",
		Data:      json.RawMessage(`{"id":"inv123","number":"INV001","amount":500.00}`),
	}

	invoice, err := invoiceEvent.ParseInvoice()
	if err != nil {
		t.Fatalf("failed to parse invoice: %v", err)
	}
	if invoice.ID != "inv123" {
		t.Errorf("expected invoice ID 'inv123', got '%s'", invoice.ID)
	}
	if invoice.Amount != 500.00 {
		t.Errorf("expected amount 500.00, got %f", invoice.Amount)
	}

	// Test ParsePayment
	paymentEvent := &WebhookEvent{
		EventType: "payment.created",
		Data:      json.RawMessage(`{"id":"pay123","amount":100.00,"client_id":"client123"}`),
	}

	payment, err := paymentEvent.ParsePayment()
	if err != nil {
		t.Fatalf("failed to parse payment: %v", err)
	}
	if payment.ID != "pay123" {
		t.Errorf("expected payment ID 'pay123', got '%s'", payment.ID)
	}

	// Test ParseClient
	clientEvent := &WebhookEvent{
		EventType: "client.created",
		Data:      json.RawMessage(`{"id":"client123","name":"Acme Corp"}`),
	}

	client, err := clientEvent.ParseClient()
	if err != nil {
		t.Fatalf("failed to parse client: %v", err)
	}
	if client.ID != "client123" {
		t.Errorf("expected client ID 'client123', got '%s'", client.ID)
	}
	if client.Name != "Acme Corp" {
		t.Errorf("expected client name 'Acme Corp', got '%s'", client.Name)
	}

	// Test ParseCredit
	creditEvent := &WebhookEvent{
		EventType: "credit.created",
		Data:      json.RawMessage(`{"id":"credit123","number":"CR001","amount":50.00}`),
	}

	credit, err := creditEvent.ParseCredit()
	if err != nil {
		t.Fatalf("failed to parse credit: %v", err)
	}
	if credit.ID != "credit123" {
		t.Errorf("expected credit ID 'credit123', got '%s'", credit.ID)
	}
}

func TestWebhookHandlerServeHTTP(t *testing.T) {
	handler := NewWebhookHandler("")

	called := false
	handler.OnPaymentCreated(func(event *WebhookEvent) error {
		called = true
		return nil
	})

	payload := []byte(`{"event_type":"payment.created","data":{"id":"pay123"}}`)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestWebhookHandlerRegistrations(t *testing.T) {
	handler := NewWebhookHandler("")

	events := []string{
		"invoice.created",
		"invoice.updated",
		"invoice.deleted",
		"payment.created",
		"payment.updated",
		"payment.deleted",
		"client.created",
		"client.updated",
		"credit.created",
		"quote.created",
	}

	// Register handlers for all event types
	handler.OnInvoiceCreated(func(e *WebhookEvent) error { return nil })
	handler.OnInvoiceUpdated(func(e *WebhookEvent) error { return nil })
	handler.OnInvoiceDeleted(func(e *WebhookEvent) error { return nil })
	handler.OnPaymentCreated(func(e *WebhookEvent) error { return nil })
	handler.OnPaymentUpdated(func(e *WebhookEvent) error { return nil })
	handler.OnPaymentDeleted(func(e *WebhookEvent) error { return nil })
	handler.OnClientCreated(func(e *WebhookEvent) error { return nil })
	handler.OnClientUpdated(func(e *WebhookEvent) error { return nil })
	handler.OnCreditCreated(func(e *WebhookEvent) error { return nil })
	handler.OnQuoteCreated(func(e *WebhookEvent) error { return nil })

	// Verify all handlers are registered
	for _, eventType := range events {
		if _, ok := handler.handlers[eventType]; !ok {
			t.Errorf("expected handler for event type '%s' to be registered", eventType)
		}
	}
}

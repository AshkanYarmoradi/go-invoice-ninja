package invoiceninja

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestWebhookHandlerMethodNotAllowed(t *testing.T) {
	handler := NewWebhookHandler("")

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)

	w := httptest.NewRecorder()
	handler.HandleRequest(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", w.Code)
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

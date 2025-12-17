package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInvoicesServiceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices" {
			t.Errorf("expected path /api/v1/invoices, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "inv123", "number": "INV001", "amount": 500.00},
				{"id": "inv456", "number": "INV002", "amount": 750.00},
			},
			"meta": map[string]interface{}{
				"pagination": map[string]interface{}{
					"total":        25,
					"count":        2,
					"per_page":     20,
					"current_page": 1,
					"total_pages":  2,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	resp, err := client.Invoices.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 invoices, got %d", len(resp.Data))
	}

	if resp.Data[0].Number != "INV001" {
		t.Errorf("expected first invoice number to be 'INV001', got '%s'", resp.Data[0].Number)
	}
}

func TestInvoicesServiceGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices/inv123" {
			t.Errorf("expected path /api/v1/invoices/inv123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "inv123",
				"number":    "INV001",
				"client_id": "client123",
				"amount":    500.00,
				"balance":   250.00,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	invoice, err := client.Invoices.Get(context.Background(), "inv123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.ID != "inv123" {
		t.Errorf("expected invoice ID to be 'inv123', got '%s'", invoice.ID)
	}

	if invoice.Balance != 250.00 {
		t.Errorf("expected balance to be 250.00, got %f", invoice.Balance)
	}
}

func TestInvoicesServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices" {
			t.Errorf("expected path /api/v1/invoices, got %s", r.URL.Path)
		}

		var body Invoice
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.ClientID != "client123" {
			t.Errorf("expected client_id to be 'client123', got '%s'", body.ClientID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "newinv123",
				"client_id": "client123",
				"number":    "INV003",
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	req := &Invoice{
		ClientID: "client123",
		LineItems: []LineItem{
			{ProductKey: "Product A", Quantity: 2, Cost: 100.00},
		},
	}

	invoice, err := client.Invoices.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.ID != "newinv123" {
		t.Errorf("expected invoice ID to be 'newinv123', got '%s'", invoice.ID)
	}
}

func TestInvoicesServiceUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices/inv123" {
			t.Errorf("expected path /api/v1/invoices/inv123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "inv123",
				"po_number": "PO-12345",
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	req := &Invoice{
		PONumber: "PO-12345",
	}

	invoice, err := client.Invoices.Update(context.Background(), "inv123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.ID != "inv123" {
		t.Errorf("expected invoice ID to be 'inv123', got '%s'", invoice.ID)
	}
}

func TestInvoicesServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices/inv123" {
			t.Errorf("expected path /api/v1/invoices/inv123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	err := client.Invoices.Delete(context.Background(), "inv123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvoicesServiceBulk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/invoices/bulk" {
			t.Errorf("expected path /api/v1/invoices/bulk, got %s", r.URL.Path)
		}

		var body BulkAction
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.Action != "mark_paid" {
			t.Errorf("expected action to be 'mark_paid', got '%s'", body.Action)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "inv123", "status_id": "4"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	invoices, err := client.Invoices.Bulk(context.Background(), "mark_paid", []string{"inv123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invoices) != 1 {
		t.Errorf("expected 1 invoice, got %d", len(invoices))
	}
}

func TestInvoicesServiceMarkPaid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "inv123", "status_id": "4"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	invoice, err := client.Invoices.MarkPaid(context.Background(), "inv123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.ID != "inv123" {
		t.Errorf("expected invoice ID to be 'inv123', got '%s'", invoice.ID)
	}
}

func TestInvoiceListOptionsToQuery(t *testing.T) {
	isDeleted := false
	opts := &InvoiceListOptions{
		PerPage:   25,
		Page:      3,
		Filter:    "search term",
		ClientID:  "client456",
		Status:    "active",
		CreatedAt: "2024-02-01",
		UpdatedAt: "2024-02-15",
		IsDeleted: &isDeleted,
		Sort:      "number|asc",
		Include:   "payments",
	}

	q := opts.toQuery()

	if q.Get("per_page") != "25" {
		t.Errorf("expected per_page=25, got %s", q.Get("per_page"))
	}
	if q.Get("page") != "3" {
		t.Errorf("expected page=3, got %s", q.Get("page"))
	}
	if q.Get("filter") != "search term" {
		t.Errorf("expected filter='search term', got %s", q.Get("filter"))
	}
	if q.Get("client_id") != "client456" {
		t.Errorf("expected client_id=client456, got %s", q.Get("client_id"))
	}
	if q.Get("is_deleted") != "false" {
		t.Errorf("expected is_deleted=false, got %s", q.Get("is_deleted"))
	}
}

func TestInvoiceListOptionsNilToQuery(t *testing.T) {
	var opts *InvoiceListOptions = nil
	q := opts.toQuery()
	if q != nil {
		t.Error("expected nil query for nil options")
	}
}

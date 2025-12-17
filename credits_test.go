package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreditsServiceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/credits" {
			t.Errorf("expected path /api/v1/credits, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "credit1", "number": "CR001", "amount": 100.00},
				{"id": "credit2", "number": "CR002", "amount": 200.00},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	resp, err := client.Credits.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 credits, got %d", len(resp.Data))
	}

	if resp.Data[0].Number != "CR001" {
		t.Errorf("expected first credit number to be 'CR001', got '%s'", resp.Data[0].Number)
	}
}

func TestCreditsServiceGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/credits/credit1" {
			t.Errorf("expected path /api/v1/credits/credit1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "credit1",
				"number":    "CR001",
				"amount":    100.00,
				"balance":   50.00,
				"client_id": "client123",
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	credit, err := client.Credits.Get(context.Background(), "credit1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if credit.ID != "credit1" {
		t.Errorf("expected credit ID to be 'credit1', got '%s'", credit.ID)
	}

	if credit.Balance != 50.00 {
		t.Errorf("expected balance to be 50.00, got %f", credit.Balance)
	}
}

func TestCreditsServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var body Credit
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.ClientID != "client123" {
			t.Errorf("expected client_id to be 'client123', got '%s'", body.ClientID)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "newcredit",
				"client_id": "client123",
				"number":    "CR003",
				"amount":    150.00,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	credit, err := client.Credits.Create(context.Background(), &Credit{
		ClientID: "client123",
		LineItems: []LineItem{
			{ProductKey: "Credit Item", Quantity: 1, Cost: 150.00},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if credit.ID != "newcredit" {
		t.Errorf("expected credit ID to be 'newcredit', got '%s'", credit.ID)
	}
}

func TestCreditsServiceUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":            "credit1",
				"private_notes": "Updated notes",
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	credit, err := client.Credits.Update(context.Background(), "credit1", &Credit{
		PrivateNotes: "Updated notes",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if credit.ID != "credit1" {
		t.Errorf("expected credit ID to be 'credit1', got '%s'", credit.ID)
	}
}

func TestCreditsServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	err := client.Credits.Delete(context.Background(), "credit1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreditsServiceBulk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/credits/bulk" {
			t.Errorf("expected path /api/v1/credits/bulk, got %s", r.URL.Path)
		}

		var body BulkAction
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.Action != "mark_sent" {
			t.Errorf("expected action to be 'mark_sent', got '%s'", body.Action)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "credit1"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	credits, err := client.Credits.Bulk(context.Background(), "mark_sent", []string{"credit1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(credits) != 1 {
		t.Errorf("expected 1 credit, got %d", len(credits))
	}
}

func TestCreditListOptionsToQuery(t *testing.T) {
	isDeleted := false
	opts := &CreditListOptions{
		PerPage:   25,
		Page:      3,
		Filter:    "search term",
		ClientID:  "client456",
		Status:    "active",
		CreatedAt: "2024-02-01",
		UpdatedAt: "2024-02-15",
		IsDeleted: &isDeleted,
		Sort:      "number|asc",
		Include:   "client",
	}

	q := opts.toQuery()

	if q.Get("per_page") != "25" {
		t.Errorf("expected per_page=25, got %s", q.Get("per_page"))
	}
	if q.Get("client_id") != "client456" {
		t.Errorf("expected client_id=client456, got %s", q.Get("client_id"))
	}
	if q.Get("is_deleted") != "false" {
		t.Errorf("expected is_deleted=false, got %s", q.Get("is_deleted"))
	}
}

func TestCreditListOptionsNilToQuery(t *testing.T) {
	var opts *CreditListOptions = nil
	q := opts.toQuery()
	if q != nil {
		t.Error("expected nil query for nil options")
	}
}

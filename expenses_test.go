package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExpensesServiceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/expenses" {
			t.Errorf("expected path /api/v1/expenses, got %s", r.URL.Path)
		}

		// Check query parameters
		if r.URL.Query().Get("per_page") != "10" {
			t.Errorf("expected per_page=10, got %s", r.URL.Query().Get("per_page"))
		}
		if r.URL.Query().Get("page") != "2" {
			t.Errorf("expected page=2, got %s", r.URL.Query().Get("page"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "abc123", "amount": 100.00},
				{"id": "def456", "amount": 200.00},
			},
			"meta": map[string]interface{}{
				"pagination": map[string]interface{}{
					"total":        50,
					"count":        2,
					"per_page":     10,
					"current_page": 2,
					"total_pages":  5,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	opts := &ExpenseListOptions{
		PerPage: 10,
		Page:    2,
	}

	resp, err := client.Expenses.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 expenses, got %d", len(resp.Data))
	}

	if resp.Data[0].ID != "abc123" {
		t.Errorf("expected first expense ID to be 'abc123', got '%s'", resp.Data[0].ID)
	}

	if resp.Meta.Pagination.Total != 50 {
		t.Errorf("expected total to be 50, got %d", resp.Meta.Pagination.Total)
	}
}

func TestExpensesServiceGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/expenses/abc123" {
			t.Errorf("expected path /api/v1/expenses/abc123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":         "abc123",
				"amount":     100.00,
				"client_id":  "client123",
				"is_deleted": false,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	expense, err := client.Expenses.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if expense.ID != "abc123" {
		t.Errorf("expected expense ID to be 'abc123', got '%s'", expense.ID)
	}

	if expense.Amount != 100.00 {
		t.Errorf("expected amount to be 100.00, got %f", expense.Amount)
	}
}

func TestExpensesServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/expenses" {
			t.Errorf("expected path /api/v1/expenses, got %s", r.URL.Path)
		}

		var body Expense
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.ClientID != "client123" {
			t.Errorf("expected client_id to be 'client123', got '%s'", body.ClientID)
		}
		if body.Amount != 150.00 {
			t.Errorf("expected amount to be 150.00, got %f", body.Amount)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":        "new123",
				"client_id": "client123",
				"amount":    150.00,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	req := &Expense{
		ClientID: "client123",
		Amount:   150.00,
		Date:     "2024-01-15",
	}

	expense, err := client.Expenses.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if expense.ID != "new123" {
		t.Errorf("expected expense ID to be 'new123', got '%s'", expense.ID)
	}
}

func TestExpensesServiceUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/expenses/abc123" {
			t.Errorf("expected path /api/v1/expenses/abc123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":            "abc123",
				"private_notes": "Updated notes",
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	req := &Expense{
		PrivateNotes: "Updated notes",
	}

	expense, err := client.Expenses.Update(context.Background(), "abc123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if expense.ID != "abc123" {
		t.Errorf("expected expense ID to be 'abc123', got '%s'", expense.ID)
	}
}

func TestExpensesServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/expenses/abc123" {
			t.Errorf("expected path /api/v1/expenses/abc123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	err := client.Expenses.Delete(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpenseListOptionsToQuery(t *testing.T) {
	isDeleted := true
	opts := &ExpenseListOptions{
		PerPage:   10,
		Page:      2,
		Filter:    "test",
		ClientID:  "client123",
		Status:    "active",
		CreatedAt: "2024-01-01",
		UpdatedAt: "2024-01-15",
		IsDeleted: &isDeleted,
		Sort:      "amount|desc",
	}

	q := opts.toQuery()

	if q.Get("per_page") != "10" {
		t.Errorf("expected per_page=10, got %s", q.Get("per_page"))
	}
	if q.Get("page") != "2" {
		t.Errorf("expected page=2, got %s", q.Get("page"))
	}
	if q.Get("filter") != "test" {
		t.Errorf("expected filter=test, got %s", q.Get("filter"))
	}
	if q.Get("is_deleted") != "true" {
		t.Errorf("expected is_deleted=true, got %s", q.Get("is_deleted"))
	}
}

func TestExpenseListOptionsNilToQuery(t *testing.T) {
	var opts *ExpenseListOptions = nil
	q := opts.toQuery()
	if q != nil {
		t.Error("expected nil query for nil options")
	}
}

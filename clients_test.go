package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientsServiceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients" {
			t.Errorf("expected path /api/v1/clients, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "client123", "name": "Acme Corp", "balance": 1000.00},
				{"id": "client456", "name": "Widgets Inc", "balance": 500.00},
			},
			"meta": map[string]interface{}{
				"pagination": map[string]interface{}{
					"total":        100,
					"count":        2,
					"per_page":     20,
					"current_page": 1,
					"total_pages":  5,
				},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	resp, err := client.Clients.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 clients, got %d", len(resp.Data))
	}

	if resp.Data[0].Name != "Acme Corp" {
		t.Errorf("expected first client name to be 'Acme Corp', got '%s'", resp.Data[0].Name)
	}
}

func TestClientsServiceGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients/client123" {
			t.Errorf("expected path /api/v1/clients/client123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":            "client123",
				"name":          "Acme Corp",
				"balance":       1000.00,
				"credit_balance": 50.00,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	c, err := client.Clients.Get(context.Background(), "client123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ID != "client123" {
		t.Errorf("expected client ID to be 'client123', got '%s'", c.ID)
	}

	if c.Balance != 1000.00 {
		t.Errorf("expected balance to be 1000.00, got %f", c.Balance)
	}
}

func TestClientsServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients" {
			t.Errorf("expected path /api/v1/clients, got %s", r.URL.Path)
		}

		var body INClient
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.Name != "New Client" {
			t.Errorf("expected name to be 'New Client', got '%s'", body.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "newclient123",
				"name": "New Client",
			},
		})
	}))
	defer server.Close()

	apiClient := NewClient("test-token", WithBaseURL(server.URL))

	req := &INClient{
		Name: "New Client",
		Contacts: []ClientContact{
			{FirstName: "John", LastName: "Doe", Email: "john@example.com", IsPrimary: true},
		},
	}

	c, err := apiClient.Clients.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ID != "newclient123" {
		t.Errorf("expected client ID to be 'newclient123', got '%s'", c.ID)
	}
}

func TestClientsServiceUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients/client123" {
			t.Errorf("expected path /api/v1/clients/client123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "client123",
				"name": "Updated Name",
			},
		})
	}))
	defer server.Close()

	apiClient := NewClient("test-token", WithBaseURL(server.URL))

	req := &INClient{
		Name: "Updated Name",
	}

	c, err := apiClient.Clients.Update(context.Background(), "client123", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Name != "Updated Name" {
		t.Errorf("expected name to be 'Updated Name', got '%s'", c.Name)
	}
}

func TestClientsServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients/client123" {
			t.Errorf("expected path /api/v1/clients/client123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiClient := NewClient("test-token", WithBaseURL(server.URL))

	err := apiClient.Clients.Delete(context.Background(), "client123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientsServicePurge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients/client123/purge" {
			t.Errorf("expected path /api/v1/clients/client123/purge, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiClient := NewClient("test-token", WithBaseURL(server.URL))

	err := apiClient.Clients.Purge(context.Background(), "client123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientsServiceMerge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients/primary123/merge456/merge" {
			t.Errorf("expected path /api/v1/clients/primary123/merge456/merge, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":   "primary123",
				"name": "Merged Client",
			},
		})
	}))
	defer server.Close()

	apiClient := NewClient("test-token", WithBaseURL(server.URL))

	c, err := apiClient.Clients.Merge(context.Background(), "primary123", "merge456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.ID != "primary123" {
		t.Errorf("expected client ID to be 'primary123', got '%s'", c.ID)
	}
}

func TestClientsServiceBulk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/clients/bulk" {
			t.Errorf("expected path /api/v1/clients/bulk, got %s", r.URL.Path)
		}

		var body BulkAction
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.Action != "archive" {
			t.Errorf("expected action to be 'archive', got '%s'", body.Action)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "client123"},
				{"id": "client456"},
			},
		})
	}))
	defer server.Close()

	apiClient := NewClient("test-token", WithBaseURL(server.URL))

	clients, err := apiClient.Clients.Bulk(context.Background(), "archive", []string{"client123", "client456"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clients) != 2 {
		t.Errorf("expected 2 clients, got %d", len(clients))
	}
}

func TestClientListOptionsToQuery(t *testing.T) {
	isDeleted := true
	opts := &ClientListOptions{
		PerPage:   15,
		Page:      2,
		Filter:    "acme",
		Balance:   "gt:1000",
		Status:    "active",
		CreatedAt: "2024-01-01",
		UpdatedAt: "2024-01-15",
		IsDeleted: &isDeleted,
		Sort:      "name|asc",
		Include:   "contacts,documents",
	}

	q := opts.toQuery()

	if q.Get("per_page") != "15" {
		t.Errorf("expected per_page=15, got %s", q.Get("per_page"))
	}
	if q.Get("page") != "2" {
		t.Errorf("expected page=2, got %s", q.Get("page"))
	}
	if q.Get("filter") != "acme" {
		t.Errorf("expected filter=acme, got %s", q.Get("filter"))
	}
	if q.Get("balance") != "gt:1000" {
		t.Errorf("expected balance=gt:1000, got %s", q.Get("balance"))
	}
	if q.Get("include") != "contacts,documents" {
		t.Errorf("expected include=contacts,documents, got %s", q.Get("include"))
	}
}

func TestClientListOptionsNilToQuery(t *testing.T) {
	var opts *ClientListOptions = nil
	q := opts.toQuery()
	if q != nil {
		t.Error("expected nil query for nil options")
	}
}

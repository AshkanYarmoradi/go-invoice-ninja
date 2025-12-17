package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaymentTermsServiceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/payment_terms" {
			t.Errorf("expected path /api/v1/payment_terms, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "term1", "name": "Net 30", "num_days": 30},
				{"id": "term2", "name": "Net 60", "num_days": 60},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	resp, err := client.PaymentTerms.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Errorf("expected 2 payment terms, got %d", len(resp.Data))
	}

	if resp.Data[0].Name != "Net 30" {
		t.Errorf("expected first term name to be 'Net 30', got '%s'", resp.Data[0].Name)
	}

	if resp.Data[0].NumDays != 30 {
		t.Errorf("expected first term num_days to be 30, got %d", resp.Data[0].NumDays)
	}
}

func TestPaymentTermsServiceGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payment_terms/term1" {
			t.Errorf("expected path /api/v1/payment_terms/term1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "term1",
				"name":     "Net 30",
				"num_days": 30,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	term, err := client.PaymentTerms.Get(context.Background(), "term1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if term.ID != "term1" {
		t.Errorf("expected term ID to be 'term1', got '%s'", term.ID)
	}
}

func TestPaymentTermsServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var body PaymentTerm
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body.Name != "Net 45" {
			t.Errorf("expected name to be 'Net 45', got '%s'", body.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "newterm",
				"name":     "Net 45",
				"num_days": 45,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	term, err := client.PaymentTerms.Create(context.Background(), &PaymentTerm{
		Name:    "Net 45",
		NumDays: 45,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if term.ID != "newterm" {
		t.Errorf("expected term ID to be 'newterm', got '%s'", term.ID)
	}
}

func TestPaymentTermsServiceUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT method, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"id":       "term1",
				"name":     "Updated Term",
				"num_days": 90,
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	term, err := client.PaymentTerms.Update(context.Background(), "term1", &PaymentTerm{
		Name:    "Updated Term",
		NumDays: 90,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if term.Name != "Updated Term" {
		t.Errorf("expected term name to be 'Updated Term', got '%s'", term.Name)
	}
}

func TestPaymentTermsServiceDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	err := client.PaymentTerms.Delete(context.Background(), "term1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPaymentTermListOptionsToQuery(t *testing.T) {
	opts := &PaymentTermListOptions{
		PerPage: 10,
		Page:    2,
		Include: "company",
	}

	q := opts.toQuery()

	if q.Get("per_page") != "10" {
		t.Errorf("expected per_page=10, got %s", q.Get("per_page"))
	}
	if q.Get("page") != "2" {
		t.Errorf("expected page=2, got %s", q.Get("page"))
	}
	if q.Get("include") != "company" {
		t.Errorf("expected include=company, got %s", q.Get("include"))
	}
}

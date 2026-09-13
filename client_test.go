package invoiceninja

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")

	if client.apiToken != "test-token" {
		t.Errorf("expected apiToken to be 'test-token', got '%s'", client.apiToken)
	}

	if client.baseURL != DefaultBaseURL {
		t.Errorf("expected baseURL to be '%s', got '%s'", DefaultBaseURL, client.baseURL)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customHTTP := &http.Client{}
	customURL := "https://custom.example.com"

	client := NewClient("test-token",
		WithHTTPClient(customHTTP),
		WithBaseURL(customURL),
	)

	if client.httpClient != customHTTP {
		t.Error("expected custom HTTP client to be set")
	}

	if client.baseURL != customURL {
		t.Errorf("expected baseURL to be '%s', got '%s'", customURL, client.baseURL)
	}
}

func TestSetBaseURL(t *testing.T) {
	client := NewClient("test-token")

	client.SetBaseURL("https://custom.example.com/")

	// Should trim trailing slash
	if client.baseURL != "https://custom.example.com" {
		t.Errorf("expected baseURL to be 'https://custom.example.com', got '%s'", client.baseURL)
	}
}

func TestWithTimeout(t *testing.T) {
	client := NewClient("test-token", WithTimeout(5*time.Second))

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout to be 5s, got %v", client.httpClient.Timeout)
	}
}

func TestWithTimeoutBeforeWithHTTPClient(t *testing.T) {
	customHTTP := &http.Client{}

	client := NewClient("test-token", WithTimeout(5*time.Second), WithHTTPClient(customHTTP))

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout to be 5s, got %v", client.httpClient.Timeout)
	}

	if customHTTP.Timeout != 0 {
		t.Errorf("expected custom HTTP client not to be modified, got timeout %v", customHTTP.Timeout)
	}
}

func TestWithTimeoutDoesNotModifyCustomHTTPClient(t *testing.T) {
	customHTTP := &http.Client{Timeout: time.Minute}

	client := NewClient("test-token", WithHTTPClient(customHTTP), WithTimeout(5*time.Second))

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout to be 5s, got %v", client.httpClient.Timeout)
	}

	if customHTTP.Timeout != time.Minute {
		t.Errorf("expected custom HTTP client timeout to stay 1m, got %v", customHTTP.Timeout)
	}
}

func TestWithHTTPClientNil(t *testing.T) {
	client := NewClient("test-token", WithHTTPClient(nil))

	if client.httpClient == nil {
		t.Fatal("expected the default HTTP client to be kept")
	}

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("expected default timeout, got %v", client.httpClient.Timeout)
	}
}

func TestClientRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("X-API-TOKEN") != "test-token" {
			t.Errorf("expected X-API-TOKEN header to be 'test-token'")
		}
		if r.Header.Get("X-Requested-With") != "XMLHttpRequest" {
			t.Errorf("expected X-Requested-With header to be 'XMLHttpRequest'")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type header to be 'application/json'")
		}

		// Return a test response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	var result map[string]string
	err := client.Request(context.Background(), "GET", "/test", nil, &result)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status to be 'ok', got '%s'", result["status"])
	}
}

func TestClientRequestError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Invalid API token"})
	}))
	defer server.Close()

	client := NewClient("invalid-token", WithBaseURL(server.URL))

	var result map[string]string
	err := client.Request(context.Background(), "GET", "/test", nil, &result)

	if err == nil {
		t.Error("expected error, got nil")
	}

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Errorf("expected APIError, got %T", err)
	}

	if !apiErr.IsUnauthorized() {
		t.Errorf("expected IsUnauthorized to be true")
	}
}

func TestClientRequestErrorHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	err := client.Request(context.Background(), "GET", "/test", nil, nil)

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.Headers.Get("Retry-After") != "30" {
		t.Errorf("expected Retry-After header to be '30', got '%s'", apiErr.Headers.Get("Retry-After"))
	}
}

func TestClientRequestNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	serverURL := server.URL
	server.Close()

	client := NewClient("test-token", WithBaseURL(serverURL))

	err := client.Request(context.Background(), "GET", "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.HasPrefix(err.Error(), "request failed: ") {
		t.Errorf("expected error to start with 'request failed: ', got '%v'", err)
	}

	if _, ok := IsAPIError(err); ok {
		t.Error("expected network error not to be an APIError")
	}
}

func TestClientRequestWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if body["name"] != "Test Client" {
			t.Errorf("expected name to be 'Test Client', got '%v'", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]string{"id": "abc123", "name": "Test Client"},
		})
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	reqBody := map[string]string{"name": "Test Client"}
	var result map[string]interface{}
	err := client.Request(context.Background(), "POST", "/test", reqBody, &result)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

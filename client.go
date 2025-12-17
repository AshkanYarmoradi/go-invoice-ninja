// Package invoiceninja provides a Go SDK for the Invoice Ninja API.
//
// This SDK supports both cloud-hosted (invoicing.co) and self-hosted Invoice Ninja instances.
// It focuses on payment-related functionality while providing a generic request method
// for accessing other API endpoints.
//
// # Authentication
//
// All requests require an API token obtained from Settings > Account Management > Integrations > API tokens.
//
// # Usage
//
//	client := invoiceninja.NewClient("your-api-token")
//	// For self-hosted instances:
//	client.SetBaseURL("https://your-instance.com")
//
//	// List payments
//	payments, err := client.Payments.List(ctx, nil)
//
// # Generic Requests
//
// For endpoints not covered by specialized methods, use the generic request:
//
//	var result json.RawMessage
//	err := client.Request(ctx, "GET", "/api/v1/activities", nil, &result)
package invoiceninja

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production Invoice Ninja cloud API endpoint.
	DefaultBaseURL = "https://invoicing.co"

	// DemoBaseURL is the demo Invoice Ninja API endpoint.
	DemoBaseURL = "https://demo.invoiceninja.com"

	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	// Version is the SDK version.
	Version = "1.0.0"
)

// Client is the Invoice Ninja API client.
type Client struct {
	// httpClient is the underlying HTTP client used for requests.
	httpClient *http.Client

	// baseURL is the API base URL.
	baseURL string

	// apiToken is the API authentication token.
	apiToken string

	// Payments provides access to payment-related endpoints.
	Payments *PaymentsService

	// Invoices provides access to invoice-related endpoints.
	Invoices *InvoicesService

	// Clients provides access to client-related endpoints.
	Clients *ClientsService
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets a custom base URL (for self-hosted instances).
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithTimeout sets a custom timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// NewClient creates a new Invoice Ninja API client.
func NewClient(apiToken string, opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:  DefaultBaseURL,
		apiToken: apiToken,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Initialize services
	c.Payments = &PaymentsService{client: c}
	c.Invoices = &InvoicesService{client: c}
	c.Clients = &ClientsService{client: c}

	return c
}

// SetBaseURL sets the API base URL. Use this for self-hosted instances.
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = strings.TrimSuffix(baseURL, "/")
}

// Request performs a generic API request.
// This method can be used to access any API endpoint not covered by specialized methods.
func (c *Client) Request(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, method, path, nil, body, result)
}

// RequestWithQuery performs a generic API request with query parameters.
func (c *Client) RequestWithQuery(ctx context.Context, method, path string, query url.Values, body interface{}, result interface{}) error {
	return c.doRequest(ctx, method, path, query, body, result)
}

// doRequest performs the actual HTTP request.
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body interface{}, result interface{}) error {
	// Build URL
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	// Prepare request body
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("X-API-TOKEN", c.apiToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "go-invoice-ninja/"+Version)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, respBody)
	}

	// Parse response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

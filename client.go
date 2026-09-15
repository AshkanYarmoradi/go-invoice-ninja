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
// # Rate Limiting and Retries
//
// Rate limiting and automatic retries are off by default. Enable them with
// WithRateLimiter and WithRetryConfig, or use NewRateLimitedClient:
//
//	client := invoiceninja.NewClient("your-api-token",
//		invoiceninja.WithRateLimiter(invoiceninja.NewRateLimiter(10)),
//		invoiceninja.WithRetryConfig(invoiceninja.DefaultRetryConfig()),
//	)
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
	"sync"
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

	// timeout is the request timeout set with WithTimeout.
	timeout time.Duration

	// hasTimeout reports whether WithTimeout was used.
	hasTimeout bool

	// mu guards rateLimiter and retryConfig.
	mu sync.RWMutex

	// rateLimiter limits outgoing requests. Nil means no rate limiting.
	rateLimiter *RateLimiter

	// retryConfig controls automatic retries. Nil means no retries.
	retryConfig *RetryConfig

	// Payments provides access to payment-related endpoints.
	Payments *PaymentsService

	// Invoices provides access to invoice-related endpoints.
	Invoices *InvoicesService

	// Expenses provides access to expense-related endpoints.
	Expenses *ExpensesService

	// Clients provides access to client-related endpoints.
	Clients *ClientsService

	// PaymentTerms provides access to payment terms endpoints.
	PaymentTerms *PaymentTermsService

	// Credits provides access to credit-related endpoints.
	Credits *CreditsService

	// Downloads provides access to file download operations.
	Downloads *DownloadsService

	// Uploads provides access to file upload operations.
	Uploads *UploadsService
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client. A nil client is ignored.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithBaseURL sets a custom base URL (for self-hosted instances).
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// WithTimeout sets a custom timeout for HTTP requests.
//
// The timeout is applied after all other options, so the order of WithTimeout and
// WithHTTPClient does not matter. An *http.Client passed to WithHTTPClient is not
// modified; the client uses a copy of it with the new timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
		c.hasTimeout = true
	}
}

// WithRateLimiter limits how many requests per second the client sends.
// Every service method and generic request waits for the limiter.
func WithRateLimiter(limiter *RateLimiter) ClientOption {
	return func(c *Client) {
		c.rateLimiter = limiter
	}
}

// WithRetryConfig enables automatic retries of failed requests.
// See RetryConfig for which requests are retried.
func WithRetryConfig(config *RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// NewClient creates a new Invoice Ninja API client.
func NewClient(apiToken string, opts ...ClientOption) *Client {
	defaultHTTPClient := &http.Client{
		Timeout: DefaultTimeout,
	}

	c := &Client{
		httpClient: defaultHTTPClient,
		baseURL:    DefaultBaseURL,
		apiToken:   apiToken,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.hasTimeout {
		if c.httpClient != defaultHTTPClient {
			// Copy the caller's HTTP client so its timeout is not changed.
			httpClient := *c.httpClient
			c.httpClient = &httpClient
		}
		c.httpClient.Timeout = c.timeout
	}

	// Initialize services
	c.Payments = &PaymentsService{client: c}
	c.Invoices = &InvoicesService{client: c}
	c.Expenses = &ExpensesService{client: c}
	c.Clients = &ClientsService{client: c}
	c.PaymentTerms = &PaymentTermsService{client: c}
	c.Credits = &CreditsService{client: c}
	c.Downloads = &DownloadsService{client: c}
	c.Uploads = &UploadsService{client: c}

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

// jsonMediaType is the media type of JSON request and response bodies.
const jsonMediaType = "application/json"

// apiRequest describes an HTTP request. The body is kept as bytes so the request can be retried.
type apiRequest struct {
	method      string
	url         string
	body        []byte
	contentType string
	accept      string
}

// doRequest performs a JSON API request.
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, body, result interface{}) error {
	// Build URL
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if query != nil {
		u.RawQuery = query.Encode()
	}

	req := &apiRequest{
		method:      method,
		url:         u.String(),
		contentType: jsonMediaType,
		accept:      jsonMediaType,
	}

	// Prepare request body
	if body != nil {
		jsonBody, marshalErr := json.Marshal(body)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal request body: %w", marshalErr)
		}
		req.body = jsonBody
	}

	respBody, err := c.execute(ctx, req)
	if err != nil {
		return err
	}

	// Parse response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// execute sends req and returns the response body. Before each attempt it waits for
// the client's rate limiter, and it retries failed attempts according to the client's
// retry configuration.
func (c *Client) execute(ctx context.Context, req *apiRequest) ([]byte, error) {
	limiter, retryConfig := c.requestPolicy()

	for attempt := 0; ; attempt++ {
		if limiter != nil {
			if waitErr := limiter.Wait(ctx); waitErr != nil {
				return nil, waitErr
			}
		}

		respBody, err := c.send(ctx, req)
		if err == nil {
			return respBody, nil
		}

		if ctx.Err() != nil || !retryConfig.shouldRetry(req.method, err, attempt) {
			return nil, err
		}

		backoff, ok := retryConfig.calculateBackoff(attempt, err)
		if !ok {
			return nil, err
		}

		if sleepErr := sleepContext(ctx, backoff); sleepErr != nil {
			return nil, sleepErr
		}
	}
}

// send performs a single HTTP request and returns the response body.
func (c *Client) send(ctx context.Context, req *apiRequest) ([]byte, error) {
	var bodyReader io.Reader
	if req.body != nil {
		bodyReader = bytes.NewReader(req.body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.method, req.url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("X-API-TOKEN", c.apiToken)
	httpReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	if req.contentType != "" {
		httpReq.Header.Set("Content-Type", req.contentType)
	}
	httpReq.Header.Set("Accept", req.accept)
	httpReq.Header.Set("User-Agent", "go-invoice-ninja/"+Version)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &networkError{op: "request failed", err: err}
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &networkError{op: "failed to read response body", err: err}
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		apiErr := parseAPIError(resp.StatusCode, respBody)
		apiErr.Headers = resp.Header
		return nil, apiErr
	}

	return respBody, nil
}

// requestPolicy returns the rate limiter and retry configuration used for requests.
func (c *Client) requestPolicy() (*RateLimiter, *RetryConfig) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.rateLimiter, c.retryConfig
}

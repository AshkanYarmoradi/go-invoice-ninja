package invoiceninja

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fastRetryConfig returns a retry configuration with short backoffs for tests.
func fastRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:         3,
		InitialBackoff:     time.Millisecond,
		MaxBackoff:         10 * time.Millisecond,
		BackoffMultiplier:  2.0,
		RetryOnStatusCodes: []int{429, 500, 502, 503, 504},
	}
}

// newFlakyServer starts a server that responds with failStatus to the first failures
// requests and with successBody afterwards. It returns a pointer to the request count.
func newFlakyServer(t *testing.T, failures int32, failStatus int, successBody string) (server *httptest.Server, calls *int32) {
	t.Helper()

	var count int32
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) <= failures {
			w.WriteHeader(failStatus)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(successBody))
	}))
	t.Cleanup(server.Close)

	return server, &count
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.InitialBackoff != 1*time.Second {
		t.Errorf("expected InitialBackoff=1s, got %v", config.InitialBackoff)
	}

	if config.MaxBackoff != 30*time.Second {
		t.Errorf("expected MaxBackoff=30s, got %v", config.MaxBackoff)
	}

	if config.BackoffMultiplier != 2.0 {
		t.Errorf("expected BackoffMultiplier=2.0, got %f", config.BackoffMultiplier)
	}

	if !config.Jitter {
		t.Error("expected Jitter=true")
	}

	if config.RetryNonIdempotent {
		t.Error("expected RetryNonIdempotent=false")
	}

	expectedCodes := []int{429, 500, 502, 503, 504}
	if len(config.RetryOnStatusCodes) != len(expectedCodes) {
		t.Errorf("expected %d retry status codes, got %d", len(expectedCodes), len(config.RetryOnStatusCodes))
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(5) // 5 requests per second

	ctx := context.Background()
	start := time.Now()

	// Make 5 requests quickly - should all succeed immediately
	for i := 0; i < 5; i++ {
		if err := limiter.Wait(ctx); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("first 5 requests should be immediate, took %v", elapsed)
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	limiter := NewRateLimiter(10)
	ctx := context.Background()

	var wg sync.WaitGroup
	var count int32

	// Launch 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := limiter.Wait(ctx); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			atomic.AddInt32(&count, 1)
		}()
	}

	wg.Wait()

	if count != 10 {
		t.Errorf("expected 10 requests to complete, got %d", count)
	}
}

func TestRateLimiterContextCancellation(t *testing.T) {
	limiter := NewRateLimiter(1)

	ctx := context.Background()

	// Use up the rate limit
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cancel context immediately
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// This should return immediately with context error
	err := limiter.Wait(cancelCtx)
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func TestRateLimiterDisabled(t *testing.T) {
	for _, limit := range []int{0, -1} {
		limiter := NewRateLimiter(limit)

		for i := 0; i < 100; i++ {
			if err := limiter.Wait(context.Background()); err != nil {
				t.Fatalf("NewRateLimiter(%d): unexpected error: %v", limit, err)
			}
		}
	}
}

func TestRateLimiterCanceledContextDoesNotUseSlot(t *testing.T) {
	limiter := NewRateLimiter(1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := limiter.Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// The canceled call must not have used the only slot
	start := time.Now()
	if err := limiter.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("expected a free slot, waited %v", elapsed)
	}
}

func TestParseRetryAfterTooLarge(t *testing.T) {
	const maxDuration = time.Duration(1<<63 - 1)
	now := time.Now()

	for _, value := range []string{"9223372037", "99999999999999999999"} {
		wait, ok := parseRetryAfter(value, now)
		if !ok || wait != maxDuration {
			t.Errorf("parseRetryAfter(%q) = %v, %v, want the maximum duration", value, wait, ok)
		}
	}

	if _, ok := parseRetryAfter("-99999999999999999999", now); ok {
		t.Error("expected a negative out-of-range value to be invalid")
	}

	// A delay too long to represent stops retrying instead of retrying immediately
	headers := http.Header{}
	headers.Set("Retry-After", "9223372037")
	if _, ok := DefaultRetryConfig().calculateBackoff(0, &APIError{StatusCode: 429, Headers: headers}); ok {
		t.Error("expected no retry when Retry-After is longer than MaxBackoff")
	}
}

func TestNewRateLimitedClient(t *testing.T) {
	client := NewRateLimitedClient("test-token")

	if client.Client == nil {
		t.Error("expected embedded Client to be initialized")
	}

	if client.rateLimiter == nil {
		t.Error("expected rateLimiter to be initialized")
	}

	if client.retryConfig == nil {
		t.Error("expected retryConfig to be initialized")
	}
}

func TestNewRateLimitedClientOptionsOverrideDefaults(t *testing.T) {
	config := fastRetryConfig()
	limiter := NewRateLimiter(3)

	client := NewRateLimitedClient("test-token", WithRetryConfig(config), WithRateLimiter(limiter))

	if client.retryConfig != config {
		t.Error("expected WithRetryConfig to override the default retry config")
	}

	if client.rateLimiter != limiter {
		t.Error("expected WithRateLimiter to override the default rate limiter")
	}
}

func TestRateLimitedClientSetRateLimit(t *testing.T) {
	client := NewRateLimitedClient("test-token")

	client.SetRateLimit(20)

	if client.rateLimiter.requestsLimit != 20 {
		t.Errorf("expected rate limit 20, got %d", client.rateLimiter.requestsLimit)
	}
}

func TestRateLimitedClientSetRetryConfig(t *testing.T) {
	client := NewRateLimitedClient("test-token")

	customConfig := &RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 2 * time.Second,
		MaxBackoff:     60 * time.Second,
	}

	client.SetRetryConfig(customConfig)

	if client.retryConfig.MaxRetries != 5 {
		t.Errorf("expected MaxRetries=5, got %d", client.retryConfig.MaxRetries)
	}
}

func TestParseRateLimitHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("X-RateLimit-Limit", "100")
	headers.Set("X-RateLimit-Remaining", "95")

	info := ParseRateLimitHeaders(headers)

	if info.Limit != 100 {
		t.Errorf("expected Limit=100, got %d", info.Limit)
	}

	if info.Remaining != 95 {
		t.Errorf("expected Remaining=95, got %d", info.Remaining)
	}
}

func TestParseRateLimitHeadersEmpty(t *testing.T) {
	headers := http.Header{}

	info := ParseRateLimitHeaders(headers)

	if info.Limit != 0 {
		t.Errorf("expected Limit=0 for empty headers, got %d", info.Limit)
	}

	if info.Remaining != 0 {
		t.Errorf("expected Remaining=0 for empty headers, got %d", info.Remaining)
	}
}

func TestShouldRetry(t *testing.T) {
	config := DefaultRetryConfig()
	netErr := &networkError{op: "request failed", err: errors.New("connection reset")}

	tests := []struct {
		name     string
		method   string
		err      error
		attempt  int
		expected bool
	}{
		{
			name:     "rate limited error",
			method:   http.MethodGet,
			err:      &APIError{StatusCode: 429},
			attempt:  0,
			expected: true,
		},
		{
			name:     "server error 500",
			method:   http.MethodGet,
			err:      &APIError{StatusCode: 500},
			attempt:  0,
			expected: true,
		},
		{
			name:     "server error 503",
			method:   http.MethodGet,
			err:      &APIError{StatusCode: 503},
			attempt:  0,
			expected: true,
		},
		{
			name:     "client error 400",
			method:   http.MethodGet,
			err:      &APIError{StatusCode: 400},
			attempt:  0,
			expected: false,
		},
		{
			name:     "not found error",
			method:   http.MethodGet,
			err:      &APIError{StatusCode: 404},
			attempt:  0,
			expected: false,
		},
		{
			name:     "max retries exceeded",
			method:   http.MethodGet,
			err:      &APIError{StatusCode: 500},
			attempt:  3,
			expected: false,
		},
		{
			name:     "network error on GET",
			method:   http.MethodGet,
			err:      netErr,
			attempt:  0,
			expected: true,
		},
		{
			name:     "network error on POST",
			method:   http.MethodPost,
			err:      netErr,
			attempt:  0,
			expected: false,
		},
		{
			name:     "server error on POST",
			method:   http.MethodPost,
			err:      &APIError{StatusCode: 503},
			attempt:  0,
			expected: false,
		},
		{
			name:     "rate limited POST",
			method:   http.MethodPost,
			err:      &APIError{StatusCode: 429},
			attempt:  0,
			expected: true,
		},
		{
			name:     "server error on PUT",
			method:   http.MethodPut,
			err:      &APIError{StatusCode: 502},
			attempt:  0,
			expected: true,
		},
		{
			name:     "other error",
			method:   http.MethodGet,
			err:      errors.New("failed to unmarshal response"),
			attempt:  0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.shouldRetry(tt.method, tt.err, tt.attempt)
			if result != tt.expected {
				t.Errorf("shouldRetry() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldRetryNonIdempotentEnabled(t *testing.T) {
	config := DefaultRetryConfig()
	config.RetryNonIdempotent = true

	if !config.shouldRetry(http.MethodPost, &APIError{StatusCode: 503}, 0) {
		t.Error("expected POST to be retried when RetryNonIdempotent is set")
	}
}

func TestShouldRetryNilConfig(t *testing.T) {
	var config *RetryConfig

	if config.shouldRetry(http.MethodGet, &APIError{StatusCode: 503}, 0) {
		t.Error("expected no retry without a retry config")
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := DefaultRetryConfig()
	config.Jitter = false // Disable jitter for predictable tests

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first attempt",
			attempt:  0,
			expected: 1 * time.Second,
		},
		{
			name:     "second attempt",
			attempt:  1,
			expected: 2 * time.Second,
		},
		{
			name:     "third attempt",
			attempt:  2,
			expected: 4 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := config.calculateBackoff(tt.attempt, &APIError{StatusCode: 500})
			if !ok || result != tt.expected {
				t.Errorf("calculateBackoff() = %v, %v, want %v, true", result, ok, tt.expected)
			}
		})
	}
}

func TestCalculateBackoffMaxCap(t *testing.T) {
	config := DefaultRetryConfig()
	config.Jitter = false
	config.MaxBackoff = 5 * time.Second

	// After several attempts, backoff should be capped
	backoff, ok := config.calculateBackoff(10, &APIError{StatusCode: 500})
	if !ok || backoff != config.MaxBackoff {
		t.Errorf("calculateBackoff() = %v, %v, want %v, true", backoff, ok, config.MaxBackoff)
	}
}

func TestCalculateBackoffZeroMultiplier(t *testing.T) {
	config := &RetryConfig{InitialBackoff: time.Second}

	for attempt := 0; attempt < 3; attempt++ {
		backoff, ok := config.calculateBackoff(attempt, &APIError{StatusCode: 500})
		if !ok || backoff != time.Second {
			t.Errorf("attempt %d: calculateBackoff() = %v, %v, want 1s, true", attempt, backoff, ok)
		}
	}
}

func TestCalculateBackoffRateLimited(t *testing.T) {
	config := DefaultRetryConfig()
	config.Jitter = false

	// Without Retry-After, rate limited errors use exponential backoff
	backoff, ok := config.calculateBackoff(1, &APIError{StatusCode: 429})
	if !ok || backoff != 2*time.Second {
		t.Errorf("expected 2s backoff without Retry-After, got %v, %v", backoff, ok)
	}

	// Retry-After is honored
	headers := http.Header{}
	headers.Set("Retry-After", "7")
	backoff, ok = config.calculateBackoff(0, &APIError{StatusCode: 429, Headers: headers})
	if !ok || backoff != 7*time.Second {
		t.Errorf("expected 7s backoff from Retry-After, got %v, %v", backoff, ok)
	}

	// A Retry-After longer than MaxBackoff stops retrying
	headers.Set("Retry-After", "120")
	if _, ok = config.calculateBackoff(0, &APIError{StatusCode: 429, Headers: headers}); ok {
		t.Error("expected no retry when Retry-After exceeds MaxBackoff")
	}
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		name     string
		value    string
		expected time.Duration
		ok       bool
	}{
		{name: "empty", value: "", expected: 0, ok: false},
		{name: "seconds", value: "5", expected: 5 * time.Second, ok: true},
		{name: "negative seconds", value: "-1", expected: 0, ok: false},
		{name: "invalid", value: "soon", expected: 0, ok: false},
		{name: "future date", value: now.Add(10 * time.Second).Format(http.TimeFormat), expected: 10 * time.Second, ok: true},
		{name: "past date", value: now.Add(-time.Minute).Format(http.TimeFormat), expected: 0, ok: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseRetryAfter(tt.value, now)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("parseRetryAfter(%q) = %v, %v, want %v, %v", tt.value, result, ok, tt.expected, tt.ok)
			}
		})
	}
}

func TestRateLimitedClientRetriesServiceMethods(t *testing.T) {
	server, calls := newFlakyServer(t, 1, http.StatusServiceUnavailable, `{"data":[{"id":"abc123"}]}`)

	client := NewRateLimitedClient("test-token", WithBaseURL(server.URL))
	client.SetRetryConfig(fastRetryConfig())

	resp, err := client.Payments.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Errorf("expected 1 payment, got %d", len(resp.Data))
	}

	if got := atomic.LoadInt32(calls); got != 2 {
		t.Errorf("expected 2 requests, got %d", got)
	}
}

func TestClientWithRetryConfigRetriesRequests(t *testing.T) {
	server, calls := newFlakyServer(t, 2, http.StatusBadGateway, `{"status":"ok"}`)

	client := NewClient("test-token", WithBaseURL(server.URL), WithRetryConfig(fastRetryConfig()))

	var result map[string]string
	if err := client.Request(context.Background(), http.MethodGet, "/test", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status to be 'ok', got '%s'", result["status"])
	}

	if got := atomic.LoadInt32(calls); got != 3 {
		t.Errorf("expected 3 requests, got %d", got)
	}
}

func TestClientWithoutRetryConfigDoesNotRetry(t *testing.T) {
	server, calls := newFlakyServer(t, 1, http.StatusServiceUnavailable, `{}`)

	client := NewClient("test-token", WithBaseURL(server.URL))

	err := client.Request(context.Background(), http.MethodGet, "/test", nil, nil)
	apiErr, ok := IsAPIError(err)
	if !ok || apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 APIError, got %v", err)
	}

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Errorf("expected 1 request, got %d", got)
	}
}

func TestRetryDoesNotRepeatPostAfterServerError(t *testing.T) {
	server, calls := newFlakyServer(t, 1, http.StatusServiceUnavailable, `{"data":{"id":"new123"}}`)

	client := NewRateLimitedClient("test-token", WithBaseURL(server.URL))
	client.SetRetryConfig(fastRetryConfig())

	_, err := client.Payments.Create(context.Background(), &PaymentRequest{ClientID: "client123", Amount: 10})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Errorf("expected POST to be sent once, got %d requests", got)
	}
}

func TestRetryRepeatsPostAfterRateLimit(t *testing.T) {
	server, calls := newFlakyServer(t, 1, http.StatusTooManyRequests, `{"data":{"id":"new123"}}`)

	client := NewRateLimitedClient("test-token", WithBaseURL(server.URL))
	client.SetRetryConfig(fastRetryConfig())

	payment, err := client.Payments.Create(context.Background(), &PaymentRequest{ClientID: "client123", Amount: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if payment.ID != "new123" {
		t.Errorf("expected payment ID to be 'new123', got '%s'", payment.ID)
	}

	if got := atomic.LoadInt32(calls); got != 2 {
		t.Errorf("expected 2 requests, got %d", got)
	}
}

func TestRetryStopsWhenRetryAfterExceedsMaxBackoff(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL), WithRetryConfig(fastRetryConfig()))

	err := client.Request(context.Background(), http.MethodGet, "/test", nil, nil)
	apiErr, ok := IsAPIError(err)
	if !ok || !apiErr.IsRateLimited() {
		t.Fatalf("expected rate limit error, got %v", err)
	}

	if apiErr.Headers.Get("Retry-After") != "120" {
		t.Errorf("expected Retry-After header on error, got %q", apiErr.Headers.Get("Retry-After"))
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 request, got %d", got)
	}
}

func TestRetryReplaysRequestBody(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	var calls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		mu.Unlock()

		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":"inv123"}}`))
	}))
	defer server.Close()

	client := NewRateLimitedClient("test-token", WithBaseURL(server.URL))
	client.SetRetryConfig(fastRetryConfig())

	if _, err := client.Invoices.Update(context.Background(), "inv123", &Invoice{PONumber: "PO-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(bodies) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(bodies))
	}

	if bodies[0] == "" || bodies[0] != bodies[1] {
		t.Errorf("expected the same body on both attempts, got %q and %q", bodies[0], bodies[1])
	}
}

func TestRetryStopsWhenContextIsDone(t *testing.T) {
	server, calls := newFlakyServer(t, 100, http.StatusServiceUnavailable, `{}`)

	config := fastRetryConfig()
	config.InitialBackoff = time.Hour
	config.MaxBackoff = time.Hour

	client := NewClient("test-token", WithBaseURL(server.URL), WithRetryConfig(config))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := client.Request(ctx, http.MethodGet, "/test", nil, nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}

	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("expected request to stop when the context is done, took %v", elapsed)
	}

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Errorf("expected 1 request, got %d", got)
	}
}

func TestClientWithRateLimiterLimitsRequests(t *testing.T) {
	server, calls := newFlakyServer(t, 0, http.StatusOK, `{"data":[]}`)

	client := NewClient("test-token", WithBaseURL(server.URL), WithRateLimiter(NewRateLimiter(1)))

	if _, err := client.Payments.List(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The limit is used up, so the next request has to wait for the limiter
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.Payments.List(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Errorf("expected 1 request to reach the server, got %d", got)
	}
}

func TestDownloadsRetryServerErrors(t *testing.T) {
	server, calls := newFlakyServer(t, 1, http.StatusServiceUnavailable, "%PDF-1.4")

	client := NewClient("test-token", WithBaseURL(server.URL), WithRetryConfig(fastRetryConfig()))

	pdf, err := client.Downloads.DownloadInvoicePDF(context.Background(), "inv-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(pdf) != "%PDF-1.4" {
		t.Errorf("expected PDF content, got %q", pdf)
	}

	if got := atomic.LoadInt32(calls); got != 2 {
		t.Errorf("expected 2 requests, got %d", got)
	}
}

func TestDoRequestWithRetrySendsQuery(t *testing.T) {
	var mu sync.Mutex
	var queries []url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		queries = append(queries, r.URL.Query())
		mu.Unlock()
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewRateLimitedClient("test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	for _, query := range []interface{}{
		url.Values{"per_page": {"5"}},
		map[string][]string{"per_page": {"5"}},
		map[string]string{"per_page": "5"},
	} {
		if err := client.DoRequestWithRetry(ctx, http.MethodGet, "/api/v1/payments", query, nil, nil); err != nil {
			t.Fatalf("unexpected error for %T: %v", query, err)
		}
	}

	mu.Lock()
	if len(queries) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(queries))
	}
	for i, query := range queries {
		if query.Get("per_page") != "5" {
			t.Errorf("request %d: expected per_page=5, got %q", i, query.Get("per_page"))
		}
	}
	mu.Unlock()

	if err := client.DoRequestWithRetry(ctx, http.MethodGet, "/api/v1/payments", 42, nil, nil); err == nil {
		t.Error("expected error for unsupported query type")
	}
}

func TestDoRequestWithRetryNegativeMaxRetries(t *testing.T) {
	server, calls := newFlakyServer(t, 100, http.StatusInternalServerError, `{}`)

	client := NewRateLimitedClient("test-token", WithBaseURL(server.URL))
	client.SetRetryConfig(&RetryConfig{MaxRetries: -1})

	if err := client.DoRequestWithRetry(context.Background(), http.MethodGet, "/test", nil, nil, nil); err == nil {
		t.Error("expected error, got nil")
	}

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Errorf("expected exactly 1 request, got %d", got)
	}
}

func TestRateLimitedClientSetRetryConfigNil(t *testing.T) {
	server, calls := newFlakyServer(t, 1, http.StatusServiceUnavailable, `{}`)

	client := NewRateLimitedClient("test-token", WithBaseURL(server.URL))
	client.SetRetryConfig(nil)

	if err := client.Request(context.Background(), http.MethodGet, "/test", nil, nil); err == nil {
		t.Error("expected error, got nil")
	}

	if got := atomic.LoadInt32(calls); got != 1 {
		t.Errorf("expected 1 request, got %d", got)
	}
}

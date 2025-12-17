package invoiceninja

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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
	client := NewRateLimitedClient("test-token")

	tests := []struct {
		name     string
		err      error
		attempt  int
		expected bool
	}{
		{
			name:     "rate limited error",
			err:      &APIError{StatusCode: 429},
			attempt:  0,
			expected: true,
		},
		{
			name:     "server error 500",
			err:      &APIError{StatusCode: 500},
			attempt:  0,
			expected: true,
		},
		{
			name:     "server error 503",
			err:      &APIError{StatusCode: 503},
			attempt:  0,
			expected: true,
		},
		{
			name:     "client error 400",
			err:      &APIError{StatusCode: 400},
			attempt:  0,
			expected: false,
		},
		{
			name:     "not found error",
			err:      &APIError{StatusCode: 404},
			attempt:  0,
			expected: false,
		},
		{
			name:     "max retries exceeded",
			err:      &APIError{StatusCode: 500},
			attempt:  3,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.shouldRetry(tt.err, tt.attempt)
			if result != tt.expected {
				t.Errorf("shouldRetry() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	client := NewRateLimitedClient("test-token")
	client.retryConfig.Jitter = false // Disable jitter for predictable tests

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
			result := client.calculateBackoff(tt.attempt, &APIError{StatusCode: 500})
			if result != tt.expected {
				t.Errorf("calculateBackoff() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateBackoffMaxCap(t *testing.T) {
	client := NewRateLimitedClient("test-token")
	client.retryConfig.Jitter = false
	client.retryConfig.MaxBackoff = 5 * time.Second

	// After several attempts, backoff should be capped
	backoff := client.calculateBackoff(10, &APIError{StatusCode: 500})
	if backoff > client.retryConfig.MaxBackoff {
		t.Errorf("backoff %v exceeded max %v", backoff, client.retryConfig.MaxBackoff)
	}
}

func TestCalculateBackoffRateLimited(t *testing.T) {
	client := NewRateLimitedClient("test-token")

	// Rate limited errors should have longer backoff
	backoff := client.calculateBackoff(0, &APIError{StatusCode: 429})
	if backoff != 60*time.Second {
		t.Errorf("expected 60s backoff for rate limited, got %v", backoff)
	}
}

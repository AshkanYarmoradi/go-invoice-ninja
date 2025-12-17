package invoiceninja

import (
	"context"
	"crypto/rand"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RetryConfig configures retry behavior for API requests.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int

	// InitialBackoff is the initial backoff duration.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration

	// BackoffMultiplier is the multiplier applied to backoff after each retry.
	BackoffMultiplier float64

	// RetryOnStatusCodes specifies which HTTP status codes should trigger a retry.
	RetryOnStatusCodes []int

	// Jitter adds randomness to backoff to prevent thundering herd.
	Jitter bool
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:         3,
		InitialBackoff:     1 * time.Second,
		MaxBackoff:         30 * time.Second,
		BackoffMultiplier:  2.0,
		RetryOnStatusCodes: []int{429, 500, 502, 503, 504},
		Jitter:             true,
	}
}

// RateLimiter implements client-side rate limiting.
type RateLimiter struct {
	mu            sync.Mutex
	requestsLimit int
	windowSize    time.Duration
	requests      []time.Time
}

// NewRateLimiter creates a new rate limiter.
// requestsPerSecond specifies the maximum requests per second allowed.
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		requestsLimit: requestsPerSecond,
		windowSize:    time.Second,
		requests:      make([]time.Time, 0, requestsPerSecond),
	}
}

// Wait blocks until a request is allowed under the rate limit.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()

		now := time.Now()

		// Remove expired requests from the window
		cutoff := now.Add(-r.windowSize)
		validRequests := make([]time.Time, 0, len(r.requests))
		for _, t := range r.requests {
			if t.After(cutoff) {
				validRequests = append(validRequests, t)
			}
		}
		r.requests = validRequests

		// Check if we're at the limit
		if len(r.requests) >= r.requestsLimit {
			// Calculate wait time until the oldest request expires
			oldestRequest := r.requests[0]
			waitTime := oldestRequest.Add(r.windowSize).Sub(now)
			r.mu.Unlock()

			if waitTime > 0 {
				select {
				case <-time.After(waitTime):
					// Retry the loop
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			continue
		}

		// Record this request and return
		r.requests = append(r.requests, time.Now())
		r.mu.Unlock()
		return nil
	}
}

// RateLimitedClient wraps a Client with rate limiting and retry logic.
type RateLimitedClient struct {
	*Client
	rateLimiter *RateLimiter
	retryConfig *RetryConfig
}

// NewRateLimitedClient creates a new client with rate limiting and retry logic.
func NewRateLimitedClient(apiToken string, opts ...ClientOption) *RateLimitedClient {
	client := NewClient(apiToken, opts...)
	return &RateLimitedClient{
		Client:      client,
		rateLimiter: NewRateLimiter(10), // Default: 10 requests per second
		retryConfig: DefaultRetryConfig(),
	}
}

// SetRateLimit sets the rate limit for API requests.
func (c *RateLimitedClient) SetRateLimit(requestsPerSecond int) {
	c.rateLimiter = NewRateLimiter(requestsPerSecond)
}

// SetRetryConfig sets the retry configuration.
func (c *RateLimitedClient) SetRetryConfig(config *RetryConfig) {
	c.retryConfig = config
}

// DoRequestWithRetry performs a request with rate limiting and retry logic.
// This method provides automatic retries with exponential backoff for transient errors.
func (c *RateLimitedClient) DoRequestWithRetry(ctx context.Context, method, path string, query, body, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Wait for rate limit
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return err
		}

		// Make the request
		err := c.Client.doRequest(ctx, method, path, nil, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if we should retry
		if !c.shouldRetry(err, attempt) {
			return err
		}

		// Calculate backoff
		backoff := c.calculateBackoff(attempt, err)

		// Wait before retrying
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// shouldRetry determines if a request should be retried.
func (c *RateLimitedClient) shouldRetry(err error, attempt int) bool {
	if attempt >= c.retryConfig.MaxRetries {
		return false
	}

	apiErr, ok := IsAPIError(err)
	if !ok {
		// Network errors should be retried
		return true
	}

	// Check if status code is in retry list
	for _, code := range c.retryConfig.RetryOnStatusCodes {
		if apiErr.StatusCode == code {
			return true
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration for a retry attempt.
func (c *RateLimitedClient) calculateBackoff(attempt int, err error) time.Duration {
	// Check for Retry-After header hint
	if apiErr, ok := IsAPIError(err); ok && apiErr.StatusCode == http.StatusTooManyRequests {
		// In a real implementation, we'd parse the Retry-After header
		// For now, use a reasonable default for rate limiting
		return 60 * time.Second
	}

	// Exponential backoff
	backoff := float64(c.retryConfig.InitialBackoff) * math.Pow(c.retryConfig.BackoffMultiplier, float64(attempt))

	// Apply jitter
	if c.retryConfig.Jitter {
		// Use crypto/rand for secure random number generation
		randInt, randErr := rand.Int(rand.Reader, big.NewInt(1000))
		if randErr == nil {
			jitter := (float64(randInt.Int64()) / 1000.0) * 0.3 * backoff // Up to 30% jitter
			backoff += jitter
		}
	}

	// Cap at max backoff
	if backoff > float64(c.retryConfig.MaxBackoff) {
		backoff = float64(c.retryConfig.MaxBackoff)
	}

	return time.Duration(backoff)
}

// RateLimitInfo contains rate limit information from API response headers.
type RateLimitInfo struct {
	// Limit is the maximum number of requests allowed per window.
	Limit int

	// Remaining is the number of requests remaining in the current window.
	Remaining int

	// Reset is the time when the rate limit window resets.
	Reset time.Time
}

// ParseRateLimitHeaders parses rate limit information from HTTP response headers.
func ParseRateLimitHeaders(headers http.Header) *RateLimitInfo {
	info := &RateLimitInfo{}

	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		info.Limit, _ = strconv.Atoi(limit)
	}

	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		info.Remaining, _ = strconv.Atoi(remaining)
	}

	// Note: Invoice Ninja may not provide a reset timestamp
	// This is a placeholder for when it does
	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.Reset = time.Unix(timestamp, 0)
		}
	}

	return info
}

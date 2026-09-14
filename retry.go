package invoiceninja

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// defaultRequestsPerSecond is the rate limit used by NewRateLimitedClient.
	defaultRequestsPerSecond = 10

	// maxJitterFraction is the largest fraction of the backoff added as random jitter.
	maxJitterFraction = 0.3
)

// RetryConfig configures automatic retries for API requests.
//
// Requests that fail with a network error, or with a status code listed in
// RetryOnStatusCodes, are retried up to MaxRetries times with exponential backoff.
// POST and PATCH requests, such as creating a payment or emailing an invoice, are
// only retried after a 429 response unless RetryNonIdempotent is set: after a network
// or server error the server may already have processed them.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts. Zero or less disables retries.
	MaxRetries int

	// InitialBackoff is the initial backoff duration.
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration. Zero means no limit.
	// If the Retry-After header of a 429 response asks for a longer wait,
	// the error is returned instead of retrying.
	MaxBackoff time.Duration

	// BackoffMultiplier is the multiplier applied to backoff after each retry.
	// Values of zero or less are treated as 1.
	BackoffMultiplier float64

	// RetryOnStatusCodes specifies which HTTP status codes should trigger a retry.
	RetryOnStatusCodes []int

	// Jitter adds randomness to backoff to prevent thundering herd.
	Jitter bool

	// RetryNonIdempotent also retries POST and PATCH requests after network errors
	// and server errors. Only enable it if repeating those requests is safe.
	RetryNonIdempotent bool
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

// shouldRetry reports whether a request with the given method that failed with err
// on the given attempt (counting from 0) should be retried.
func (rc *RetryConfig) shouldRetry(method string, err error, attempt int) bool {
	if rc == nil || attempt >= rc.MaxRetries {
		return false
	}

	var netErr *networkError
	if errors.As(err, &netErr) {
		return rc.RetryNonIdempotent || isIdempotent(method)
	}

	apiErr, ok := IsAPIError(err)
	if !ok || !rc.retriesStatusCode(apiErr.StatusCode) {
		return false
	}

	// A 429 response means the request was rejected without being processed.
	if apiErr.StatusCode == http.StatusTooManyRequests {
		return true
	}

	return rc.RetryNonIdempotent || isIdempotent(method)
}

// retriesStatusCode reports whether statusCode is listed in RetryOnStatusCodes.
func (rc *RetryConfig) retriesStatusCode(statusCode int) bool {
	for _, code := range rc.RetryOnStatusCodes {
		if code == statusCode {
			return true
		}
	}

	return false
}

// calculateBackoff returns how long to wait before retrying a request that failed with
// err on the given attempt. It returns false if the server asked for a longer wait
// than MaxBackoff.
func (rc *RetryConfig) calculateBackoff(attempt int, err error) (time.Duration, bool) {
	// Honor the Retry-After header of 429 responses
	if apiErr, ok := IsAPIError(err); ok && apiErr.StatusCode == http.StatusTooManyRequests {
		if wait, found := parseRetryAfter(apiErr.Headers.Get("Retry-After"), time.Now()); found {
			if rc.MaxBackoff > 0 && wait > rc.MaxBackoff {
				return 0, false
			}
			return wait, true
		}
	}

	multiplier := rc.BackoffMultiplier
	if multiplier <= 0 {
		multiplier = 1
	}

	// Exponential backoff
	backoff := float64(rc.InitialBackoff) * math.Pow(multiplier, float64(attempt))

	// Apply jitter
	if rc.Jitter {
		// Use crypto/rand for secure random number generation
		randInt, randErr := rand.Int(rand.Reader, big.NewInt(1000))
		if randErr == nil {
			jitter := (float64(randInt.Int64()) / 1000.0) * maxJitterFraction * backoff
			backoff += jitter
		}
	}

	// Cap at max backoff
	if rc.MaxBackoff > 0 && backoff > float64(rc.MaxBackoff) {
		backoff = float64(rc.MaxBackoff)
	}

	return time.Duration(backoff), true
}

// isIdempotent reports whether a request with this HTTP method can safely be repeated.
func isIdempotent(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// maxRetryAfterSeconds is the largest number of seconds that fits in a time.Duration.
const maxRetryAfterSeconds = math.MaxInt64 / int64(time.Second)

// parseRetryAfter parses a Retry-After header value, which is either a number of
// seconds or an HTTP date. A delay too long for a time.Duration is returned as the
// longest possible duration, so it is never mistaken for a short or negative wait.
func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}

	seconds, err := strconv.ParseInt(value, 10, 64)
	if errors.Is(err, strconv.ErrRange) && !strings.HasPrefix(value, "-") {
		// Too many seconds for an int64, so wait as long as possible
		return time.Duration(math.MaxInt64), true
	}
	if err == nil {
		switch {
		case seconds < 0:
			return 0, false
		case seconds > maxRetryAfterSeconds:
			// Too many seconds for a time.Duration, so wait as long as possible
			return time.Duration(math.MaxInt64), true
		default:
			return time.Duration(seconds) * time.Second, true
		}
	}

	if date, parseErr := http.ParseTime(value); parseErr == nil {
		if wait := date.Sub(now); wait > 0 {
			return wait, true
		}
		return 0, true
	}

	return 0, false
}

// sleepContext waits for the given duration or until ctx is done.
func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
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
// A value of zero or less disables rate limiting.
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	if requestsPerSecond < 0 {
		requestsPerSecond = 0
	}

	return &RateLimiter{
		requestsLimit: requestsPerSecond,
		windowSize:    time.Second,
		requests:      make([]time.Time, 0, requestsPerSecond),
	}
}

// Wait blocks until a request is allowed under the rate limit or ctx is done.
// A context that is already done returns its error without using a slot.
func (r *RateLimiter) Wait(ctx context.Context) error {
	if r.requestsLimit <= 0 {
		return ctx.Err()
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

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
			// Wait until the oldest request expires
			waitTime := r.requests[0].Add(r.windowSize).Sub(now)
			r.mu.Unlock()

			if err := sleepContext(ctx, waitTime); err != nil {
				return err
			}
			continue
		}

		// Record this request and return
		r.requests = append(r.requests, now)
		r.mu.Unlock()
		return nil
	}
}

// RateLimitedClient is a Client with rate limiting and automatic retries enabled.
// Service methods (such as Payments.List) and generic requests all use them.
type RateLimitedClient struct {
	*Client
}

// NewRateLimitedClient creates a client that sends at most 10 requests per second and
// retries failed requests with DefaultRetryConfig. Pass WithRateLimiter or
// WithRetryConfig to use other settings.
func NewRateLimitedClient(apiToken string, opts ...ClientOption) *RateLimitedClient {
	// Defaults come first so that opts can override them.
	options := make([]ClientOption, 0, 2+len(opts))
	options = append(options,
		WithRateLimiter(NewRateLimiter(defaultRequestsPerSecond)),
		WithRetryConfig(DefaultRetryConfig()),
	)
	options = append(options, opts...)

	return &RateLimitedClient{
		Client: NewClient(apiToken, options...),
	}
}

// SetRateLimit sets the rate limit for API requests.
// A value of zero or less disables rate limiting.
func (c *RateLimitedClient) SetRateLimit(requestsPerSecond int) {
	limiter := NewRateLimiter(requestsPerSecond)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.rateLimiter = limiter
}

// SetRetryConfig sets the retry configuration. A nil config disables retries.
func (c *RateLimitedClient) SetRetryConfig(config *RetryConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.retryConfig = config
}

// DoRequestWithRetry performs a generic API request with rate limiting and retries.
// query may be nil, url.Values, map[string][]string or map[string]string.
func (c *RateLimitedClient) DoRequestWithRetry(ctx context.Context, method, path string, query, body, result interface{}) error {
	values, err := queryValues(query)
	if err != nil {
		return err
	}

	return c.doRequest(ctx, method, path, values, body, result)
}

// queryValues converts the supported query parameter types to url.Values.
func queryValues(query interface{}) (url.Values, error) {
	switch q := query.(type) {
	case nil:
		return nil, nil
	case url.Values:
		return q, nil
	case map[string][]string:
		return url.Values(q), nil
	case map[string]string:
		values := make(url.Values, len(q))
		for key, value := range q {
			values.Set(key, value)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unsupported query type %T: use url.Values", query)
	}
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

// ParseRateLimitHeaders parses rate limit information from HTTP response headers,
// such as APIError.Headers.
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

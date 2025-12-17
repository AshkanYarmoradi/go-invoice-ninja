package invoiceninja

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the Invoice Ninja API.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int `json:"-"`

	// Message is the error message.
	Message string `json:"message,omitempty"`

	// Errors contains field-specific validation errors.
	Errors map[string][]string `json:"errors,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Invoice Ninja API error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Invoice Ninja API error (status %d)", e.StatusCode)
}

// IsNotFound returns true if the error is a 404 Not Found error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized returns true if the error is a 401 Unauthorized error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403 Forbidden error.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsValidationError returns true if the error is a 422 Unprocessable Entity error.
func (e *APIError) IsValidationError() bool {
	return e.StatusCode == http.StatusUnprocessableEntity
}

// IsRateLimited returns true if the error is a 429 Too Many Requests error.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServerError returns true if the error is a 5xx server error.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}

// parseAPIError parses an API error response.
func parseAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: statusCode,
	}

	// Try to parse the error response
	if len(body) > 0 {
		var errResp struct {
			Message string              `json:"message"`
			Errors  map[string][]string `json:"errors"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil {
			apiErr.Message = errResp.Message
			apiErr.Errors = errResp.Errors
		}
	}

	// Set default messages for common status codes
	if apiErr.Message == "" {
		switch statusCode {
		case http.StatusBadRequest:
			apiErr.Message = "bad request"
		case http.StatusUnauthorized:
			apiErr.Message = "unauthorized - check your API token"
		case http.StatusForbidden:
			apiErr.Message = "forbidden - you don't have permission to access this resource"
		case http.StatusNotFound:
			apiErr.Message = "resource not found"
		case http.StatusUnprocessableEntity:
			apiErr.Message = "validation error"
		case http.StatusTooManyRequests:
			apiErr.Message = "rate limit exceeded"
		default:
			if statusCode >= 500 {
				apiErr.Message = "server error"
			}
		}
	}

	return apiErr
}

// IsAPIError checks if an error is an APIError and returns it.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

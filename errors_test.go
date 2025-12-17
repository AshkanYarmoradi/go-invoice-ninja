package invoiceninja

import (
	"net/http"
	"testing"
)

func TestAPIErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name:     "with message",
			err:      &APIError{StatusCode: 400, Message: "bad request"},
			expected: "Invoice Ninja API error (status 400): bad request",
		},
		{
			name:     "without message",
			err:      &APIError{StatusCode: 500},
			expected: "Invoice Ninja API error (status 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIErrorMethods(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		method   func(*APIError) bool
		expected bool
	}{
		{
			name:     "IsNotFound true",
			err:      &APIError{StatusCode: http.StatusNotFound},
			method:   (*APIError).IsNotFound,
			expected: true,
		},
		{
			name:     "IsNotFound false",
			err:      &APIError{StatusCode: http.StatusOK},
			method:   (*APIError).IsNotFound,
			expected: false,
		},
		{
			name:     "IsUnauthorized true",
			err:      &APIError{StatusCode: http.StatusUnauthorized},
			method:   (*APIError).IsUnauthorized,
			expected: true,
		},
		{
			name:     "IsForbidden true",
			err:      &APIError{StatusCode: http.StatusForbidden},
			method:   (*APIError).IsForbidden,
			expected: true,
		},
		{
			name:     "IsValidationError true",
			err:      &APIError{StatusCode: http.StatusUnprocessableEntity},
			method:   (*APIError).IsValidationError,
			expected: true,
		},
		{
			name:     "IsRateLimited true",
			err:      &APIError{StatusCode: http.StatusTooManyRequests},
			method:   (*APIError).IsRateLimited,
			expected: true,
		},
		{
			name:     "IsServerError true",
			err:      &APIError{StatusCode: 500},
			method:   (*APIError).IsServerError,
			expected: true,
		},
		{
			name:     "IsServerError true for 503",
			err:      &APIError{StatusCode: 503},
			method:   (*APIError).IsServerError,
			expected: true,
		},
		{
			name:     "IsServerError false for 400",
			err:      &APIError{StatusCode: 400},
			method:   (*APIError).IsServerError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.method(tt.err); got != tt.expected {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestParseAPIError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           []byte
		expectedMsg    string
		expectedErrors map[string][]string
	}{
		{
			name:        "parse JSON error",
			statusCode:  422,
			body:        []byte(`{"message":"Validation failed","errors":{"email":["Email is required"]}}`),
			expectedMsg: "Validation failed",
			expectedErrors: map[string][]string{
				"email": {"Email is required"},
			},
		},
		{
			name:        "invalid JSON",
			statusCode:  400,
			body:        []byte(`invalid json`),
			expectedMsg: "bad request",
		},
		{
			name:        "empty body 401",
			statusCode:  401,
			body:        nil,
			expectedMsg: "unauthorized - check your API token",
		},
		{
			name:        "empty body 403",
			statusCode:  403,
			body:        nil,
			expectedMsg: "forbidden - you don't have permission to access this resource",
		},
		{
			name:        "empty body 404",
			statusCode:  404,
			body:        nil,
			expectedMsg: "resource not found",
		},
		{
			name:        "empty body 429",
			statusCode:  429,
			body:        nil,
			expectedMsg: "rate limit exceeded",
		},
		{
			name:        "empty body 500",
			statusCode:  500,
			body:        nil,
			expectedMsg: "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.statusCode, tt.body)

			if err.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %v, want %v", err.StatusCode, tt.statusCode)
			}

			if err.Message != tt.expectedMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.expectedMsg)
			}

			if tt.expectedErrors != nil {
				for k, v := range tt.expectedErrors {
					if err.Errors[k] == nil {
						t.Errorf("expected error for key %s", k)
					} else if err.Errors[k][0] != v[0] {
						t.Errorf("Errors[%s] = %v, want %v", k, err.Errors[k], v)
					}
				}
			}
		})
	}
}

func TestIsAPIError(t *testing.T) {
	apiErr := &APIError{StatusCode: 400}

	got, ok := IsAPIError(apiErr)
	if !ok {
		t.Error("expected ok to be true")
	}
	if got != apiErr {
		t.Error("expected returned error to be same as input")
	}

	// Test with non-APIError
	var regularErr error = nil
	_, ok = IsAPIError(regularErr)
	if ok {
		t.Error("expected ok to be false for nil error")
	}
}

package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// PaymentTermsService handles payment terms-related API operations.
type PaymentTermsService struct {
	client *Client
}

// PaymentTerm represents a payment term in Invoice Ninja.
type PaymentTerm struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	NumDays    int    `json:"num_days,omitempty"`
	IsDefault  bool   `json:"is_default,omitempty"`
	IsDeleted  bool   `json:"is_deleted,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	UpdatedAt  int64  `json:"updated_at,omitempty"`
	ArchivedAt int64  `json:"archived_at,omitempty"`
}

// PaymentTermListOptions specifies the optional parameters for listing payment terms.
type PaymentTermListOptions struct {
	PerPage int
	Page    int
	Include string
}

// toQuery converts options to URL query parameters.
func (o *PaymentTermListOptions) toQuery() url.Values {
	if o == nil {
		return nil
	}

	q := url.Values{}

	if o.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(o.PerPage))
	}
	if o.Page > 0 {
		q.Set("page", strconv.Itoa(o.Page))
	}
	if o.Include != "" {
		q.Set("include", o.Include)
	}

	return q
}

// List retrieves a list of payment terms.
func (s *PaymentTermsService) List(ctx context.Context, opts *PaymentTermListOptions) (*ListResponse[PaymentTerm], error) {
	var resp ListResponse[PaymentTerm]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/payment_terms", opts.toQuery(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single payment term by ID.
func (s *PaymentTermsService) Get(ctx context.Context, id string) (*PaymentTerm, error) {
	var resp SingleResponse[PaymentTerm]
	if err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/payment_terms/%s", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Create creates a new payment term.
func (s *PaymentTermsService) Create(ctx context.Context, term *PaymentTerm) (*PaymentTerm, error) {
	var resp SingleResponse[PaymentTerm]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/payment_terms", nil, term, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Update updates an existing payment term.
func (s *PaymentTermsService) Update(ctx context.Context, id string, term *PaymentTerm) (*PaymentTerm, error) {
	var resp SingleResponse[PaymentTerm]
	if err := s.client.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/payment_terms/%s", id), nil, term, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Delete deletes a payment term by ID.
func (s *PaymentTermsService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/payment_terms/%s", id), nil, nil, nil)
}

// Bulk performs a bulk action on multiple payment terms.
func (s *PaymentTermsService) Bulk(ctx context.Context, action string, ids []string) ([]PaymentTerm, error) {
	req := BulkAction{
		Action: action,
		IDs:    ids,
	}

	var resp ListResponse[PaymentTerm]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/payment_terms/bulk", nil, req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Archive archives a payment term.
func (s *PaymentTermsService) Archive(ctx context.Context, id string) (*PaymentTerm, error) {
	terms, err := s.Bulk(ctx, "archive", []string{id})
	if err != nil {
		return nil, err
	}
	if len(terms) == 0 {
		return nil, fmt.Errorf("no payment term returned from bulk action")
	}
	return &terms[0], nil
}

// Restore restores an archived payment term.
func (s *PaymentTermsService) Restore(ctx context.Context, id string) (*PaymentTerm, error) {
	terms, err := s.Bulk(ctx, "restore", []string{id})
	if err != nil {
		return nil, err
	}
	if len(terms) == 0 {
		return nil, fmt.Errorf("no payment term returned from bulk action")
	}
	return &terms[0], nil
}

// GetBlank retrieves a blank payment term object with default values.
func (s *PaymentTermsService) GetBlank(ctx context.Context) (*PaymentTerm, error) {
	var resp SingleResponse[PaymentTerm]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/payment_terms/create", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

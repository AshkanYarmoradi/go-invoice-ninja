package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// InvoicesService handles invoice-related API operations.
type InvoicesService struct {
	client *Client
}

// InvoiceListOptions specifies the optional parameters for listing invoices.
type InvoiceListOptions struct {
	// PerPage is the number of results per page (default 20).
	PerPage int

	// Page is the page number.
	Page int

	// Filter searches across multiple fields.
	Filter string

	// ClientID filters by client.
	ClientID string

	// Status filters by status (comma-separated: active, archived, deleted).
	Status string

	// CreatedAt filters by creation date.
	CreatedAt string

	// UpdatedAt filters by update date.
	UpdatedAt string

	// IsDeleted filters by deleted status.
	IsDeleted *bool

	// Sort specifies the sort order (e.g., "id|desc", "number|asc").
	Sort string

	// Include specifies related entities to include.
	Include string
}

// toQuery converts options to URL query parameters.
func (o *InvoiceListOptions) toQuery() url.Values {
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
	if o.Filter != "" {
		q.Set("filter", o.Filter)
	}
	if o.ClientID != "" {
		q.Set("client_id", o.ClientID)
	}
	if o.Status != "" {
		q.Set("status", o.Status)
	}
	if o.CreatedAt != "" {
		q.Set("created_at", o.CreatedAt)
	}
	if o.UpdatedAt != "" {
		q.Set("updated_at", o.UpdatedAt)
	}
	if o.IsDeleted != nil {
		q.Set("is_deleted", strconv.FormatBool(*o.IsDeleted))
	}
	if o.Sort != "" {
		q.Set("sort", o.Sort)
	}
	if o.Include != "" {
		q.Set("include", o.Include)
	}

	return q
}

// List retrieves a list of invoices.
func (s *InvoicesService) List(ctx context.Context, opts *InvoiceListOptions) (*ListResponse[Invoice], error) {
	var resp ListResponse[Invoice]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/invoices", opts.toQuery(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single invoice by ID.
func (s *InvoicesService) Get(ctx context.Context, id string) (*Invoice, error) {
	var resp SingleResponse[Invoice]
	if err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/invoices/%s", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Create creates a new invoice.
func (s *InvoicesService) Create(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	var resp SingleResponse[Invoice]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/invoices", nil, invoice, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Update updates an existing invoice.
func (s *InvoicesService) Update(ctx context.Context, id string, invoice *Invoice) (*Invoice, error) {
	var resp SingleResponse[Invoice]
	if err := s.client.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/invoices/%s", id), nil, invoice, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Delete deletes an invoice by ID.
func (s *InvoicesService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/invoices/%s", id), nil, nil, nil)
}

// Archive archives an invoice.
func (s *InvoicesService) Archive(ctx context.Context, id string) (*Invoice, error) {
	return s.bulkAction(ctx, "archive", id)
}

// Restore restores an archived invoice.
func (s *InvoicesService) Restore(ctx context.Context, id string) (*Invoice, error) {
	return s.bulkAction(ctx, "restore", id)
}

// MarkPaid marks an invoice as paid.
func (s *InvoicesService) MarkPaid(ctx context.Context, id string) (*Invoice, error) {
	return s.bulkAction(ctx, "mark_paid", id)
}

// MarkSent marks an invoice as sent.
func (s *InvoicesService) MarkSent(ctx context.Context, id string) (*Invoice, error) {
	return s.bulkAction(ctx, "mark_sent", id)
}

// Email sends an invoice via email.
func (s *InvoicesService) Email(ctx context.Context, id string) (*Invoice, error) {
	return s.bulkAction(ctx, "email", id)
}

// Bulk performs a bulk action on multiple invoices.
func (s *InvoicesService) Bulk(ctx context.Context, action string, ids []string) ([]Invoice, error) {
	req := BulkAction{
		Action: action,
		IDs:    ids,
	}

	var resp ListResponse[Invoice]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/invoices/bulk", nil, req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// bulkAction performs a single-item bulk action.
func (s *InvoicesService) bulkAction(ctx context.Context, action, id string) (*Invoice, error) {
	invoices, err := s.Bulk(ctx, action, []string{id})
	if err != nil {
		return nil, err
	}
	if len(invoices) == 0 {
		return nil, fmt.Errorf("no invoice returned from bulk action")
	}
	return &invoices[0], nil
}

// GetBlank retrieves a blank invoice object with default values.
func (s *InvoicesService) GetBlank(ctx context.Context) (*Invoice, error) {
	var resp SingleResponse[Invoice]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/invoices/create", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Download downloads an invoice PDF.
func (s *InvoicesService) Download(ctx context.Context, invitationKey string) ([]byte, error) {
	// This would need special handling for binary response
	// For now, we'll return the raw bytes
	return nil, fmt.Errorf("not implemented - use client.Request with custom handling")
}

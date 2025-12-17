package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// PaymentsService handles payment-related API operations.
type PaymentsService struct {
	client *Client
}

// PaymentListOptions specifies the optional parameters for listing payments.
type PaymentListOptions struct {
	// PerPage is the number of results per page (default 20).
	PerPage int

	// Page is the page number.
	Page int

	// Filter searches across amount, date, and custom values.
	Filter string

	// Number searches by payment number.
	Number string

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

	// VendorID filters by vendor.
	VendorID string

	// Sort specifies the sort order (e.g., "id|desc", "number|asc").
	Sort string

	// Include specifies related entities to include.
	Include string
}

// toQuery converts options to URL query parameters.
func (o *PaymentListOptions) toQuery() url.Values {
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
	if o.Number != "" {
		q.Set("number", o.Number)
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
	if o.VendorID != "" {
		q.Set("vendor_id", o.VendorID)
	}
	if o.Sort != "" {
		q.Set("sort", o.Sort)
	}
	if o.Include != "" {
		q.Set("include", o.Include)
	}

	return q
}

// List retrieves a list of payments.
func (s *PaymentsService) List(ctx context.Context, opts *PaymentListOptions) (*ListResponse[Payment], error) {
	var resp ListResponse[Payment]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/payments", opts.toQuery(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single payment by ID.
func (s *PaymentsService) Get(ctx context.Context, id string) (*Payment, error) {
	var resp SingleResponse[Payment]
	if err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/payments/%s", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Create creates a new payment.
func (s *PaymentsService) Create(ctx context.Context, payment *PaymentRequest) (*Payment, error) {
	var resp SingleResponse[Payment]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/payments", nil, payment, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// CreateWithEmailReceipt creates a new payment and optionally sends an email receipt.
func (s *PaymentsService) CreateWithEmailReceipt(ctx context.Context, payment *PaymentRequest, sendEmail bool) (*Payment, error) {
	q := url.Values{}
	q.Set("email_receipt", strconv.FormatBool(sendEmail))

	var resp SingleResponse[Payment]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/payments", q, payment, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Update updates an existing payment.
func (s *PaymentsService) Update(ctx context.Context, id string, payment *PaymentRequest) (*Payment, error) {
	var resp SingleResponse[Payment]
	if err := s.client.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/payments/%s", id), nil, payment, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Delete deletes a payment by ID.
func (s *PaymentsService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/payments/%s", id), nil, nil, nil)
}

// Refund creates a refund for a payment.
func (s *PaymentsService) Refund(ctx context.Context, refund *RefundRequest) (*Payment, error) {
	var resp SingleResponse[Payment]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/payments/refund", nil, refund, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Archive archives a payment.
func (s *PaymentsService) Archive(ctx context.Context, id string) (*Payment, error) {
	return s.bulkAction(ctx, "archive", id)
}

// Restore restores an archived payment.
func (s *PaymentsService) Restore(ctx context.Context, id string) (*Payment, error) {
	return s.bulkAction(ctx, "restore", id)
}

// Bulk performs a bulk action on multiple payments.
func (s *PaymentsService) Bulk(ctx context.Context, action string, ids []string) ([]Payment, error) {
	req := BulkAction{
		Action: action,
		IDs:    ids,
	}

	var resp ListResponse[Payment]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/payments/bulk", nil, req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// bulkAction performs a single-item bulk action.
func (s *PaymentsService) bulkAction(ctx context.Context, action, id string) (*Payment, error) {
	payments, err := s.Bulk(ctx, action, []string{id})
	if err != nil {
		return nil, err
	}
	if len(payments) == 0 {
		return nil, fmt.Errorf("no payment returned from bulk action")
	}
	return &payments[0], nil
}

// GetBlank retrieves a blank payment object with default values.
func (s *PaymentsService) GetBlank(ctx context.Context) (*Payment, error) {
	var resp SingleResponse[Payment]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/payments/create", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

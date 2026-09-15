package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ExpensesService handles expense-related API operations.
type ExpensesService struct {
	client *Client
}

// ExpenseListOptions specifies the optional parameters for listing expenses.
type ExpenseListOptions struct {
	// PerPage is the number of results per page (default 20).
	PerPage int

	// Page is the page number.
	Page int

	// Filter searches across multiple fields.
	Filter string

	// ClientID filters by client.
	ClientID string

	// Status filters by status (comma-separated: active, deleted).
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
func (o *ExpenseListOptions) toQuery() url.Values {
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

// List retrieves a list of expenses.
func (s *ExpensesService) List(ctx context.Context, opts *ExpenseListOptions) (*ListResponse[Expense], error) {
	var resp ListResponse[Expense]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/expenses", opts.toQuery(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single expense by ID.
func (s *ExpensesService) Get(ctx context.Context, id string) (*Expense, error) {
	var resp SingleResponse[Expense]
	if err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/expenses/%s", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Create creates a new expense.
func (s *ExpensesService) Create(ctx context.Context, expense *Expense) (*Expense, error) {
	var resp SingleResponse[Expense]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/expenses", nil, expense, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Update updates an existing expense.
func (s *ExpensesService) Update(ctx context.Context, id string, expense *Expense) (*Expense, error) {
	var resp SingleResponse[Expense]
	if err := s.client.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/expenses/%s", id), nil, expense, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Delete deletes an expense by ID.
func (s *ExpensesService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/expenses/%s", id), nil, nil, nil)
}

// GetBlank retrieves a blank expense object with default values.
func (s *ExpensesService) GetBlank(ctx context.Context) (*Expense, error) {
	var resp SingleResponse[Expense]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/expenses/create", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

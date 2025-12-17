package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ClientsService handles client-related API operations.
type ClientsService struct {
	client *Client
}

// ClientListOptions specifies the optional parameters for listing clients.
type ClientListOptions struct {
	// PerPage is the number of results per page (default 20).
	PerPage int

	// Page is the page number.
	Page int

	// Filter searches across multiple fields.
	Filter string

	// Balance filters by balance (e.g., "gt:1000", "lt:500").
	Balance string

	// Status filters by status (comma-separated: active, archived, deleted).
	Status string

	// CreatedAt filters by creation date.
	CreatedAt string

	// UpdatedAt filters by update date.
	UpdatedAt string

	// IsDeleted filters by deleted status.
	IsDeleted *bool

	// Sort specifies the sort order (e.g., "name|desc", "balance|asc").
	Sort string

	// Include specifies related entities to include (contacts, documents, activities).
	Include string
}

// toQuery converts options to URL query parameters.
func (o *ClientListOptions) toQuery() url.Values {
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
	if o.Balance != "" {
		q.Set("balance", o.Balance)
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

// List retrieves a list of clients.
func (s *ClientsService) List(ctx context.Context, opts *ClientListOptions) (*ListResponse[INClient], error) {
	var resp ListResponse[INClient]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/clients", opts.toQuery(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single client by ID.
func (s *ClientsService) Get(ctx context.Context, id string) (*INClient, error) {
	var resp SingleResponse[INClient]
	if err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/clients/%s", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Create creates a new client.
func (s *ClientsService) Create(ctx context.Context, client *INClient) (*INClient, error) {
	var resp SingleResponse[INClient]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/clients", nil, client, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Update updates an existing client.
func (s *ClientsService) Update(ctx context.Context, id string, client *INClient) (*INClient, error) {
	var resp SingleResponse[INClient]
	if err := s.client.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/clients/%s", id), nil, client, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Delete deletes a client by ID (soft delete).
func (s *ClientsService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/clients/%s", id), nil, nil, nil)
}

// Purge permanently removes a client and all their records.
func (s *ClientsService) Purge(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/clients/%s/purge", id), nil, nil, nil)
}

// Archive archives a client.
func (s *ClientsService) Archive(ctx context.Context, id string) (*INClient, error) {
	return s.bulkAction(ctx, "archive", id)
}

// Restore restores an archived client.
func (s *ClientsService) Restore(ctx context.Context, id string) (*INClient, error) {
	return s.bulkAction(ctx, "restore", id)
}

// Merge merges two clients.
func (s *ClientsService) Merge(ctx context.Context, primaryID, mergeableID string) (*INClient, error) {
	var resp SingleResponse[INClient]
	path := fmt.Sprintf("/api/v1/clients/%s/%s/merge", primaryID, mergeableID)
	if err := s.client.doRequest(ctx, "POST", path, nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Bulk performs a bulk action on multiple clients.
func (s *ClientsService) Bulk(ctx context.Context, action string, ids []string) ([]INClient, error) {
	req := BulkAction{
		Action: action,
		IDs:    ids,
	}

	var resp ListResponse[INClient]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/clients/bulk", nil, req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// bulkAction performs a single-item bulk action.
func (s *ClientsService) bulkAction(ctx context.Context, action, id string) (*INClient, error) {
	clients, err := s.Bulk(ctx, action, []string{id})
	if err != nil {
		return nil, err
	}
	if len(clients) == 0 {
		return nil, fmt.Errorf("no client returned from bulk action")
	}
	return &clients[0], nil
}

// GetBlank retrieves a blank client object with default values.
func (s *ClientsService) GetBlank(ctx context.Context) (*INClient, error) {
	var resp SingleResponse[INClient]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/clients/create", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// StatementRequest represents a client statement request.
type StatementRequest struct {
	ClientID   string `json:"client_id"`
	StartDate  string `json:"start_date,omitempty"`
	EndDate    string `json:"end_date,omitempty"`
	ShowPayments bool `json:"show_payments_table,omitempty"`
	ShowAging  bool   `json:"show_aging_table,omitempty"`
	ShowCredits bool  `json:"show_credits_table,omitempty"`
	Status     string `json:"status,omitempty"`
}

// GetStatement generates a client statement.
func (s *ClientsService) GetStatement(ctx context.Context, req *StatementRequest) ([]byte, error) {
	// This would need special handling for PDF response
	return nil, fmt.Errorf("not implemented - use client.Request with custom handling")
}

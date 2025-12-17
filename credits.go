package invoiceninja

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// CreditsService handles credit-related API operations.
type CreditsService struct {
	client *Client
}

// Credit represents a credit note in Invoice Ninja.
type Credit struct {
	ID                 string     `json:"id,omitempty"`
	UserID             string     `json:"user_id,omitempty"`
	AssignedUserID     string     `json:"assigned_user_id,omitempty"`
	ClientID           string     `json:"client_id,omitempty"`
	StatusID           string     `json:"status_id,omitempty"`
	InvoiceID          string     `json:"invoice_id,omitempty"`
	Number             string     `json:"number,omitempty"`
	PONumber           string     `json:"po_number,omitempty"`
	Terms              string     `json:"terms,omitempty"`
	PublicNotes        string     `json:"public_notes,omitempty"`
	PrivateNotes       string     `json:"private_notes,omitempty"`
	Footer             string     `json:"footer,omitempty"`
	CustomValue1       string     `json:"custom_value1,omitempty"`
	CustomValue2       string     `json:"custom_value2,omitempty"`
	CustomValue3       string     `json:"custom_value3,omitempty"`
	CustomValue4       string     `json:"custom_value4,omitempty"`
	TaxName1           string     `json:"tax_name1,omitempty"`
	TaxName2           string     `json:"tax_name2,omitempty"`
	TaxName3           string     `json:"tax_name3,omitempty"`
	TaxRate1           float64    `json:"tax_rate1,omitempty"`
	TaxRate2           float64    `json:"tax_rate2,omitempty"`
	TaxRate3           float64    `json:"tax_rate3,omitempty"`
	TotalTaxes         float64    `json:"total_taxes,omitempty"`
	Amount             float64    `json:"amount,omitempty"`
	Balance            float64    `json:"balance,omitempty"`
	PaidToDate         float64    `json:"paid_to_date,omitempty"`
	Discount           float64    `json:"discount,omitempty"`
	Partial            float64    `json:"partial,omitempty"`
	IsAmountDiscount   bool       `json:"is_amount_discount,omitempty"`
	IsDeleted          bool       `json:"is_deleted,omitempty"`
	UsesInclusiveTaxes bool       `json:"uses_inclusive_taxes,omitempty"`
	Date               string     `json:"date,omitempty"`
	LastSentDate       string     `json:"last_sent_date,omitempty"`
	NextSendDate       string     `json:"next_send_date,omitempty"`
	PartialDueDate     string     `json:"partial_due_date,omitempty"`
	DueDate            string     `json:"due_date,omitempty"`
	LineItems          []LineItem `json:"line_items,omitempty"`
	UpdatedAt          int64      `json:"updated_at,omitempty"`
	ArchivedAt         int64      `json:"archived_at,omitempty"`
	CreatedAt          int64      `json:"created_at,omitempty"`
}

// CreditListOptions specifies the optional parameters for listing credits.
type CreditListOptions struct {
	PerPage   int
	Page      int
	Filter    string
	ClientID  string
	Status    string
	CreatedAt string
	UpdatedAt string
	IsDeleted *bool
	Sort      string
	Include   string
}

// toQuery converts options to URL query parameters.
func (o *CreditListOptions) toQuery() url.Values {
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

// List retrieves a list of credits.
func (s *CreditsService) List(ctx context.Context, opts *CreditListOptions) (*ListResponse[Credit], error) {
	var resp ListResponse[Credit]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/credits", opts.toQuery(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a single credit by ID.
func (s *CreditsService) Get(ctx context.Context, id string) (*Credit, error) {
	var resp SingleResponse[Credit]
	if err := s.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/credits/%s", id), nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Create creates a new credit.
func (s *CreditsService) Create(ctx context.Context, credit *Credit) (*Credit, error) {
	var resp SingleResponse[Credit]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/credits", nil, credit, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Update updates an existing credit.
func (s *CreditsService) Update(ctx context.Context, id string, credit *Credit) (*Credit, error) {
	var resp SingleResponse[Credit]
	if err := s.client.doRequest(ctx, "PUT", fmt.Sprintf("/api/v1/credits/%s", id), nil, credit, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// Delete deletes a credit by ID.
func (s *CreditsService) Delete(ctx context.Context, id string) error {
	return s.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/credits/%s", id), nil, nil, nil)
}

// Bulk performs a bulk action on multiple credits.
func (s *CreditsService) Bulk(ctx context.Context, action string, ids []string) ([]Credit, error) {
	req := BulkAction{
		Action: action,
		IDs:    ids,
	}

	var resp ListResponse[Credit]
	if err := s.client.doRequest(ctx, "POST", "/api/v1/credits/bulk", nil, req, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Archive archives a credit.
func (s *CreditsService) Archive(ctx context.Context, id string) (*Credit, error) {
	return s.bulkAction(ctx, "archive", id)
}

// Restore restores an archived credit.
func (s *CreditsService) Restore(ctx context.Context, id string) (*Credit, error) {
	return s.bulkAction(ctx, "restore", id)
}

// MarkSent marks a credit as sent.
func (s *CreditsService) MarkSent(ctx context.Context, id string) (*Credit, error) {
	return s.bulkAction(ctx, "mark_sent", id)
}

// Email sends a credit via email.
func (s *CreditsService) Email(ctx context.Context, id string) (*Credit, error) {
	return s.bulkAction(ctx, "email", id)
}

// bulkAction performs a single-item bulk action.
func (s *CreditsService) bulkAction(ctx context.Context, action, id string) (*Credit, error) {
	credits, err := s.Bulk(ctx, action, []string{id})
	if err != nil {
		return nil, err
	}
	if len(credits) == 0 {
		return nil, fmt.Errorf("no credit returned from bulk action")
	}
	return &credits[0], nil
}

// GetBlank retrieves a blank credit object with default values.
func (s *CreditsService) GetBlank(ctx context.Context) (*Credit, error) {
	var resp SingleResponse[Credit]
	if err := s.client.doRequest(ctx, "GET", "/api/v1/credits/create", nil, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

package invoiceninja

import "encoding/json"

// Payment represents a payment in Invoice Ninja.
type Payment struct {
	ID                 string           `json:"id,omitempty"`
	ClientID           string           `json:"client_id,omitempty"`
	InvitationID       string           `json:"invitation_id,omitempty"`
	ClientContactID    string           `json:"client_contact_id,omitempty"`
	UserID             string           `json:"user_id,omitempty"`
	TypeID             string           `json:"type_id,omitempty"`
	Date               string           `json:"date,omitempty"`
	TransactionRef     string           `json:"transaction_reference,omitempty"`
	AssignedUserID     string           `json:"assigned_user_id,omitempty"`
	PrivateNotes       string           `json:"private_notes,omitempty"`
	IsManual           bool             `json:"is_manual,omitempty"`
	IsDeleted          bool             `json:"is_deleted,omitempty"`
	Amount             float64          `json:"amount,omitempty"`
	Refunded           float64          `json:"refunded,omitempty"`
	UpdatedAt          int64            `json:"updated_at,omitempty"`
	ArchivedAt         int64            `json:"archived_at,omitempty"`
	CompanyGatewayID   string           `json:"company_gateway_id,omitempty"`
	Number             string           `json:"number,omitempty"`
	CategoryID         string           `json:"category_id,omitempty"`
	CustomValue1       string           `json:"custom_value1,omitempty"`
	CustomValue2       string           `json:"custom_value2,omitempty"`
	CustomValue3       string           `json:"custom_value3,omitempty"`
	CustomValue4       string           `json:"custom_value4,omitempty"`
	ExchangeCurrencyID string           `json:"exchange_currency_id,omitempty"`
	ExchangeRate       float64          `json:"exchange_rate,omitempty"`
	IdempotencyKey     string           `json:"idempotency_key,omitempty"`
	Paymentables       []Paymentable    `json:"paymentables,omitempty"`
	Invoices           []PaymentInvoice `json:"invoices,omitempty"`
	Credits            []PaymentCredit  `json:"credits,omitempty"`
}

// PaymentRequest represents a request to create or update a payment.
type PaymentRequest struct {
	ClientID        string           `json:"client_id,omitempty"`
	ClientContactID string           `json:"client_contact_id,omitempty"`
	UserID          string           `json:"user_id,omitempty"`
	TypeID          string           `json:"type_id,omitempty"`
	Date            string           `json:"date,omitempty"`
	TransactionRef  string           `json:"transaction_reference,omitempty"`
	AssignedUserID  string           `json:"assigned_user_id,omitempty"`
	PrivateNotes    string           `json:"private_notes,omitempty"`
	Amount          float64          `json:"amount,omitempty"`
	Invoices        []PaymentInvoice `json:"invoices,omitempty"`
	Credits         []PaymentCredit  `json:"credits,omitempty"`
	Number          string           `json:"number,omitempty"`
}

// PaymentInvoice represents an invoice applied to a payment.
type PaymentInvoice struct {
	InvoiceID string  `json:"invoice_id,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
}

// PaymentCredit represents a credit applied to a payment.
type PaymentCredit struct {
	CreditID string  `json:"credit_id,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
}

// Paymentable represents a paymentable entity (invoice or credit attached to a payment).
type Paymentable struct {
	ID        string  `json:"id,omitempty"`
	InvoiceID string  `json:"invoice_id,omitempty"`
	CreditID  string  `json:"credit_id,omitempty"`
	Refunded  float64 `json:"refunded,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	UpdatedAt int64   `json:"updated_at,omitempty"`
	CreatedAt int64   `json:"created_at,omitempty"`
}

// Invoice represents an invoice in Invoice Ninja.
type Invoice struct {
	ID             string     `json:"id,omitempty"`
	UserID         string     `json:"user_id,omitempty"`
	AssignedUserID string     `json:"assigned_user_id,omitempty"`
	ClientID       string     `json:"client_id,omitempty"`
	StatusID       string     `json:"status_id,omitempty"`
	Number         string     `json:"number,omitempty"`
	PONumber       string     `json:"po_number,omitempty"`
	Terms          string     `json:"terms,omitempty"`
	PublicNotes    string     `json:"public_notes,omitempty"`
	PrivateNotes   string     `json:"private_notes,omitempty"`
	Footer         string     `json:"footer,omitempty"`
	CustomValue1   string     `json:"custom_value1,omitempty"`
	CustomValue2   string     `json:"custom_value2,omitempty"`
	CustomValue3   string     `json:"custom_value3,omitempty"`
	CustomValue4   string     `json:"custom_value4,omitempty"`
	TaxName1       string     `json:"tax_name1,omitempty"`
	TaxName2       string     `json:"tax_name2,omitempty"`
	TaxName3       string     `json:"tax_name3,omitempty"`
	TaxRate1       float64    `json:"tax_rate1,omitempty"`
	TaxRate2       float64    `json:"tax_rate2,omitempty"`
	TaxRate3       float64    `json:"tax_rate3,omitempty"`
	TotalTaxes     float64    `json:"total_taxes,omitempty"`
	Amount         float64    `json:"amount,omitempty"`
	Balance        float64    `json:"balance,omitempty"`
	PaidToDate     float64    `json:"paid_to_date,omitempty"`
	Discount       float64    `json:"discount,omitempty"`
	PartialDueDate string     `json:"partial_due_date,omitempty"`
	DueDate        string     `json:"due_date,omitempty"`
	Date           string     `json:"date,omitempty"`
	LineItems      []LineItem `json:"line_items,omitempty"`
	IsDeleted      bool       `json:"is_deleted,omitempty"`
	UpdatedAt      int64      `json:"updated_at,omitempty"`
	ArchivedAt     int64      `json:"archived_at,omitempty"`
	CreatedAt      int64      `json:"created_at,omitempty"`
}

// LineItem represents a line item on an invoice.
type LineItem struct {
	Quantity     float64 `json:"quantity,omitempty"`
	Cost         float64 `json:"cost,omitempty"`
	ProductKey   string  `json:"product_key,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	Discount     float64 `json:"discount,omitempty"`
	IsAmountDisc bool    `json:"is_amount_discount,omitempty"`
	TaxName1     string  `json:"tax_name1,omitempty"`
	TaxRate1     float64 `json:"tax_rate1,omitempty"`
	TaxName2     string  `json:"tax_name2,omitempty"`
	TaxRate2     float64 `json:"tax_rate2,omitempty"`
	TaxName3     string  `json:"tax_name3,omitempty"`
	TaxRate3     float64 `json:"tax_rate3,omitempty"`
	CustomValue1 string  `json:"custom_value1,omitempty"`
	CustomValue2 string  `json:"custom_value2,omitempty"`
	CustomValue3 string  `json:"custom_value3,omitempty"`
	CustomValue4 string  `json:"custom_value4,omitempty"`
	TypeID       string  `json:"type_id,omitempty"`
}

// INClient represents a client in Invoice Ninja.
type INClient struct {
	ID               string          `json:"id,omitempty"`
	UserID           string          `json:"user_id,omitempty"`
	AssignedUserID   string          `json:"assigned_user_id,omitempty"`
	Name             string          `json:"name,omitempty"`
	Website          string          `json:"website,omitempty"`
	PrivateNotes     string          `json:"private_notes,omitempty"`
	PublicNotes      string          `json:"public_notes,omitempty"`
	Balance          float64         `json:"balance,omitempty"`
	PaidToDate       float64         `json:"paid_to_date,omitempty"`
	CreditBalance    float64         `json:"credit_balance,omitempty"`
	Phone            string          `json:"phone,omitempty"`
	Address1         string          `json:"address1,omitempty"`
	Address2         string          `json:"address2,omitempty"`
	City             string          `json:"city,omitempty"`
	State            string          `json:"state,omitempty"`
	PostalCode       string          `json:"postal_code,omitempty"`
	CountryID        string          `json:"country_id,omitempty"`
	IndustryID       string          `json:"industry_id,omitempty"`
	CustomValue1     string          `json:"custom_value1,omitempty"`
	CustomValue2     string          `json:"custom_value2,omitempty"`
	CustomValue3     string          `json:"custom_value3,omitempty"`
	CustomValue4     string          `json:"custom_value4,omitempty"`
	VatNumber        string          `json:"vat_number,omitempty"`
	IDNumber         string          `json:"id_number,omitempty"`
	Number           string          `json:"number,omitempty"`
	ShippingAddress1 string          `json:"shipping_address1,omitempty"`
	ShippingAddress2 string          `json:"shipping_address2,omitempty"`
	ShippingCity     string          `json:"shipping_city,omitempty"`
	ShippingState    string          `json:"shipping_state,omitempty"`
	ShippingPostal   string          `json:"shipping_postal_code,omitempty"`
	ShippingCountry  string          `json:"shipping_country_id,omitempty"`
	IsDeleted        bool            `json:"is_deleted,omitempty"`
	Contacts         []ClientContact `json:"contacts,omitempty"`
	UpdatedAt        int64           `json:"updated_at,omitempty"`
	ArchivedAt       int64           `json:"archived_at,omitempty"`
	CreatedAt        int64           `json:"created_at,omitempty"`
}

// ClientContact represents a contact for a client.
type ClientContact struct {
	ID           string `json:"id,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	IsPrimary    bool   `json:"is_primary,omitempty"`
	CustomValue1 string `json:"custom_value1,omitempty"`
	CustomValue2 string `json:"custom_value2,omitempty"`
	CustomValue3 string `json:"custom_value3,omitempty"`
	CustomValue4 string `json:"custom_value4,omitempty"`
}

// Meta represents pagination metadata.
type Meta struct {
	Pagination Pagination `json:"pagination,omitempty"`
}

// Pagination contains pagination details.
type Pagination struct {
	Total       int    `json:"total,omitempty"`
	Count       int    `json:"count,omitempty"`
	PerPage     int    `json:"per_page,omitempty"`
	CurrentPage int    `json:"current_page,omitempty"`
	TotalPages  int    `json:"total_pages,omitempty"`
	Links       *Links `json:"links,omitempty"`
}

// Links contains pagination links.
type Links struct {
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// ListResponse is a generic response structure for list endpoints.
type ListResponse[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta,omitempty"`
}

// SingleResponse is a generic response structure for single entity endpoints.
type SingleResponse[T any] struct {
	Data T `json:"data"`
}

// BulkAction represents a bulk action request.
type BulkAction struct {
	Action string   `json:"action"`
	IDs    []string `json:"ids"`
}

// RefundRequest represents a refund request.
type RefundRequest struct {
	ID            string           `json:"id"`
	Amount        float64          `json:"amount,omitempty"`
	Invoices      []PaymentInvoice `json:"invoices,omitempty"`
	Date          string           `json:"date,omitempty"`
	GatewayRefund bool             `json:"gateway_refund,omitempty"`
	SendEmail     bool             `json:"send_email,omitempty"`
}

// GenericResponse is used for arbitrary JSON responses.
type GenericResponse = json.RawMessage

package invoiceninja

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadsService handles file download operations.
type DownloadsService struct {
	client *Client
}

// DownloadInvoicePDF downloads an invoice PDF by invitation key.
func (s *DownloadsService) DownloadInvoicePDF(ctx context.Context, invitationKey string) ([]byte, error) {
	return s.downloadFile(ctx, fmt.Sprintf("/api/v1/invoice/%s/download", invitationKey))
}

// DownloadInvoiceDeliveryNote downloads an invoice delivery note PDF.
func (s *DownloadsService) DownloadInvoiceDeliveryNote(ctx context.Context, invoiceID string) ([]byte, error) {
	return s.downloadFile(ctx, fmt.Sprintf("/api/v1/invoices/%s/delivery_note", invoiceID))
}

// DownloadCreditPDF downloads a credit PDF by invitation key.
func (s *DownloadsService) DownloadCreditPDF(ctx context.Context, invitationKey string) ([]byte, error) {
	return s.downloadFile(ctx, fmt.Sprintf("/api/v1/credit/%s/download", invitationKey))
}

// DownloadQuotePDF downloads a quote PDF by invitation key.
func (s *DownloadsService) DownloadQuotePDF(ctx context.Context, invitationKey string) ([]byte, error) {
	return s.downloadFile(ctx, fmt.Sprintf("/api/v1/quote/%s/download", invitationKey))
}

// downloadFile performs a file download request.
func (s *DownloadsService) downloadFile(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.client.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-TOKEN", s.client.apiToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "application/pdf")

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, parseAPIError(resp.StatusCode, body)
	}

	return io.ReadAll(resp.Body)
}

// UploadsService handles file upload operations.
type UploadsService struct {
	client *Client
}

// UploadDocument uploads a document to an entity.
func (s *UploadsService) UploadDocument(ctx context.Context, entityType, entityID string, filePath string) error {
	return s.uploadFile(ctx, fmt.Sprintf("/api/v1/%s/%s/upload", entityType, entityID), filePath)
}

// UploadInvoiceDocument uploads a document to an invoice.
func (s *UploadsService) UploadInvoiceDocument(ctx context.Context, invoiceID string, filePath string) error {
	return s.uploadFile(ctx, fmt.Sprintf("/api/v1/invoices/%s/upload", invoiceID), filePath)
}

// UploadPaymentDocument uploads a document to a payment.
func (s *UploadsService) UploadPaymentDocument(ctx context.Context, paymentID string, filePath string) error {
	return s.uploadFile(ctx, fmt.Sprintf("/api/v1/payments/%s/upload", paymentID), filePath)
}

// UploadClientDocument uploads a document to a client.
func (s *UploadsService) UploadClientDocument(ctx context.Context, clientID string, filePath string) error {
	return s.uploadFile(ctx, fmt.Sprintf("/api/v1/clients/%s/upload", clientID), filePath)
}

// UploadCreditDocument uploads a document to a credit.
func (s *UploadsService) UploadCreditDocument(ctx context.Context, creditID string, filePath string) error {
	return s.uploadFile(ctx, fmt.Sprintf("/api/v1/credits/%s/upload", creditID), filePath)
}

// UploadDocumentFromReader uploads a document from an io.Reader.
func (s *UploadsService) UploadDocumentFromReader(ctx context.Context, entityType, entityID, filename string, reader io.Reader) error {
	return s.uploadFromReader(ctx, fmt.Sprintf("/api/v1/%s/%s/upload", entityType, entityID), filename, reader)
}

// uploadFile uploads a file from the filesystem.
func (s *UploadsService) uploadFile(ctx context.Context, path, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return s.uploadFromReader(ctx, path, filepath.Base(filePath), file)
}

// uploadFromReader uploads a file from an io.Reader.
func (s *UploadsService) uploadFromReader(ctx context.Context, path, filename string, reader io.Reader) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add _method field for PUT override
	if err := writer.WriteField("_method", "PUT"); err != nil {
		return fmt.Errorf("failed to write method field: %w", err)
	}

	// Create form file
	part, err := writer.CreateFormFile("documents[]", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, reader); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.client.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-TOKEN", s.client.apiToken)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return parseAPIError(resp.StatusCode, body)
	}

	return nil
}

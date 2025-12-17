package invoiceninja

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDownloadsServiceDownloadInvoicePDF(t *testing.T) {
	expectedPDF := []byte("%PDF-1.4 fake pdf content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/api/v1/invoice/") {
			t.Errorf("expected path to start with /api/v1/invoice/, got %s", r.URL.Path)
		}
		if !strings.HasSuffix(r.URL.Path, "/download") {
			t.Errorf("expected path to end with /download, got %s", r.URL.Path)
		}

		// Verify headers
		if r.Header.Get("X-API-TOKEN") != "test-token" {
			t.Errorf("expected X-API-TOKEN header")
		}
		if r.Header.Get("Accept") != "application/pdf" {
			t.Errorf("expected Accept: application/pdf header")
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Write(expectedPDF)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	pdf, err := client.Downloads.DownloadInvoicePDF(context.Background(), "inv-key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(pdf, expectedPDF) {
		t.Errorf("expected PDF content to match")
	}
}

func TestDownloadsServiceDownloadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "Invoice not found"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	_, err := client.Downloads.DownloadInvoicePDF(context.Background(), "invalid-key")
	if err == nil {
		t.Error("expected error, got nil")
	}

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Errorf("expected APIError, got %T", err)
	}

	if !apiErr.IsNotFound() {
		t.Errorf("expected IsNotFound to be true")
	}
}

func TestUploadsServiceUploadFromReader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/upload") {
			t.Errorf("expected path to end with /upload, got %s", r.URL.Path)
		}

		// Check content type is multipart
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("expected multipart/form-data content type, got %s", contentType)
		}

		// Parse multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse multipart form: %v", err)
		}

		// Check _method field
		if r.FormValue("_method") != "PUT" {
			t.Errorf("expected _method=PUT, got %s", r.FormValue("_method"))
		}

		// Check file was uploaded
		file, header, err := r.FormFile("documents[]")
		if err != nil {
			t.Errorf("failed to get uploaded file: %v", err)
		}
		defer file.Close()

		if header.Filename != "test.pdf" {
			t.Errorf("expected filename 'test.pdf', got '%s'", header.Filename)
		}

		content, _ := io.ReadAll(file)
		if string(content) != "test content" {
			t.Errorf("expected file content 'test content', got '%s'", string(content))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	reader := strings.NewReader("test content")
	err := client.Uploads.UploadDocumentFromReader(context.Background(), "invoices", "inv123", "test.pdf", reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUploadsServiceUploadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message": "Invalid file type"}`))
	}))
	defer server.Close()

	client := NewClient("test-token", WithBaseURL(server.URL))

	reader := strings.NewReader("test content")
	err := client.Uploads.UploadDocumentFromReader(context.Background(), "invoices", "inv123", "test.exe", reader)
	if err == nil {
		t.Error("expected error, got nil")
	}

	apiErr, ok := IsAPIError(err)
	if !ok {
		t.Errorf("expected APIError, got %T", err)
	}

	if !apiErr.IsValidationError() {
		t.Errorf("expected IsValidationError to be true")
	}
}

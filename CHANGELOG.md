# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial SDK release

## [1.0.0] - 2024-01-15

### Added
- Core client with configurable options (timeout, base URL, HTTP client)
- Payments service with full CRUD operations and refund support
- Invoices service with CRUD, bulk actions, and PDF download
- Clients service with CRUD, bulk actions, and merge functionality
- Credits service with CRUD and PDF download
- Payment Terms service with CRUD operations
- Webhooks service with signature verification
- File download support for invoices, credits, quotes, and purchase orders
- Client-side rate limiting with automatic retry logic
- Comprehensive error handling with typed errors
- Generic request method for accessing any API endpoint
- Context support for cancellation and timeouts
- Extensive test coverage (90+ tests)

### Security
- Webhook signature verification using HMAC-SHA256

[Unreleased]: https://github.com/AshkanYarmoradi/go-invoice-ninja/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/AshkanYarmoradi/go-invoice-ninja/releases/tag/v1.0.0

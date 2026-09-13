# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `UsesInclusiveTaxes` field on `Invoice` (#7).
- `WithRateLimiter` and `WithRetryConfig` client options, so any client can use rate limiting and retries.
- `RetryConfig.RetryNonIdempotent` and `APIError.Headers`.
- `WebhookEventHeader`, `WebhookEventQueryParam`, `WebhookSecretHeader` and `MaxWebhookBodyBytes` for configuring webhooks.

### Changed
- `POST` and `PATCH` requests are retried only after a 429 response unless `RetryConfig.RetryNonIdempotent` is set, because the server may already have processed them.
- A 429 response's `Retry-After` header is honored instead of a fixed 60 second wait. If it asks for longer than `MaxBackoff`, the error is returned. A `MaxBackoff` of zero now means no limit.
- `WebhookHandler` reads the event name from the `X-Webhook-Event` header or the `event` query parameter, and checks the secret in the `X-Webhook-Secret` header. The `{"event_type": ..., "data": ...}` body format and HMAC signatures are still accepted.
- `WebhookHandler` accepts `PUT` requests, rejects bodies over 10 MB, returns 400 when the event name is missing, and no longer includes handler error messages in responses.

### Fixed
- Service methods on `RateLimitedClient` now use its rate limiter and retries. Previously only `DoRequestWithRetry` did.
- `DoRequestWithRetry` sends its query parameters instead of dropping them.
- `NewRateLimiter(0)` no longer panics, and a negative `MaxRetries` no longer skips the request.
- `WithTimeout` no longer depends on the option order or changes the `*http.Client` passed to `WithHTTPClient`.
- `WebhookHandler` works with Invoice Ninja webhooks, which contain neither an event name nor a signature.
- `InvoicesService.Download` now downloads the invoice PDF by delegating to `DownloadsService.DownloadInvoicePDF`. It previously always returned a "not implemented" error.
- Documentation now matches the SDK code:
  - The rate limiting, retry and webhook sections of the README and guides describe the new behavior.
  - `docs/api-reference.md` uses the real method signatures, list option field types and model fields, and documents the Downloads and Uploads services, client portal links, `RequestWithQuery` and `SetBaseURL`.
  - The guides and the basic example check errors with `invoiceninja.IsAPIError` instead of type assertions, which miss wrapped errors.
  - The authentication guide no longer shows the non-existent `WithAPISecret` option, and the getting started guide no longer lists a `client.Webhooks` service.
  - `CONTRIBUTING.md` and the `Makefile` help text name the environment variables that the integration tests read.
  - `testdata/README.md`, the invoices example description and the `GetStatement` doc comment describe what the code does.
  - This changelog lists the real release tags.

## [1.1.260118] - 2026-01-17

### Added
- `ContactKey` field on `ClientContact`.

### Changed
- `ClientsService.SwitchToClientPortal` no longer calls `POST /api/v1/client_portal/{clientID}/switch_company`. It fetches the client and builds the portal URL `{baseURL}/client/key_login/{contact_key}` from the contact matching `contactID`, or, when `contactID` is empty, from the primary contact, falling back to the first contact.
- `SwitchToClientPortal` returns an error when the client can't be fetched, has no contacts, or has no contact matching `contactID`, and when the selected contact has no `contact_key`.

## [1.1.260117] - 2026-01-17

### Added
- `ClientsService.SwitchToClientPortal` and `ClientsService.GetClientPortalURL` for generating client portal links, with the `ClientPortalSwitchRequest` and `ClientPortalSwitchResponse` types.

## [1.0.251220] - 2025-12-18

### Added
- Funding links (GitHub Sponsors and Buy Me a Coffee) and a Dependabot configuration for daily Go module updates (#2).
- Issue templates for bug reports, documentation issues and feature requests, a code of conduct, and a security policy (#3).
- Project banner and logo in the README (#4).

## [1.0.251218] - 2025-12-17

### Changed
- The CI workflow uploads coverage with `codecov/codecov-action@v5`, and the release workflow now runs tests with the race detector and uploads coverage to Codecov (#1).

## [1.0.251217] - 2025-12-17

### Added
- `Client` with the `WithBaseURL`, `WithHTTPClient` and `WithTimeout` options and `SetBaseURL`.
- Payments service with list, get, create (optionally emailing a receipt), update, delete, refund, archive, restore, bulk actions and blank payment defaults.
- Invoices service with list, get, create, update, delete, archive, restore, mark paid, mark sent, email, bulk actions and blank invoice defaults. `InvoicesService.Download` was a stub that always returned a "not implemented" error.
- Clients service with list, get, create, update, delete, purge, archive, restore, merge, bulk actions and blank client defaults. `GetStatement` was a stub that always returned a "not implemented" error.
- Credits service with list, get, create, update, delete, archive, restore, mark sent, email, bulk actions and blank credit defaults. It had no PDF download method.
- Payment Terms service with list, get, create, update, delete, archive, restore, bulk actions and blank payment term defaults.
- Downloads service for invoice PDFs, invoice delivery notes, credit PDFs and quote PDFs.
- Uploads service for attaching documents to invoices, payments, clients, credits or any other entity type, from a file path or an `io.Reader`.
- Webhook handler (`NewWebhookHandler`) with per-event callbacks, payload parsing, and HMAC-SHA256 signature verification.
- Client-side rate limiting and retries with exponential backoff, available only through `RateLimitedClient` and its `DoRequestWithRetry` method. The service methods don't use them.
- `APIError` with status helpers, and `IsAPIError` for detecting it.
- Generic `Request` and `RequestWithQuery` methods for other API endpoints.
- `context.Context` support in all API methods.
- Documentation, runnable examples, unit tests, and integration tests behind the `integration` build tag.

### Security
- Webhook signature verification using HMAC-SHA256.

[Unreleased]: https://github.com/AshkanYarmoradi/go-invoice-ninja/compare/v1.1.260118...HEAD
[1.1.260118]: https://github.com/AshkanYarmoradi/go-invoice-ninja/compare/v1.1.260117...v1.1.260118
[1.1.260117]: https://github.com/AshkanYarmoradi/go-invoice-ninja/compare/v1.0.251220...v1.1.260117
[1.0.251220]: https://github.com/AshkanYarmoradi/go-invoice-ninja/compare/v1.0.251218...v1.0.251220
[1.0.251218]: https://github.com/AshkanYarmoradi/go-invoice-ninja/compare/v1.0.251217...v1.0.251218
[1.0.251217]: https://github.com/AshkanYarmoradi/go-invoice-ninja/releases/tag/v1.0.251217

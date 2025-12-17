# Contributing to Go Invoice Ninja SDK

First off, thank you for considering contributing to the Go Invoice Ninja SDK! 🎉

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (code snippets, API responses)
- **Describe the behavior you observed and what you expected**
- **Include your Go version** (`go version`)
- **Include the SDK version** you're using

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description of the proposed functionality**
- **Explain why this enhancement would be useful**
- **List any alternative solutions you've considered**

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Write tests** for any new functionality
3. **Ensure all tests pass** (`make test`)
4. **Run the linter** (`make lint`)
5. **Update documentation** if needed
6. **Write a clear commit message**

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make (optional, but recommended)
- golangci-lint (for linting)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/go-invoice-ninja.git
cd go-invoice-ninja

# Add upstream remote
git remote add upstream https://github.com/AshkanYarmoradi/go-invoice-ninja.git

# Install dependencies
go mod download

# Run tests
make test

# Run linter
make lint
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run tests with race detector
make test-race

# Run integration tests (requires API token)
INVOICE_NINJA_TOKEN=your-token make test-integration
```

### Code Style

This project uses `golangci-lint` for code quality. Please ensure your code passes all linter checks:

```bash
make lint
```

Key style guidelines:
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Keep functions focused and under 60 statements
- Add comments for exported functions, types, and constants
- Use meaningful variable names
- Handle errors explicitly

### Commit Messages

We follow conventional commits. Each commit message should be structured as:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code changes that neither fix bugs nor add features
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(payments): add support for bulk refunds
fix(client): handle nil response body
docs: update README with webhook examples
```

## Project Structure

```
go-invoice-ninja/
├── .github/workflows/    # CI/CD workflows
├── docs/                 # Additional documentation
├── examples/             # Runnable examples
├── testdata/             # Test fixtures
│
├── client.go            # Main client implementation
├── clients.go           # Clients service
├── credits.go           # Credits service
├── errors.go            # Error types
├── files.go             # File operations
├── invoices.go          # Invoices service
├── models.go            # Data models
├── payments.go          # Payments service
├── payment_terms.go     # Payment terms service
├── retry.go             # Retry and rate limiting
├── webhooks.go          # Webhook handling
└── *_test.go            # Test files
```

## API Design Guidelines

When adding new functionality:

1. **Follow existing patterns** - Look at how similar features are implemented
2. **Support pagination** - Use `ListOptions` for list operations
3. **Return typed errors** - Use `APIError` for API errors
4. **Support context** - All operations should accept `context.Context`
5. **Add tests** - Unit tests and integration tests where applicable

### Adding a New Service

1. Create `service_name.go` with the service struct and methods
2. Create `service_name_test.go` with unit tests
3. Add the service to `Client` struct in `client.go`
4. Initialize the service in `NewClient`
5. Add models to `models.go` if needed
6. Update `README.md` with usage examples

## Questions?

Feel free to open an issue with your question or reach out to the maintainers.

Thank you for contributing! 🙏

# Makefile for Go Invoice Ninja SDK

.PHONY: all build test test-race test-integration coverage lint fmt vet clean help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint

# Build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Default target
all: lint test build

## build: Build the package
build:
	@echo "Building..."
	$(GOBUILD) ./...

## test: Run unit tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	$(GOTEST) -v -race ./...

## test-integration: Run integration tests (requires INVOICE_NINJA_TOKEN)
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./...

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## coverage-text: Show coverage in terminal
coverage-text:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -func=coverage.out

## lint: Run linter
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

## tidy: Tidy and verify go modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	$(GOMOD) verify

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f coverage.out coverage.html
	$(GOCMD) clean -cache -testcache

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## pre-commit: Run before committing
pre-commit: tidy fmt lint test

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## docs: Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all . > docs/api.txt

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

# Default help
.DEFAULT_GOAL := help

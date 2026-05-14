SHELL := /bin/bash

PKGS := ./...
GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './execution/contracts/*')
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || true)

.PHONY: test vet lint check-format format generate-swagger init-db migrate ci

# Run the full test suite.
test:
	go test $(PKGS)

# Run go vet on all packages.
vet:
	go vet $(PKGS)

# Verify Go formatting.
check-format:
	@echo "Checking Go formatting..."
	@unformatted=$$(gofmt -l $(GOFILES)); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Format all Go files in place.
format:
	gofmt -s -w $(GOFILES)

# Generate Swagger/OpenAPI docs from annotations.
generate-swagger:
	swag init -g ./cmd/api/main.go --parseDependency -o docs

# Initialize the MongoDB schema and indexes required by the API.
init-db:
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL is not set"; exit 1; \
	fi
	go run ./cmd/db-init

# Apply MongoDB migrations.
migrate: init-db

# Run static analysis and vet.
lint: vet
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint is not installed."; \
		echo "Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run ./...

# Continuous integration target used by GitHub Actions.
ci: check-format lint test

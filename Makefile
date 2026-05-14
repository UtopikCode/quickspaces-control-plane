SHELL := /bin/bash

PKGS := ./...
GOFILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './execution/contracts/*')
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || true)

.PHONY: test vet lint check-format format generate-swagger generate-ent migrate-ent generate-ent-ddl init-db ci

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

# Generate Ent entity code from schema definitions.
generate-ent:
	go run -mod=mod entgo.io/ent/cmd/ent generate ./ent/schema

# Generate the current DDL for the Ent schema.
generate-ent-ddl:
	go run -mod=mod entgo.io/ent/cmd/ent schema ./ent/schema --dialect postgres --version 15

# Apply Ent schema changes to the database.
migrate-ent:
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL is not set"; exit 1; \
	fi
	go run ./cmd/ent-migrate

# Initialize the PostgreSQL schema for the API.
init-db:
	@echo "Initializing PostgreSQL schema from db/schema.sql"
	@if [ -z "$$DATABASE_URL" ]; then \
		echo "DATABASE_URL is not set"; exit 1; \
	fi
	psql "$$DATABASE_URL" -f db/schema.sql

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

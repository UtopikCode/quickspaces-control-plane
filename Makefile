SHELL := /bin/bash

PKGS := ./...
GO_FILES := $(shell find . -type f -name '*.go' -not -path './.git/*' -not -path './execution/contracts/*')
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || true)

.PHONY: all test go-vet lint check-format format ci

all: ci

# Run the full test suite.
test:
	go test $(PKGS)

# Run go vet on all packages.
go-vet:
	go vet $(PKGS)

# Verify Go formatting.
check-format:
	@echo "Checking Go formatting..."
	@unformatted=$$(gofmt -l $(GO_FILES)); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files need gofmt:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Format all Go files in place.
format:
	gofmt -s -w $(GO_FILES)

# Run static analysis and vet.
lint: go-vet
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint is not installed."; \
		echo "Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	$(GOLANGCI_LINT) run ./...

# Continuous integration target used by GitHub Actions.
ci: check-format lint test

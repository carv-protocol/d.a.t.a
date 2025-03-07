# Variables
BINARY_NAME=d.a.t.a
MAIN_PATH=./src/cmd/agent  # Adjust this to your main.go location

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOFMT=gofmt

# Find golangci-lint and gci in PATH or use default location
GOPATH ?= $(shell go env GOPATH)
GOLINT ?= $(shell which golangci-lint 2>/dev/null || echo $(GOPATH)/bin/golangci-lint)
GCI ?= $(shell which gci 2>/dev/null || echo $(GOPATH)/bin/gci)

# Build flags
LDFLAGS=-ldflags "-w -s"

.PHONY: all build run test clean vet fmt tidy help lint lint-install gci gci-install

# Default target
all: build

# Build the application
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	$(GORUN) $(MAIN_PATH)

# Run tests
test:
	$(GOTEST) -v ./...

# Clean build files
clean:
	rm -f $(BINARY_NAME)
	go clean

# Run go vet
vet:
	$(GOVET) ./...

# Format code
fmt:
	$(GOFMT) -w .

# Format imports with gci
gci:
	@echo "Formatting imports with gci..."
	@if [ ! -x "$(GCI)" ]; then \
		echo "gci not found, installing..."; \
		$(MAKE) gci-install; \
	fi
	@echo "Using gci: $(GCI)"
	@$(GCI) --version
	@find . -name "*.go" -not -path "./vendor/*" -print0 | xargs -0 $(GCI) write --skip-generated -s "standard" -s "prefix(github.com/carv-protocol/d.a.t.a)" -s "default" --custom-order
	@echo "✅ Import formatting completed"

# Install gci
gci-install:
	@echo "Installing gci..."
	@$(GOCMD) install github.com/daixiang0/gci@latest

# Tidy and verify dependencies
tidy:
	$(GOMOD) tidy
	$(GOMOD) verify

# Install golangci-lint
lint-install:
	@echo "Installing golangci-lint..."
	@$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@if [ ! -x "$(GOLINT)" ]; then \
		echo "golangci-lint not found, installing..."; \
		$(MAKE) lint-install; \
	fi
	@$(GOLINT) run --timeout=5m ./...

# Run golangci-lint on a specific file or directory
lint-file:
	@echo "Running golangci-lint on $(FILE)..."
	@$(GOLINT) run $(FILE)

# Development: format, vet, lint, and test
dev: fmt gci vet lint test

# Help target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build files"
	@echo "  vet          - Run go vet"
	@echo "  fmt          - Format code"
	@echo "  gci          - Format imports with gci"
	@echo "  gci-install  - Install gci"
	@echo "  tidy         - Tidy and verify dependencies"
	@echo "  lint         - Run golangci-lint on entire project"
	@echo "  lint-file    - Run golangci-lint on a specific file (usage: make lint-file FILE=path/to/file.go)"
	@echo "  lint-install - Install golangci-lint"
	@echo "  dev          - Run format, vet, lint, and test"
	@echo "  help         - Show this help message"
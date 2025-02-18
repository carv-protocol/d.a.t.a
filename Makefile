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

# Build flags
LDFLAGS=-ldflags "-w -s"

.PHONY: all build run test clean vet fmt tidy help

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

# Tidy and verify dependencies
tidy:
	$(GOMOD) tidy
	$(GOMOD) verify

# Development: format, vet, and test
dev: fmt vet test

# Help target
help:
	@echo "Available targets:"
	@echo "  build   - Build the application"
	@echo "  run     - Run the application"
	@echo "  test    - Run tests"
	@echo "  clean   - Clean build files"
	@echo "  vet     - Run go vet"
	@echo "  fmt     - Format code"
	@echo "  tidy    - Tidy and verify dependencies"
	@echo "  dev     - Run format, vet, and test"
	@echo "  help    - Show this help message"
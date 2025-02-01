.PHONY: all build test integration clean lint fmt vet sec deps help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=llm

# Tool versions
GOLANGCI_LINT_VERSION=v1.55.2

# Test flags
TEST_FLAGS=-race -v
INTEGRATION_FLAGS=-tags=integration
COVERAGE_FLAGS=-coverprofile=coverage.out

# Directories
ALL_PACKAGES=$(shell go list ./...)
TEST_PACKAGES=$(shell go list ./... | grep -v /integration)
INTEGRATION_PACKAGES=$(shell go list ./... | grep /integration)

all: deps lint test build

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

# Run all tests except integration tests
test:
	$(GOTEST) $(TEST_FLAGS) $(COVERAGE_FLAGS) $(TEST_PACKAGES)

# Run only integration tests
integration:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Create one with your API keys first."; \
		exit 1; \
	fi
	$(GOTEST) $(TEST_FLAGS) $(INTEGRATION_FLAGS) $(ALL_PACKAGES)

# Run all tests including integration
test-all: test integration

# Generate test coverage report
coverage: test
	$(GOCMD) tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	go clean -testcache

# Format code
fmt:
	go fmt $(ALL_PACKAGES)

# Run go vet
vet:
	go vet $(ALL_PACKAGES)

# Install golangci-lint if not present
install-lint:
	@which golangci-lint > /dev/null 2>&1 || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

# Run golangci-lint
lint: install-lint
	golangci-lint run

# Run security check
sec:
	$(GOGET) -u github.com/securego/gosec/v2/cmd/gosec
	gosec ./...

# Update dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOMOD) verify

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOGET) -u golang.org/x/tools/cmd/godoc
	@echo "Run: godoc -http=:6060"
	@echo "Then visit: http://localhost:6060/pkg/github.com/ksred/llm/"

# Show help
help:
	@echo "Available targets:"
	@echo "  all          - Run deps, lint, test, and build"
	@echo "  build        - Build the application"
	@echo "  test         - Run unit tests"
	@echo "  integration  - Run integration tests (requires .env)"
	@echo "  test-all     - Run all tests including integration"
	@echo "  coverage     - Generate test coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo "  fmt          - Format code"
	@echo "  vet         - Run go vet"
	@echo "  lint        - Run golangci-lint"
	@echo "  sec         - Run security check"
	@echo "  deps        - Update dependencies"
	@echo "  docs        - Generate documentation"
	@echo "  help        - Show this help message"

# Default target
default: help

.PHONY: build test clean install fmt vet lint

# Project configuration
BINARY_NAME=btrfs-read
GO=go
GOFLAGS=-v

# Build directory
BUILD_DIR=build

# Default target
all: build

# Build CLI tool
build: fmt vet
	@echo "Building CLI tool..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/btrfs-read

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

# Code coverage
coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Static checks
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Lint (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Install binaries
install: build
	@echo "Installing..."
	$(GO) install ./cmd/btrfs-read

# Run example
run-example: build
	@echo "Running example..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Create test image
create-test-image:
	@echo "Creating test image (requires root)..."
	@if [ ! -f tests/create-test-image.sh ]; then \
		echo "Error: tests/create-test-image.sh not found"; \
		exit 1; \
	fi
	@chmod +x tests/create-test-image.sh
	sudo ./tests/create-test-image.sh

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@if [ ! -f tests/testdata/test.img ]; then \
		echo "Error: Test image not found. Run 'make create-test-image' first."; \
		exit 1; \
	fi
	$(GO) test -v ./tests/integration/... -coverprofile=coverage-integration.out

# Test CLI tool
test-cli: build
	@echo "Testing CLI tool with test image..."
	@if [ ! -f tests/testdata/test.img ]; then \
		echo "Error: Test image not found. Run 'make create-test-image' first."; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY_NAME) ls tests/testdata/test.img /

# Help
help:
	@echo "Available targets:"
	@echo "  build              - Build CLI tool"
	@echo "  test               - Run all tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-cli           - Test CLI tool with test image"
	@echo "  bench              - Run benchmarks"
	@echo "  coverage           - Generate coverage report"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  lint               - Run linter"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Download dependencies"
	@echo "  install            - Install binaries"
	@echo "  create-test-image  - Create test btrfs image (requires root)"

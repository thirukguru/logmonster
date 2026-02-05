.PHONY: build test install clean lint build-all run

# Binary name
BINARY_NAME=logmonster
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default target
all: build

# Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/logmonster

# Run without building
run:
	$(GOCMD) run ./cmd/logmonster $(ARGS)

# Run tests with race detection and coverage
test:
	$(GOTEST) -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Install to GOPATH/bin
install:
	$(GOCMD) install ./cmd/logmonster

# Build for multiple platforms
build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/logmonster
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/logmonster

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run go vet
vet:
	$(GOVET) ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  run         - Run without building (use ARGS='scan' to pass arguments)"
	@echo "  test        - Run tests with race detection"
	@echo "  test-coverage - Run tests with HTML coverage report"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  build-all   - Build for linux amd64 and arm64"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  vet         - Run go vet"
	@echo "  lint        - Run golangci-lint"
	@echo "  clean       - Remove build artifacts"

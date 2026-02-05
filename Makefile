.PHONY: build test install clean lint build-all run release snapshot

# Binary name
BINARY_NAME=logmonster
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Build flags with version injection
LDFLAGS=-ldflags "-s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)"

# Default target
all: build

# Build the binary with version info
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/logmonster
	@echo "Built $(VERSION) ($(COMMIT))"

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
	$(GOCMD) install $(LDFLAGS) ./cmd/logmonster

# Build for multiple platforms
build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/logmonster
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/logmonster
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/logmonster
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/logmonster
	@echo "Built all platforms for $(VERSION)"

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

# Release using GoReleaser (requires goreleaser)
release:
	goreleaser release --clean

# Snapshot release (no publish)
snapshot:
	goreleaser release --snapshot --clean

# Help
help:
	@echo "Log Monster Detector - Build Targets"
	@echo ""
	@echo "  build       - Build the binary with version info"
	@echo "  run         - Run without building (use ARGS='scan' to pass arguments)"
	@echo "  test        - Run tests with race detection"
	@echo "  test-coverage - Run tests with HTML coverage report"
	@echo "  install     - Install to GOPATH/bin"
	@echo "  build-all   - Build for linux/darwin amd64/arm64"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  vet         - Run go vet"
	@echo "  lint        - Run golangci-lint"
	@echo "  clean       - Remove build artifacts"
	@echo "  release     - Create release with GoReleaser"
	@echo "  snapshot    - Create snapshot release (no publish)"
	@echo ""
	@echo "Current version: $(VERSION)"

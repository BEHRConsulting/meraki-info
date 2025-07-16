# Meraki Info Makefile

.PHONY: help build test clean run install deps

# Default target
help:
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  test      - Run all tests"
	@echo "  test-v    - Run tests with verbose output"
	@echo "  clean     - Clean build artifacts"
	@echo "  run       - Build and run with help flag"
	@echo "  access    - Show available organizations and networks (requires MERAKI_APIKEY)"
	@echo "  install   - Install the application to GOPATH/bin"
	@echo "  deps      - Download and tidy dependencies"
	@echo "  coverage  - Run tests with coverage report"

# Build the application
build:
	go build -o meraki-info .

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage
coverage:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -f meraki-info
	go clean

# Build and run with help
run: build
	./meraki-info --help

# Show access information (requires MERAKI_APIKEY environment variable)
access: build
	@if [ -z "$$MERAKI_APIKEY" ]; then \
		echo "Error: MERAKI_APIKEY environment variable is required"; \
		echo "Usage: MERAKI_APIKEY=your-key make access"; \
		exit 1; \
	fi
	./meraki-info --access --APIKey "$$MERAKI_APIKEY"

# Install to GOPATH/bin
install:
	go install .

# Download and organize dependencies
deps:
	go mod download
	go mod tidy

# Cross-compilation targets
build-linux:
	GOOS=linux GOARCH=amd64 go build -o meraki-info-linux .

build-windows:
	GOOS=windows GOARCH=amd64 go build -o meraki-info.exe .

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o meraki-info-mac .

# Build for all platforms
build-all: build-linux build-windows build-mac

# Build for all platforms
build-all: build-linux build-windows build-mac

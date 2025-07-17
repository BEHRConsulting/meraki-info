# Meraki Info Makefile

.PHONY: help build test test-v coverage clean run access install deps build-linux build-linux-arm build-windows build-mac build-mac-arm build-all

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build the application for current platform"
	@echo "  test          - Run all tests"
	@echo "  test-v        - Run tests with verbose output"
	@echo "  coverage      - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run with help flag"
	@echo "  access        - Show available organizations and networks (requires MERAKI_APIKEY)"
	@echo "  install       - Install the application to GOPATH/bin"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  build-linux   - Build for Linux (amd64)"
	@echo "  build-linux-arm - Build for Linux (arm64)"
	@echo "  build-windows - Build for Windows (amd64)"
	@echo "  build-mac     - Build for macOS (amd64 - Intel)"
	@echo "  build-mac-arm - Build for macOS (arm64 - Apple Silicon)"
	@echo "  build-all     - Build for all platforms and architectures"

# Build the application for current platform
build:
	@echo "Building for current platform..."
	@if [ "$(shell uname -s 2>/dev/null)" = "Darwin" ]; then \
		if [ "$(shell uname -m 2>/dev/null)" = "arm64" ]; then \
			go build -o meraki-info .; \
			echo "✅ Build completed for macOS ARM64"; \
		else \
			go build -o meraki-info .; \
			echo "✅ Build completed for macOS Intel"; \
		fi \
	elif [ "$(shell uname -s 2>/dev/null | cut -c1-5)" = "Linux" ]; then \
		if [ "$(shell uname -m 2>/dev/null)" = "aarch64" ] || [ "$(shell uname -m 2>/dev/null)" = "arm64" ]; then \
			go build -o meraki-info .; \
			echo "✅ Build completed for Linux ARM64"; \
		else \
			go build -o meraki-info .; \
			echo "✅ Build completed for Linux AMD64"; \
		fi \
	else \
		go build -o meraki-info.exe .; \
		echo "✅ Build completed for Windows"; \
	fi

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
	@echo "Cleaning build artifacts..."
	@if command -v rm >/dev/null 2>&1; then \
		rm -f meraki-info meraki-info.exe meraki-info-linux meraki-info-linux-arm meraki-info-mac meraki-info-mac-arm; \
	elif command -v del >/dev/null 2>&1; then \
		del /f /q meraki-info.exe meraki-info-linux meraki-info-linux-arm meraki-info-mac meraki-info-mac-arm 2>nul || true; \
	else \
		echo "Manual cleanup required: remove meraki-info* files"; \
	fi
	go clean

# Build and run with help
run: build
	@if [ -f "./meraki-info.exe" ]; then \
		./meraki-info.exe --help; \
	elif [ -f "./meraki-info" ]; then \
		./meraki-info --help; \
	else \
		echo "No executable found to run"; \
		exit 1; \
	fi

# Show access information (requires MERAKI_APIKEY environment variable)
access: build
	@if [ -z "$$MERAKI_APIKEY" ]; then \
		echo "Error: MERAKI_APIKEY environment variable is required"; \
		echo "Usage: MERAKI_APIKEY=your-key make access"; \
		exit 1; \
	fi
	@if [ -f "./meraki-info.exe" ]; then \
		./meraki-info.exe access --apikey "$$MERAKI_APIKEY"; \
	elif [ -f "./meraki-info" ]; then \
		./meraki-info access --apikey "$$MERAKI_APIKEY"; \
	else \
		echo "No executable found to run"; \
		exit 1; \
	fi

# Install to GOPATH/bin
install:
	go install .

# Download and organize dependencies
deps:
	go mod download
	go mod tidy

# Cross-compilation targets
build-linux:
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o meraki-info-linux .
	@echo "✅ Linux AMD64 build completed: meraki-info-linux"

build-linux-arm:
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o meraki-info-linux-arm .
	@echo "✅ Linux ARM64 build completed: meraki-info-linux-arm"

build-windows:
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o meraki-info.exe .
	@echo "✅ Windows AMD64 build completed: meraki-info.exe"

build-mac:
	@echo "Building for macOS (amd64 - Intel)..."
	GOOS=darwin GOARCH=amd64 go build -o meraki-info-mac .
	@echo "✅ macOS Intel build completed: meraki-info-mac"

build-mac-arm:
	@echo "Building for macOS (arm64 - Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 go build -o meraki-info-mac-arm .
	@echo "✅ macOS Apple Silicon build completed: meraki-info-mac-arm"

# Build for all platforms and architectures
build-all: build-linux build-linux-arm build-windows build-mac build-mac-arm
	@echo "🎉 All platform builds completed!"
	@echo "📦 Built files:"
	@ls -la meraki-info* 2>/dev/null || dir meraki-info* 2>nul || echo "  Check directory for meraki-info* files"

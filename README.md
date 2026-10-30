# Meraki Info

This project is a Golang application that collects Meraki network information.

`meraki-info` is a read-only command-line tool for the Cisco Meraki Dashboard API v1. With an
API key it lists the organizations and networks the key can reach (`access`), and reports route
tables (`route-tables`), licenses (`licenses`), devices that are down (`down`) and devices that
are alerting (`alerting`) as text, JSON, XML or CSV.

For contributors and AI agents, [AGENTS.md](AGENTS.md) has the layout, commands and conventions;
[WORKLOG.md](WORKLOG.md), [VERSIONS.md](VERSIONS.md) and [features.yaml](features.yaml) record
what was done, what shipped and what exists.

## AI Generated (mostly) 
- This was an experiment using Github's Copilot in agent mode and Claude Sonnet 4 AI model.
- I kept tweeking the prompts to get the functions I needed.
- The prompts I used the generate this app are on the the file prompts-initial.md.
- The prompts Claude Sonnet 4 created from prompts-initial.md are in file prompts-claude.md.

## Features

- **Cross-Platform Support**: Native builds for Windows, Linux, macOS (Intel & Apple Silicon)
- **ARM64 Compatibility**: Full support for Apple Silicon Macs and ARM64 Linux servers  
- **Enhanced Build System**: PowerShell scripts for Windows-native development experience
- **Authentication**: API key (`-apikey` or `MERAKI_APIKEY`). The client package also has an OAuth2 constructor, but the command line does not expose it.
- **Flexible Configuration**: Command line flags and environment variables
- **Multiple Output Formats**: Text, JSON, XML, and CSV
- **Structured Logging**: Configurable log levels (debug, info, error)
- **Error Handling**: Graceful error handling with clear messages
- **Retries**: Rate-limit (429) and server (5xx) responses are retried up to 3 times with exponential backoff
- **Unit Tests**: Config, API client (against local test servers) and output writers; about 40-55% statement coverage per package, none for `main`
- **Multiple Build Options**: Traditional Makefile, PowerShell scripts, and batch files

## Installation

### Prerequisites
- Go 1.24.2 or later (the `go` line in `go.mod`)
- Meraki Dashboard API access
- **Windows**: PowerShell 5.1+ (built-in with Windows 10/11)
- **Cross-platform**: make (optional, for traditional Makefile usage)

### Build from Source

#### Quick Start (Any Platform)
```bash
git clone <repository-url>
cd meraki-info
go mod tidy
go build -o meraki-info
```

#### Using Build Scripts (Recommended)

**Windows (PowerShell):**
```powershell
# Simple build for Windows
.\build.ps1

# Build for all platforms (Windows, Linux, macOS, ARM64)
.\build.ps1 -All

# Using the full make system
.\make.ps1 build-all
```

**Traditional Make (Linux/macOS/Windows with make):**
```bash
# Build for current platform
make build

# Build for all platforms including ARM64
make build-all

# Run tests
make test
```

**See [BUILD_SCRIPTS.md](BUILD_SCRIPTS.md) for comprehensive build documentation.**

## Available Platforms

This application can be built for multiple platforms and architectures:

| Platform | Architecture | Build Command | Output File | Size |
|----------|--------------|---------------|-------------|------|
| Windows | AMD64 | `.\make.ps1 build-windows` | `meraki-info.exe` | ~9.4 MB |
| Linux | AMD64 | `.\make.ps1 build-linux` | `meraki-info-linux` | ~9.3 MB |
| Linux | ARM64 | `.\make.ps1 build-linux-arm` | `meraki-info-linux-arm` | ~8.8 MB |
| macOS | AMD64 (Intel) | `.\make.ps1 build-mac` | `meraki-info-mac` | ~9.3 MB |
| macOS | ARM64 (Apple Silicon) | `.\make.ps1 build-mac-arm` | `meraki-info-mac-arm` | ~8.8 MB |

**Build all platforms at once:**
```powershell
# PowerShell (Windows)
.\make.ps1 build-all

# Traditional make (any platform)
make build-all

# WSL2 (Windows Subsystem for Linux)
make build-all              # Builds all platforms including Windows .exe
```

## Usage

### Command Line Options and Arguments

| Option | Environment Variable | Description | Required |
|------|---------------------|-------------|----------|
| `-apikey` | `MERAKI_APIKEY` | Meraki API key | Yes |
| `-org` | `MERAKI_ORG` | Meraki organization ID | Yes* |
| `-network` | `MERAKI_NET` | Specific network ID or name (optional) | No |
| `-output` | - | Output file path | No (default: stdout) |
| `-format` | - | Output format: text, json, xml, csv | No (default: text) |
| `-loglevel` | - | Log level: debug, info, error | No (default: error) |
| `-all` | - | Process every network (see below). Implied whenever `-network` is not given, except for `access` | No |

**Commands (positional arguments):**
- `access` - Show available organizations and networks
- `route-tables` - Output route tables
- `licenses` - Output license information  
- `down` - Output all devices that are down/offline (offline, alerting, dormant, down, unreachable or disconnected)
- `alerting` - Output all devices that are alerting

*Organization is not required when using `access` command.
*The `-all` and `-network` options cannot be used together.

**Options must come before the command.** Go's flag parser stops at the first non-option
argument, so `meraki-info access -apikey KEY` fails with "only one command is allowed"; write
`meraki-info -apikey KEY access`.

**What `-all` does.** Without `-network`, every network is processed:

- **To stdout** (no `-output`, or `-output -`): one consolidated report whose records carry the
  network and organization. It covers the `-org` organization, or every organization the key
  can see when `-org` is not given.
- **To a file** (`-output FILE`): each network is processed in turn and written to that same
  `FILE`, so the file ends up holding only the last network's results. (Timestamped per-network
  file names existed until commit `f376ff6` removed the `-output default` option.)
- In consolidated reports, `-format xml` and `-format csv` for `licenses`, `down` and `alerting`
  produce indented JSON, not XML or CSV.

### Examples

#### Basic usage with API key
```bash
./meraki-info -apikey your-api-key -org your-org-id route-tables
```

#### Check available organizations and networks
```bash
# Show all accessible organizations and networks
./meraki-info -apikey your-api-key access

# Show networks for a specific organization only
./meraki-info -apikey your-api-key -org "123456" access
./meraki-info -apikey your-api-key -org "Your Organization" access
```

#### Using environment variables
```bash
export MERAKI_APIKEY="your-api-key"
export MERAKI_ORG="your-org-id"
./meraki-info route-tables
```

#### Output license information
```bash
./meraki-info -apikey your-api-key -org your-org-id licenses
```

#### Output down devices
```bash
./meraki-info -apikey your-api-key -org your-org-id down
```

#### Get info for specific network to JSON
```bash
./meraki-info -apikey your-api-key -org your-org-id -network net-id -output routes.json -format json route-tables
```

#### Get info for all networks (consolidated, to stdout)
```bash
# Every network in the organization, one consolidated text report
./meraki-info -apikey your-api-key -org your-org-id -all route-tables

# The same as JSON
./meraki-info -apikey your-api-key -org your-org-id -all -format json route-tables

# The same as CSV
./meraki-info -apikey your-api-key -org your-org-id -all -format csv route-tables
```

#### Network identification
```bash
# Use network ID
./meraki-info -apikey your-api-key -org your-org-id -network "L_123456789012345678" route-tables

# Use network name (must be unique within organization)
./meraki-info -apikey your-api-key -org your-org-id -network "Main Network" route-tables
```

#### Custom output filename
```bash
# Specify custom output filename
./meraki-info -apikey your-api-key -org your-org-id -network "Main Network" -output my-routes.txt route-tables

# Output to stdout (useful for piping)
./meraki-info -apikey your-api-key -org your-org-id -network "Main Network" -output "-" route-tables

# Pipe JSON output to jq for processing
./meraki-info -apikey your-api-key -org your-org-id -network "Main Network" -output "-" -format json route-tables | jq '.[] | .name'

# Save CSV output with custom processing
./meraki-info -apikey your-api-key -org your-org-id -network "Main Network" -output "-" -format csv route-tables > processed-routes.csv
```

#### Enable debug logging
```bash
./meraki-info -apikey your-api-key -org your-org-id -loglevel debug route-tables
```

## Authentication

### API Key (Recommended for scripts)
1. Log in to the Meraki Dashboard
2. Navigate to Organization > Settings > Dashboard API access
3. Generate an API key
4. Use the key with the `-apikey` option or `MERAKI_APIKEY` environment variable

### OAuth2 (not available from the command line)
`internal/meraki` has a `NewClientWithOAuth2` constructor that wraps an already-issued access
token, but nothing calls it and there are no OAuth2 options. Use an API key.

## Output Formats

### Text (Default)
Human-readable text format with route details.

### JSON
Structured JSON format suitable for programmatic processing.

### XML
XML format with proper structure and encoding.

### CSV
Comma-separated values format for spreadsheet applications.

## File Naming

Output goes to **stdout** unless `-output` names a file; `-output -` also means stdout. With
`-output FILE` the exact name given is used, and an existing file is overwritten:

```bash
./meraki-info -apikey key -org 123 -network "City Core" -output "my-info.txt" route-tables
# Creates: my-info.txt
```

There is no automatic file naming. Timestamped
`<CommandType>-<OrganizationName>-<NetworkName>-<RFC3339 datetime>.<extension>` names were
generated until commit `f376ff6` ("Remove --alerting --output default") took that option out
(see "What `-all` does" above).

### Stdout Output
When `-output "-"` is specified, the output is sent to stdout instead of a file. This enables:

**Piping to other tools:**
```bash
# Extract route names using jq
./meraki-info -org 123 -network "Main" -output "-" -format json route-tables | jq '.[] | .name'

# Count total routes
./meraki-info -org 123 -network "Main" -output "-" -format json route-tables | jq '. | length'

# Filter enabled routes only
./meraki-info -org 123 -network "Main" -output "-" -format json route-tables | jq '.[] | select(.enabled == true)'

# Get license information in JSON
./meraki-info -org 123 -output "-" -format json licenses | jq '.[] | select(.state == "active")'

# Check down devices
./meraki-info -org 123 -output "-" -format json down | jq '.[] | .name'
```

**Processing CSV data:**
```bash
# Save to file with custom name
./meraki-info -org 123 -network "Main" -output "-" -format csv route-tables > custom-name.csv

# Process with awk
./meraki-info -org 123 -network "Main" -output "-" -format csv route-tables | awk -F',' '{print $2, $3}'
```

**Integration with scripts:**
```bash
#!/bin/bash
ROUTES=$(./meraki-info -org 123 -network "Main" -output "-" -format json route-tables)
echo "$ROUTES" | jq '.[] | select(.subnet | contains("192.168"))'
```

## Configuration

### Environment Variables
- `MERAKI_APIKEY`: Your Meraki API key
- `MERAKI_ORG`: Organization ID
- `MERAKI_NET`: Network ID (optional)

### Configuration Priority
1. Command line options (highest priority)
2. Environment variables
3. Default values (lowest priority)

## Development

### Project Structure
```
meraki-info/
├── main.go                     # Application entry point
├── internal/
│   ├── config/                 # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── logger/                 # Logging configuration
│   │   └── logger.go
│   ├── meraki/                 # Meraki API client
│   │   ├── client.go
│   │   ├── client_test.go
│   │   └── retry_test.go
│   └── output/                 # Output formatters
│       ├── writer.go
│       ├── writer_test.go
│       └── orgid_test.go
├── build.ps1                   # Simple PowerShell build script
├── build-windows.ps1           # Advanced PowerShell build script
├── make.ps1                    # Full PowerShell make system
├── make.bat                    # Batch wrapper for Command Prompt
├── PowerShell-Make-Function.ps1 # Global PowerShell make function
├── Makefile                    # Traditional make system (cross-platform)
├── BUILD_SCRIPTS.md            # Comprehensive build documentation
├── MAKE_SUPPORT.md             # PowerShell make system documentation
├── AGENTS.md                   # Guide for contributors and AI agents
├── WORKLOG.md / VERSIONS.md    # What was done and why / what shipped
├── features.yaml               # Feature status manifest
├── docs/code-review/           # Dated code reviews
├── go.mod
├── go.sum
└── README.md
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

**Using build scripts:**
```powershell
# PowerShell
.\make.ps1 test              # Run tests
.\make.ps1 test-v            # Verbose tests
.\make.ps1 coverage          # With coverage

# Traditional make
make test                    # Run tests
make test-v                  # Verbose tests
make coverage                # With coverage
```

### Building

#### Simple Build Commands
```bash
# Build for current platform
go build -o meraki-info

# Manual cross-compilation
GOOS=linux GOARCH=amd64 go build -o meraki-info-linux
GOOS=windows GOARCH=amd64 go build -o meraki-info.exe
GOOS=darwin GOARCH=arm64 go build -o meraki-info-mac-arm  # Apple Silicon
```

#### Enhanced Build System

**PowerShell (Windows - Recommended):**
```powershell
# Quick builds
.\build.ps1                  # Current platform
.\build.ps1 -All             # All platforms
.\build.ps1 -Target linux    # Specific platform

# Full make system
.\make.ps1 build             # Current platform
.\make.ps1 build-all         # All platforms including ARM64
.\make.ps1 build-mac-arm     # Apple Silicon specifically
.\make.ps1 build-linux-arm   # ARM64 Linux

# Command Prompt compatible
make.bat build
make.bat build-all
```

**WSL2 (Windows Subsystem for Linux):**
```bash
# Build Windows executable from WSL2
GOOS=windows GOARCH=amd64 go build -o meraki-info.exe

# Build all platforms from WSL2
make build-all               # Uses traditional Makefile

# Build specific Windows target
make build-windows           # Creates meraki-info.exe

# Cross-compile from WSL2 to Windows
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o meraki-info.exe
```

**Traditional Make (Cross-platform):**
```bash
make build                   # Current platform
make build-all               # All platforms including ARM64
make build-windows           # Windows x64
make build-linux             # Linux x64
make build-linux-arm         # Linux ARM64
make build-mac               # macOS Intel
make build-mac-arm           # macOS Apple Silicon
make clean                   # Clean build artifacts
```

**See [BUILD_SCRIPTS.md](BUILD_SCRIPTS.md) and [MAKE_SUPPORT.md](MAKE_SUPPORT.md) for detailed build documentation.**

## Building with WSL2

Windows Subsystem for Linux 2 (WSL2) provides a full Linux environment on Windows. Here's how to build Windows executables from WSL2:

### Prerequisites for WSL2
```bash
# Ensure Go is installed in WSL2
go version

# Install make if not available
sudo apt update && sudo apt install make

# Verify you can cross-compile to Windows
go env GOOS GOARCH
```

### Building Windows Executables from WSL2

#### Method 1: Direct Cross-Compilation
```bash
# Build Windows executable
GOOS=windows GOARCH=amd64 go build -o meraki-info.exe

# Optimized Windows build (smaller binary)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o meraki-info.exe

# Windows build with custom name
GOOS=windows GOARCH=amd64 go build -o meraki-info-windows.exe
```

#### Method 2: Using the Makefile
```bash
# Build Windows target using make
make build-windows           # Creates meraki-info.exe

# Build all platforms including Windows
make build-all              # Creates executables for all platforms

# Clean and build Windows
make clean && make build-windows
```

#### Method 3: Testing the Windows Executable
```bash
# Build Windows executable
make build-windows

# Run Windows executable using Wine (optional)
# First install Wine: sudo apt install wine
wine meraki-info.exe -help

# Or copy to Windows and test natively
cp meraki-info.exe /mnt/c/Users/YourUsername/Desktop/
```

### WSL2 Development Workflow

#### Complete Development Cycle
```bash
# 1. Clone repository in WSL2
git clone <repository-url>
cd meraki-info

# 2. Install dependencies
go mod tidy

# 3. Run tests (Linux environment)
make test

# 4. Build for Windows target
make build-windows

# 5. Build for all platforms
make build-all

# 6. Verify Windows executable exists
ls -la meraki-info.exe

# 7. Copy to Windows filesystem for testing
cp meraki-info.exe /mnt/c/temp/
```

#### Cross-Platform Testing
```bash
# Test Linux version in WSL2
./meraki-info -help

# Test Windows version (copy to Windows first)
cp meraki-info.exe /mnt/c/Users/$USER/Desktop/
# Then run from Windows: meraki-info.exe -help
```

### File System Integration

#### Accessing WSL2 Files from Windows
```bash
# WSL2 files are accessible from Windows at:
# \\wsl$\Ubuntu\home\username\meraki-info

# Copy built executable to Windows Desktop
cp meraki-info.exe /mnt/c/Users/$USER/Desktop/

# Copy to a shared Windows folder
cp meraki-info.exe /mnt/c/temp/meraki-tools/
```

#### Best Practices for WSL2 Development
```bash
# Work in Linux filesystem for better performance
cd ~/projects/meraki-info  # Not /mnt/c/...

# Use WSL2 for development, copy executables to Windows
make build-windows
cp meraki-info.exe /mnt/c/Users/$USER/Desktop/

# Version control works seamlessly
git add . && git commit -m "Added WSL2 build instructions"
```

### Troubleshooting WSL2 Builds

#### Common Issues and Solutions
```bash
# Issue: Go not found
sudo apt update && sudo apt install golang-go

# Issue: Permission denied when copying to Windows
# Ensure target directory exists and has proper permissions
mkdir -p /mnt/c/temp && cp meraki-info.exe /mnt/c/temp/

# Issue: Executable won't run on Windows
# Verify it's a Windows executable
file meraki-info.exe  # Should show "PE32+ executable"

# Issue: Large executable size
# Use build flags to reduce size
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o meraki-info.exe
```

### Performance Comparison

| Build Method | Location | Performance | File Access |
|--------------|----------|-------------|-------------|
| Native Windows | C:\ | Fast | Native |
| WSL2 (Linux FS) | ~/projects | Fast | WSL2 optimized |
| WSL2 (Windows FS) | /mnt/c | Slower | Cross-filesystem |
| PowerShell Scripts | C:\ | Fast | Native Windows |

**Recommendation**: Develop in WSL2 Linux filesystem (`~/projects`) for best performance, then copy Windows executables to Windows filesystem when needed.

## Error Handling

The application provides clear error messages for common issues:

- **Authentication failures**: Invalid API key or insufficient permissions
- **Network issues**: Connection timeouts or API unavailability
- **File system errors**: Permission issues or disk space problems
- **Invalid configuration**: Missing required parameters
- **Build errors**: Missing dependencies, unsupported platforms, or compilation issues

### Build-Specific Troubleshooting

**PowerShell Execution Policy:**
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**Missing Go Installation:**
```bash
go version  # Should show go1.24.2 or later
```

**ARM64 Build Issues:**
Requires Go 1.16+ for ARM64 support:
```bash
go version  # Ensure go1.16 or later for ARM64
```

## Rate Limiting

How the client behaves around Meraki API rate limits:
- A 429 (or 500/502/503/504, or a network error) is retried up to 3 times, waiting 1s, 1s and
  then 2s. The `Retry-After` header is not read.
- Each request has a 30-second timeout.
- Requests are made one at a time; `-all` makes several calls per network, so large
  organizations take a while.
- List endpoints are read as a single page; very large organizations may be truncated.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass (`.\make.ps1 test` or `make test`)
5. Build for all platforms (`.\make.ps1 build-all` or `make build-all`)
6. Submit a pull request

### Development Environment Setup

**Windows:**
```powershell
# Clone and setup
git clone <repository-url>
cd meraki-info
.\make.ps1 deps              # Install dependencies
.\make.ps1 test              # Run tests
.\make.ps1 build             # Build for current platform
```

**WSL2 (Windows Subsystem for Linux):**
```bash
# Clone and setup in WSL2
git clone <repository-url>
cd meraki-info
make deps                    # Install dependencies
make test                    # Run tests (Linux environment)
make build-windows           # Build Windows executable
make build-all               # Build all platforms
```

**Linux/macOS:**
```bash
# Clone and setup
git clone <repository-url>
cd meraki-info
make deps                    # Install dependencies
make test                    # Run tests
make build                   # Build for current platform
```

## License

[Add your license information here]

## Support

For issues and questions:
1. Check the error messages for troubleshooting hints
2. Enable debug logging for more detailed information (`-loglevel debug`)
3. Review build documentation in [BUILD_SCRIPTS.md](BUILD_SCRIPTS.md) and [MAKE_SUPPORT.md](MAKE_SUPPORT.md)
4. Consult the Meraki API documentation
5. Open an issue in the repository

### Documentation

- **[BUILD_SCRIPTS.md](BUILD_SCRIPTS.md)**: Comprehensive guide to all build scripts and cross-platform compilation
- **[MAKE_SUPPORT.md](MAKE_SUPPORT.md)**: PowerShell make system documentation with ARM64 support details
- **[REQUIREMENTS.md](REQUIREMENTS.md)**: Detailed project requirements and specifications
- **[CODE_REVIEW.md](CODE_REVIEW.md)**: Code review guidelines and best practices

### Quick Reference

**Windows Development:**
```powershell
.\make.ps1 help              # Show all available commands
.\build.ps1 -Help            # Simple build script help
```

**Cross-Platform:**
```bash
make help                    # Show all Makefile targets
```

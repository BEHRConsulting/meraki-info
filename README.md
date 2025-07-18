# Meraki Info

This project is a Golang application that collects Meraki network information.

## AI Generated (mostly) 
- This was an experiment using Github's Copilot in agent mode and Claude Sonnet 4 AI model.
- I kept tweeking the prompts to get the functions I needed.
- The prompts I used the generate this app are on the bottom of the file initial-prompts.md.

## Features

- **Cross-Platform Support**: Native builds for Windows, Linux, macOS (Intel & Apple Silicon)
- **ARM64 Compatibility**: Full support for Apple Silicon Macs and ARM64 Linux servers  
- **Enhanced Build System**: PowerShell scripts for Windows-native development experience
- **Secure Authentication**: Supports both API Key and OAuth2 authentication methods
- **Flexible Configuration**: Command line flags and environment variables
- **Multiple Output Formats**: Text, JSON, XML, and CSV
- **Structured Logging**: Configurable log levels (debug, info, error)
- **Error Handling**: Graceful error handling with clear messages
- **Rate Limiting Aware**: Efficient API usage to avoid hitting rate limits
- **Unit Tests**: Comprehensive test coverage
- **Multiple Build Options**: Traditional Makefile, PowerShell scripts, and batch files

## Installation

### Prerequisites
- Go 1.21 or later (Go 1.16+ required for ARM64 support)
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
| `-all` | - | Get info for all networks to separate timestamped files | No |

**Commands (positional arguments):**
- `access` - Show available organizations and networks
- `route-tables` - Output route tables
- `licenses` - Output license information  
- `down` - Output all devices that are down/offline

*Organization is not required when using `access` command.
*The `-all` and `-network` options cannot be used together.

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

#### Get info for all networks to separate files
```bash
# Get info for all networks in organization to separate timestamped files (text format)
./meraki-info -apikey your-api-key -org your-org-id -all route-tables

# Get info for all networks to JSON files
./meraki-info -apikey your-api-key -org your-org-id -all -format json route-tables

# Get info for all networks to CSV files  
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

### OAuth2 (For production applications)
The application supports OAuth2 authentication for production use cases. See the Meraki API documentation for OAuth2 setup instructions.

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

### Single Network Info
When backing up a single network:

**When `-output` is specified**: Uses the exact filename provided.
```bash
./meraki-info -apikey key -org 123 -network "City Core" -output "my-info.txt" route-tables
# Creates: my-info.txt
```

**When `-output` is NOT specified or set to "default"**: Auto-generates filename in format:
```
RouteTables-<OrganizationName>-<NetworkName>-<RFC3339 datetime>.<extension>
```

Examples:
- `RouteTables-City_of_Gardena-City_Core-2025-07-15T18-10-17-07-00.txt`
- `Licenses-YourOrganization-YourNetwork-2025-07-15T18-12-15-07-00.csv`
- `Down-MyCompany-MainOffice-2025-07-15T14-30-45-07-00.json`

### All Networks Info (`-all` option)
When using the `-all` option, each network gets its own file with the following naming convention:
```
<CommandType>-<OrganizationName>-<NetworkName>-<RFC3339 datetime>.<extension>
```

Examples:
- `RouteTables-City_of_Gardena-City_Core-2025-07-15T17-59-21-07-00.txt`
- `Licenses-YourOrganization-YourNetwork-2025-07-15T18-01-01-07-00.json`
- `Down-MyCompany-MainOffice-2025-07-15T14-30-45-07-00.csv`

**Note**: Special characters in organization and network names are replaced with underscores for filesystem compatibility.

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
│   │   └── client_test.go
│   └── output/                 # Output formatters
│       ├── writer.go
│       └── writer_test.go
├── build.ps1                   # Simple PowerShell build script
├── build-windows.ps1           # Advanced PowerShell build script
├── make.ps1                    # Full PowerShell make system
├── make.bat                    # Batch wrapper for Command Prompt
├── PowerShell-Make-Function.ps1 # Global PowerShell make function
├── Makefile                    # Traditional make system (cross-platform)
├── BUILD_SCRIPTS.md            # Comprehensive build documentation
├── MAKE_SUPPORT.md             # PowerShell make system documentation
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
go version  # Should show go1.21 or later
```

**ARM64 Build Issues:**
Requires Go 1.16+ for ARM64 support:
```bash
go version  # Ensure go1.16 or later for ARM64
```

## Rate Limiting

The application is designed to be respectful of Meraki API rate limits:
- Uses efficient API calls
- Implements proper timeouts
- Provides clear error messages if rate limits are exceeded

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

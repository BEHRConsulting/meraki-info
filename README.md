# Meraki Info

This project is a Golang application that collects Meraki network information.

## AI Generated (mostly) 
- This was an experiment using Github's Copilot in agent mode and Claude Sonnet 4 AI model.
- I kept tweeking the prompts to get the functions I needed.
- The prompts I used the generate this app are on the bottom of the file meraki-info-prompts.md.

## Features

- **Secure Authentication**: Supports both API Key and OAuth2 authentication methods
- **Flexible Configuration**: Command line flags and environment variables
- **Multiple Output Formats**: Text, JSON, XML, and CSV
- **Structured Logging**: Configurable log levels (debug, info, error)
- **Error Handling**: Graceful error handling with clear messages
- **Rate Limiting Aware**: Efficient API usage to avoid hitting rate limits
- **Unit Tests**: Comprehensive test coverage

## Installation

### Prerequisites
- Go 1.21 or later
- Meraki Dashboard API access

### Build from Source
```bash
git clone <repository-url>
cd meraki-info
go mod tidy
go build -o meraki-info
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
| `-all` | - | Backup all networks to separate timestamped files | No |

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

#### Backup specific network to JSON
```bash
./meraki-info -apikey your-api-key -org your-org-id -network net-id -output routes.json -format json route-tables
```

#### Backup all networks to separate files
```bash
# Backup all networks in organization to separate timestamped files (text format)
./meraki-info -apikey your-api-key -org your-org-id -all route-tables

# Backup all networks to JSON files
./meraki-info -apikey your-api-key -org your-org-id -all -format json route-tables

# Backup all networks to CSV files  
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

### Single Network Backup
When backing up a single network:

**When `-output` is specified**: Uses the exact filename provided.
```bash
./meraki-info -apikey key -org 123 -network "City Core" -output "my-backup.txt" route-tables
# Creates: my-backup.txt
```

**When `-output` is NOT specified or set to "default"**: Auto-generates filename in format:
```
RouteTables-<OrganizationName>-<NetworkName>-<RFC3339 datetime>.<extension>
```

Examples:
- `RouteTables-City_of_Gardena-City_Core-2025-07-15T18-10-17-07-00.txt`
- `Licenses-YourOrganization-YourNetwork-2025-07-15T18-12-15-07-00.csv`
- `Down-MyCompany-MainOffice-2025-07-15T14-30-45-07-00.json`

### All Networks Backup (`-all` option)
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

### Building
```bash
# Build for current platform
go build -o meraki-info

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o meraki-info-linux

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o meraki-info.exe
```

## Error Handling

The application provides clear error messages for common issues:

- **Authentication failures**: Invalid API key or insufficient permissions
- **Network issues**: Connection timeouts or API unavailability
- **File system errors**: Permission issues or disk space problems
- **Invalid configuration**: Missing required parameters

## Rate Limiting

The application is designed to be respectful of Meraki API rate limits:
- Uses efficient API calls
- Implements proper timeouts
- Provides clear error messages if rate limits are exceeded

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

[Add your license information here]

## Support

For issues and questions:
1. Check the error messages for troubleshooting hints
2. Enable debug logging for more detailed information
3. Consult the Meraki API documentation
4. Open an issue in the repository

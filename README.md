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

### Command Line Flags

| Flag | Environment Variable | Description | Required |
|------|---------------------|-------------|----------|
| `--apikey` | `MERAKI_APIKEY` | Meraki API key | Yes |
| `--org` | `MERAKI_ORG` | Meraki organization ID | Yes* |
| `--network` | `MERAKI_NET` | Specific network ID or name (optional) | No |
| `--output` | - | Output file path | No (default: routes.txt) |
| `--format` | - | Output format: text, json, xml, csv | No (default: text) |
| `--loglevel` | - | Log level: debug, info, error | No (default: error) |
| `--access` | - | Show available organizations and networks | No |
| `--all` | - | Backup all networks to separate timestamped files | No |

*Organization is not required when using `--access` flag.
*The `--all` and `--network` flags cannot be used together.

### Examples

#### Basic usage with API key
```bash
./meraki-info --apikey your-api-key --org your-org-id
```

#### Check available organizations and networks
```bash
# Show all accessible organizations and networks
./meraki-info --access --apikey your-api-key

# Show networks for a specific organization only
./meraki-info --access --apikey your-api-key --org "766096"
./meraki-info --access --apikey your-api-key --org "City of Gardena"
```

#### Using environment variables
```bash
export MERAKI_APIKEY="your-api-key"
export MERAKI_ORG="your-org-id"
./meraki-info
```

#### Backup specific network to JSON
```bash
./meraki-info --apikey your-api-key --org your-org-id --network net-id --output routes.json --format json
```

#### Backup all networks to separate files
```bash
# Backup all networks in organization to separate timestamped files (text format)
./meraki-info --apikey your-api-key --org your-org-id --all

# Backup all networks to JSON files
./meraki-info --apikey your-api-key --org your-org-id --all --format json

# Backup all networks to CSV files  
./meraki-info --apikey your-api-key --org your-org-id --all --format csv
```

#### Network identification
```bash
# Use network ID
./meraki-info --apikey your-api-key --org your-org-id --network "L_660903245316649608"

# Use network name (must be unique within organization)
./meraki-info --apikey your-api-key --org your-org-id --network "City Core"
```

#### Custom output filename
```bash
# Specify custom output filename
./meraki-info --apikey your-api-key --org your-org-id --network "City Core" --output my-routes.txt

# Use default auto-generated filename: <org>-<network>-RouteTables-yyyy-mm-dd-hh-mm-ss
./meraki-info --apikey your-api-key --org your-org-id --network "City Core"

# Output to stdout (useful for piping)
./meraki-info --apikey your-api-key --org your-org-id --network "City Core" --output "-"

# Pipe JSON output to jq for processing
./meraki-info --apikey your-api-key --org your-org-id --network "City Core" --output "-" --format json | jq '.[] | .name'

# Save CSV output with custom processing
./meraki-info --apikey your-api-key --org your-org-id --network "City Core" --output "-" --format csv > processed-routes.csv
```

#### Enable debug logging
```bash
./meraki-info --apikey your-api-key --org your-org-id --loglevel debug
```

## Authentication

### API Key (Recommended for scripts)
1. Log in to the Meraki Dashboard
2. Navigate to Organization > Settings > Dashboard API access
3. Generate an API key
4. Use the key with the `--apikey` flag or `MERAKI_APIKEY` environment variable

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

**When `--output` is specified**: Uses the exact filename provided.
```bash
./meraki-info --apikey key --org 123 --network "City Core" --output "my-backup.txt"
# Creates: my-backup.txt
```

**When `--output` is NOT specified**: Auto-generates filename in format:
```
<OrganizationName>-<NetworkName>-RouteTables-<YYYY-MM-DD-HH-MM-SS>.<extension>
```

Examples:
- `City_of_Gardena-City_Core-RouteTables-2025-07-15-18-10-17.txt`
- `City_of_Gardena-CoG_CityHall_-_z3-RouteTables-2025-07-15-18-12-15.csv`
- `MyCompany-MainOffice-RouteTables-2025-07-15-14-30-45.json`

### All Networks Backup (`--all` flag)
When using the `--all` flag, each network gets its own file with the following naming convention:
```
<OrganizationName>-RouteTables-<YYYY-MM-DD-HH-MM-SS>-<NetworkName>.<extension>
```

Examples:
- `City_of_Gardena-RouteTables-2025-07-15-17-59-21-City_Core.txt`
- `City_of_Gardena-RouteTables-2025-07-15-18-01-01-CoG_CityHall_-_z3.json`
- `MyCompany-RouteTables-2025-07-15-14-30-45-MainOffice.csv`

**Note**: Special characters in organization and network names are replaced with underscores for filesystem compatibility.

### Stdout Output
When `--output "-"` is specified, the route table is sent to stdout instead of a file. This enables:

**Piping to other tools:**
```bash
# Extract route names using jq
./meraki-info --org 123 --network "Main" --output "-" --format json | jq '.[] | .name'

# Count total routes
./meraki-info --org 123 --network "Main" --output "-" --format json | jq '. | length'

# Filter enabled routes only
./meraki-info --org 123 --network "Main" --output "-" --format json | jq '.[] | select(.enabled == true)'
```

**Processing CSV data:**
```bash
# Save to file with custom name
./meraki-info --org 123 --network "Main" --output "-" --format csv > custom-name.csv

# Process with awk
./meraki-info --org 123 --network "Main" --output "-" --format csv | awk -F',' '{print $2, $3}'
```

**Integration with scripts:**
```bash
#!/bin/bash
ROUTES=$(./meraki-info --org 123 --network "Main" --output "-" --format json)
echo "$ROUTES" | jq '.[] | select(.subnet | contains("192.168"))'
```

## Configuration

### Environment Variables
- `MERAKI_APIKEY`: Your Meraki API key
- `MERAKI_ORG`: Organization ID
- `MERAKI_NET`: Network ID (optional)

### Configuration Priority
1. Command line flags (highest priority)
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

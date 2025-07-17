# Meraki Info CLI Application Requirements

## Project Overview
This project is a Golang application that collects Meraki network information through the Meraki Dashboard API.

## Core Requirements

### Authentication
- **Primary Authentication**: Use production authentication methods and libraries for Meraki, such as OAuth2, to ensure secure access to the Meraki account
- **API Key Support**: Support API key authentication as an alternative method
- **Environment Variables**: Support authentication configuration through environment variables

### Commands
The application must support the following commands (displayed in alphabetical order in usage):

1. **`access`**: Print a nice text output listing the organizations and networks available for the API key
   - Allow filtering by `--org` parameter
   
2. **`alerting`**: Output all alerting devices across the specified scope

3. **`down`**: Output all devices that are down/offline

4. **`licenses`**: Output license information for the specified scope

5. **`route-tables`**: Output route tables from security appliances, switches, and switch stacks

### Command Line Options

#### Required Parameters
- **Command**: One of the above commands is required; if not provided, display usage

#### Optional Parameters
- **`--org`**: Specify organization by name or ID (case insensitive)
  - Environment variable: `MERAKI_ORG`
  
- **`--network`**: Specify network by name or ID (case insensitive)
  - Environment variable: `MERAKI_NET`
  - If not provided, defaults to `--all` behavior
  
- **`--apikey`**: API key for authentication
  - Environment variable: `MERAKI_APIKEY`
  - Usage display: Do not show default; if set, show "env MERAKI_APIKEY is set"
  
- **`--output`**: Output file name
  - Default: stdout
  - Special value: "-" for stdout
  
- **`--format`**: Output format (text, xml, json, csv)
  - Default: text
  
- **`--all`**: Generate consolidated output
  - If `--org` not specified: process all organizations
  - If `--network` not specified: process all networks
  - Consolidated output must include: org name, org ID, network name, network ID
  
- **`--loglevel`**: Set logging level (debug, info, error)
  - Default: error

### Technical Requirements

#### Error Handling
- **No Panics**: Application must handle all errors gracefully
- **Clear Error Messages**: Provide meaningful error messages for:
  - Authentication failures
  - Network issues
  - File system errors
- **Error Output**: Send all error output to stderr

#### Performance & Efficiency
- **API Rate Limiting**: Minimize API calls to avoid hitting Meraki rate limits
- **Efficient Processing**: Optimize API call patterns and data processing

#### Code Quality
- **Modular Structure**: Well-structured, modular code for maintainability and extensibility
- **Documentation**: Include comprehensive comments and documentation
- **Unit Tests**: Create comprehensive unit tests for all components
- **Go Best Practices**: Follow Go idioms and best practices

#### Output Requirements
- **Consolidated Format**: When using `--all`, provide consolidated output with organizational context
- **Multiple Formats**: Support text, XML, JSON, and CSV output formats
- **Flexible Output**: Support both file output and stdout

### Usage Display Requirements
- Commands must be listed in alphabetical order
- Sanitize all examples in documentation
- Clear parameter descriptions and defaults

## Additional Notes
- The application should be production-ready with proper error handling
- Consider implementing retry logic for API calls
- Ensure secure handling of API keys and authentication tokens
# Meraki-Info Application Requirements

## Overview
Application name: **meraki-info**

**Description:** This project is a Golang application that collects Meraki network information.

## Authentication
- The app should authenticate with Meraki cloud
- Use production authentication methods and libraries for Meraki, such as OAuth2, to ensure secure access to the Meraki account

## Command Line Parameters

### Required Parameters
One of the following parameters is required:
- `--licenses` - Output license information
- `--route-tables` - Output route tables  
- `--access` - Show available organizations and networks

### Authentication
- `--apikey` - Meraki API key (can also be set with env variable `MERAKI_APIKEY`)

### Organization & Network Selection
- `--org` - Organization ID or name (can also be set with env variable `MERAKI_ORG`)
  - Allow organization to be specified by name or id
  - This is not case sensitive
- `--network` - Network ID or name (can also be set with env variable `MERAKI_NET`)
  - Allow network to be specified by name or id  
  - This is not case sensitive

### Output Options
- `--output` - Output file path
  - If `--output` is "-" or not provided → send to stdout
  - If `--output` is "default" → use default filenames
  - Otherwise → use specified filename
- `--format` - Output format: text, xml, json, csv (default: text)

### Special Modes
- `--access` - Print a nice text output listing the organizations and networks available for the API key
  - Allow filtering by `--org` parameter
- `--all` - Generate files for all networks in the specified organization
  - If `--org` is not specified, process all organizations

### Logging
- `--loglevel` - Set logging level: debug, info, error (default: error)

## Default Filenames
- **Route tables:** `RouteTables-<org>-<network>-<RFC3339 date time>.txt`
- **Licenses:** `Licenses-<org>-<network>-<RFC3339 date time>.txt`

## Development Requirements

### Code Quality
- Create unit tests
- There should be no panics
- Handle errors gracefully with clear messages for:
  - Authentication failures
  - Network issues  
  - File system errors
- Code should be well-structured and modular for maintainability
- Include comments and documentation

### Performance
- Be efficient with API calls to avoid hitting rate limits
- Minimize the number of requests made

### Architecture
- Well-structured and modular design
- Easy to maintain and extend in the future
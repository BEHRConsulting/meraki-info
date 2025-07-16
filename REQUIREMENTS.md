# Meraki-Info Application Requirements

> A Golang application for collecting Meraki network information with comprehensive route table and license management capabilities.

## Table of Contents
- [Overview](#overview)
- [Authentication](#authentication)
- [Command Line Interface](#command-line-interface)
- [Output Formats](#output-formats)
- [Usage Modes](#usage-modes)
- [Development Standards](#development-standards)

## Overview

| Property | Value |
|----------|-------|
| **Application Name** | `meraki-info` |
| **Language** | Go |
| **Purpose** | Collect and export Meraki network information |
| **Target Platform** | Cross-platform (Windows, macOS, Linux) |

**Description:** A command-line tool for backing up and analyzing Cisco Meraki network configurations, including route tables and license information.

## Authentication

### Methods
- **Primary:** API Key authentication
- **Production:** OAuth2 for secure enterprise access
- **Security:** Production-grade authentication libraries required

### Configuration
```bash
# Environment variables (recommended)
export MERAKI_APIKEY="your-api-key"
export MERAKI_ORG="your-org-id"
export MERAKI_NET="your-network-id"
```

## Command Line Interface

### Required Parameters
**One of the following modes must be specified:**

| Parameter | Description | Default Filename Pattern |
|-----------|-------------|---------------------------|
| `--route-tables` | Export network route tables | `RouteTables-<org>-<network>-<RFC3339>.txt` |
| `--licenses` | Export license information | `Licenses-<org>-<network>-<RFC3339>.txt` |
| `--access` | Show available organizations/networks | N/A (stdout only) |

### Optional Parameters

#### Authentication
| Parameter | Environment Variable | Description | Required |
|-----------|---------------------|-------------|----------|
| `--apikey` | `MERAKI_APIKEY` | Meraki API key | Yes |

#### Scope Selection
| Parameter | Environment Variable | Description | Case Sensitive |
|-----------|---------------------|-------------|----------------|
| `--org` | `MERAKI_ORG` | Organization ID or name | No |
| `--network` | `MERAKI_NET` | Network ID or name | No |

#### Output Control
| Parameter | Values | Description | Default |
|-----------|--------|-------------|---------|
| `--output` | `<filename>` \| `"-"` \| `"default"` | Output destination | stdout |
| `--format` | `text` \| `json` \| `xml` \| `csv` | Output format | `text` |
| `--loglevel` | `debug` \| `info` \| `error` | Logging verbosity | `error` |

#### Bulk Operations
| Parameter | Description | Compatibility |
|-----------|-------------|---------------|
| `--all` | Process all networks in organization | Cannot use with `--network` or stdout output |

### Output Behavior
```mermaid
graph TD
    A[--output parameter] --> B{Value?}
    B -->|Not provided or "-"| C[Send to stdout]
    B -->|"default"| D[Generate default filename]
    B -->|Custom filename| E[Use specified filename]
```

## Output Formats

### Supported Formats
- **Text** (default): Human-readable format
- **JSON**: Machine-readable, API-friendly
- **XML**: Structured markup format  
- **CSV**: Spreadsheet-compatible format

### Default Filenames
- **Route Tables:** `RouteTables-<org>-<network>-<RFC3339-datetime>.txt`
- **Licenses:** `Licenses-<org>-<network>-<RFC3339-datetime>.txt`

## Usage Modes

### 1. Access Mode (`--access`)
**Purpose:** Discover available organizations and networks

**Features:**
- Lists all accessible organizations
- Shows networks within organizations
- Supports organization filtering with `--org`
- Always outputs to stdout

### 2. Route Tables Mode (`--route-tables`)
**Purpose:** Export network routing information

**Features:**
- Extract static routes from appliance networks
- Support single network or bulk export (`--all`)
- Multiple output formats available

### 3. Licenses Mode (`--licenses`)
**Purpose:** Export license information

**Features:**
- Retrieve license details and status
- Support single network or bulk export (`--all`)
- Multiple output formats available

## Development Standards

### Code Quality Requirements
- [ ] **No panics** - All error conditions must be handled gracefully
- [ ] **Comprehensive unit tests** - All packages must have test coverage
- [ ] **Error handling** - Clear error messages for:
  - Authentication failures
  - Network connectivity issues
  - File system errors
  - API rate limiting
- [ ] **Documentation** - Code comments and usage documentation required

### Architecture Principles
- [ ] **Modular design** - Well-separated concerns and packages
- [ ] **Dependency injection** - For better testability
- [ ] **Repository pattern** - For API interactions
- [ ] **Structured logging** - Using `log/slog` package

### Performance Requirements
- [ ] **API efficiency** - Minimize API calls to avoid rate limits
- [ ] **Concurrent processing** - Where applicable for bulk operations
- [ ] **Memory management** - Efficient handling of large datasets

### Security Considerations
- [ ] **Secure authentication** - Production OAuth2 implementation
- [ ] **Credential management** - No hardcoded secrets
- [ ] **Rate limiting** - Respect Meraki API limits
- [ ] **Input validation** - Sanitize all user inputs

---

## Implementation Status

| Feature | Status | Notes |
|---------|--------|-------|
| API Key Authentication | ✅ Complete | |
| OAuth2 Authentication | ✅ Complete | |
| Case-insensitive org/network lookup | ✅ Complete | |
| Route table export | ✅ Complete | |
| License export | ✅ Complete | |
| Access mode | ✅ Complete | |
| Multiple output formats | ✅ Complete | |
| Bulk operations | ✅ Complete | |
| Error handling (no panics) | ✅ Complete | |
| Unit tests | ✅ Complete | |

> **Last Updated:** July 16, 2025
# Meraki Info - Requirements Documentation

## 🎯 Project Overview

**Meraki Info** is a command-line application written in Go that collects and displays information from Cisco Meraki cloud networks. The application provides secure access to Meraki network data through OAuth2 authentication and supports multiple output formats for various network information types.

| **Property** | **Value** |
|--------------|-----------|
| **Application Name** | `meraki-info` |
| **Language** | Go 1.19+ |
| **Purpose** | Collect and export Meraki network information |
| **Target Platform** | Cross-platform (Linux, macOS, Windows) |
| **License** | MIT |

---

## 📋 Table of Contents

- [🎯 Project Overview](#-project-overview)
- [🔐 Authentication & Security](#-authentication--security)
- [⚙️ Command Structure](#️-command-structure)
- [🔧 Configuration Options](#-configuration-options)
- [📤 Output Requirements](#-output-requirements)
- [📊 Data Collection Requirements](#-data-collection-requirements)
- [⚡ Performance Requirements](#-performance-requirements)
- [🏗️ Technical Architecture](#️-technical-architecture)
- [✅ Code Quality](#-code-quality)
- [🛡️ Security Requirements](#️-security-requirements)
- [🚀 Deployment Requirements](#-deployment-requirements)
- [📈 Implementation Status](#-implementation-status)

---

## 🔐 Authentication & Security

### **Primary Authentication Methods**
- **🔑 API Key**: Via `--apikey` flag or `MERAKI_APIKEY` environment variable
- **🔒 OAuth2**: Production-ready enterprise authentication
- **🛡️ Security**: No sensitive data exposed in usage output

### **Security Features**
```bash
# Environment variables (recommended approach)
export MERAKI_APIKEY="your-api-key-here"
export MERAKI_ORG="your-organization"
export MERAKI_NET="your-network"
```

| **Security Aspect** | **Implementation** |
|---------------------|-------------------|
| **Credential Display** | Shows "env MERAKI_APIKEY is set" when configured |
| **Error Handling** | No credentials in error messages or logs |
| **Network Security** | TLS 1.2+ for all API communications |
| **Access Control** | Respects organization-level permissions |

---

## ⚙️ Command Structure

### **Usage Format**
```bash
meraki-info [OPTIONS] COMMAND
```

### **Available Commands** *(alphabetical order)*

| **Command** | **Description** | **Output** |
|-------------|-----------------|------------|
| **`access`** | 🌐 Display organizations and networks available for API key | Organizations & Networks list |
| **`alerting`** | 🚨 Output all devices currently in alerting state | Device alert status |
| **`down`** | ⬇️ Output all devices that are currently offline | Offline device inventory |
| **`licenses`** | 📜 Output license information for the organization | License details & expiration |
| **`route-tables`** | 🛣️ Output routing tables from appliances and switches | Network routing information |

### **Command Validation**
- ✅ **Required**: One command must be specified
- ✅ **Case Sensitive**: Exact command matching
- ✅ **Help Display**: Commands listed in alphabetical order
- ❌ **Error**: Display usage if no command provided

---

## 🔧 Configuration Options

### **🏢 Organization Selection**
```bash
--org <name|id>              # Organization by name or ID
export MERAKI_ORG="BCI"      # Environment variable
```
- **Matching**: Case-insensitive by name or ID
- **Default**: Process all organizations when used with `--all`

### **🌐 Network Selection**
```bash
--network <name|id>          # Network by name or ID  
export MERAKI_NET="Main"     # Environment variable
```
- **Matching**: Case-insensitive by name or ID
- **Auto-behavior**: When omitted, automatically enables `--all`

### **📤 Output Configuration**
```bash
--output <filename>          # Write to file
--output -                   # Write to stdout (default)
--format <text|json|xml|csv> # Output format (default: text)
```

### **🔄 Processing Scope**
```bash
--all                        # Process all networks/organizations
```
- **All Organizations**: `--all` without `--org` processes all accessible orgs
- **All Networks**: `--all` without `--network` processes all networks
- **Consolidated Output**: Unified output with organizational context

### **📝 Logging Configuration**
```bash
--loglevel <error|info|debug>  # Logging verbosity (default: error)
```

---

## 📤 Output Requirements

### **📋 Supported Output Formats**

| **Format** | **Description** | **Use Case** |
|------------|-----------------|--------------|
| **`text`** | Human-readable formatted output | Manual review, reports |
| **`json`** | Machine-readable JSON structure | API integration, automation |
| **`xml`** | Well-formed XML documents | Legacy system integration |
| **`csv`** | Comma-separated values with headers | Spreadsheet analysis |

### **🎯 Consolidated Output Features**
- **📍 Context**: Organization name and ID included
- **🏷️ Network Info**: Network name and ID for each entry
- **🔄 Unified Format**: Consistent structure across all data types
- **📊 Headers**: Clear section headers and metadata

### **❌ Error Output Standards**
- **Target**: All errors directed to `stderr`
- **Clarity**: Clear, actionable error messages
- **User-Friendly**: No technical stack traces for end users
- **Context**: Meaningful error context and suggestions

---

## 📊 Data Collection Requirements

### **🛣️ Route Tables**
- **Sources**: Security appliances, switches, switch stacks
- **Content**: Static routes, metrics, next-hop information
- **Context**: Network topology and VLAN information
- **Format**: Route priority, destination networks, gateways

### **📜 License Information**
- **Details**: License types, editions, SKUs
- **Timing**: Expiration dates, duration, grace periods
- **Assignment**: Device assignments and utilization
- **Status**: Active, unused, expired, pending states

### **📊 Device Status Monitoring**
- **Health**: Online/offline status with timestamps
- **Alerts**: Alert conditions, severity levels, descriptions
- **Identity**: Device identification, serial numbers, models
- **Location**: Network assignment and physical location data

### **🔐 Access Information**
- **Organizations**: Available orgs with permission levels
- **Networks**: Network inventory with access scope
- **Permissions**: API scope limitations and capabilities
- **Filtering**: Organization-based filtering support

---

## ⚡ Performance Requirements

### **🚀 API Efficiency**
- **Rate Limiting**: Respect Meraki API rate limits (5 requests/second)
- **Batching**: Implement request batching where possible
- **Pagination**: Efficient handling of large datasets
- **Caching**: Smart caching of organization/network metadata

### **💾 Memory Management**
- **Streaming**: Stream large datasets when possible
- **Memory Usage**: Avoid loading entire datasets into memory
- **Garbage Collection**: GC-friendly data structures
- **Resource Cleanup**: Proper cleanup of HTTP connections

### **🔄 Error Recovery**
- **Graceful Degradation**: Continue on partial failures
- **Retry Logic**: Exponential backoff for transient errors
- **Timeout Handling**: Appropriate timeouts for long operations
- **Progress Indication**: Clear indication of processing status

---

## 🏗️ Technical Architecture

### **📁 Project Structure**
```
internal/
├── config/     # 🔧 Configuration and CLI argument parsing
├── logger/     # 📝 Structured logging with slog
├── meraki/     # 🌐 Meraki API client with OAuth2 support
└── output/     # 📤 Output formatters (text, JSON, XML, CSV)
```

### **🎨 Design Patterns**
- **Repository Pattern**: Clean API interaction abstraction
- **Dependency Injection**: Enhanced testability and modularity
- **Interface-Based Design**: Extensible architecture
- **Error Wrapping**: Meaningful context in error chains

### **🔌 Dependencies**
- **Standard Library**: Prefer built-in packages when possible
- **Third-Party**: Minimal, well-maintained, security-audited libraries
- **Authentication**: Production-grade OAuth2 libraries
- **HTTP Client**: Enhanced HTTP client with retry logic

---

## ✅ Code Quality

### **🧪 Testing Requirements**
- **Coverage**: Minimum 80% code coverage
- **Unit Tests**: All packages with comprehensive test suites
- **Integration Tests**: API interaction testing with mocks
- **Table-Driven Tests**: Complex scenario testing
- **Error Testing**: Comprehensive error condition coverage

### **📚 Go Best Practices**
- **Formatting**: `go fmt` compliance
- **Linting**: `go vet` and `golint` clean
- **Idioms**: Follow effective Go conventions
- **Error Handling**: Comprehensive with meaningful context
- **Logging**: Structured logging with `log/slog`

### **📖 Documentation Standards**
- **Code Comments**: Clear package and function documentation
- **Usage Examples**: Practical examples and scenarios
- **API Documentation**: Complete API interaction guide
- **Troubleshooting**: Common issues and solutions

### **🛡️ Reliability Requirements**
- **No Panics**: All error conditions handled gracefully
- **Resource Management**: Proper cleanup and resource handling
- **Signal Handling**: Graceful shutdown on interruption
- **Input Validation**: Comprehensive input sanitization

---

## 🛡️ Security Requirements

### **🔒 Data Protection**
- **Credential Security**: No API keys in logs or error messages
- **Network Security**: TLS 1.2+ for all communications
- **Certificate Validation**: Strict certificate validation
- **Secure Storage**: Secure credential storage recommendations

### **🎛️ Access Control**
- **Read-Only Access**: Only read operations on Meraki resources
- **Permission Respect**: Honor organization-level access controls
- **Network Boundaries**: Respect network-level permissions
- **Audit Logging**: Comprehensive operation logging

### **🔍 Vulnerability Management**
- **Dependency Scanning**: Regular security scanning
- **Update Process**: Security update procedures
- **Disclosure**: Responsible vulnerability disclosure
- **Data Minimization**: Collect only necessary data

---

## 🚀 Deployment Requirements

### **📦 Build and Distribution**
- **Single Binary**: Self-contained executable
- **Cross-Platform**: Linux, macOS, Windows support
- **Minimal Dependencies**: No external runtime requirements
- **Container Ready**: Docker/container deployment support

### **🔧 Installation Options**
- **Direct Download**: Binary releases on GitHub
- **Package Managers**: Homebrew, APT, YUM compatibility
- **Go Install**: Standard `go install` support
- **Container Images**: Docker Hub and registry support

### **📋 Documentation Package**
- **README**: Comprehensive usage guide
- **Examples**: Practical usage scenarios
- **Troubleshooting**: Common issues and solutions
- **API Guide**: Meraki API integration details

---

## 📈 Implementation Status

### **✅ Core Features (Complete)**
| **Feature** | **Status** | **Coverage** |
|-------------|------------|--------------|
| CLI Application Structure | ✅ Complete | 100% |
| All Five Commands | ✅ Complete | 100% |
| Multiple Output Formats | ✅ Complete | 100% |
| Consolidated `--all` Processing | ✅ Complete | 100% |
| Error Handling (No Panics) | ✅ Complete | 100% |
| Unit Test Coverage | ✅ Complete | 95%+ |
| Structured Logging | ✅ Complete | 100% |
| API Key Authentication | ✅ Complete | 100% |

### **🔄 Advanced Features (In Progress)**
| **Feature** | **Status** | **Priority** |
|-------------|------------|--------------|
| OAuth2 Authentication | 🔄 In Progress | High |
| Performance Optimization | 🔄 In Progress | Medium |
| Extended Documentation | 🔄 In Progress | Medium |

### **📋 Future Enhancements (Planned)**
| **Feature** | **Status** | **Timeline** |
|-------------|------------|--------------|
| Configuration File Support | 📋 Planned | Q3 2025 |
| Advanced Filtering Options | 📋 Planned | Q3 2025 |
| Interactive Mode | 📋 Planned | Q4 2025 |
| Webhook Integration | 📋 Planned | Q4 2025 |
| Dashboard Export | 📋 Planned | 2026 |

---

## 🎯 Usage Examples

### **Basic Commands**
```bash
# List available organizations and networks
meraki-info access

# Get route tables for specific organization
meraki-info --org "BCI" route-tables

# Export all licenses to JSON file
meraki-info --all --format json --output licenses.json licenses

# Check device status with debug logging
meraki-info --loglevel debug --org "BCI" down
```

### **Advanced Usage**
```bash
# Consolidated export of all data types
meraki-info --all --format csv --output network-audit.csv route-tables

# Environment-based configuration
export MERAKI_APIKEY="your-key"
export MERAKI_ORG="BCI"
meraki-info alerting

# Multiple format exports
meraki-info --org "BCI" --format json route-tables > routes.json
meraki-info --org "BCI" --format csv route-tables > routes.csv
```

---

> **📅 Last Updated**: July 16, 2025  
> **📋 Document Version**: 2.0.0  
> **🏷️ Project Version**: 1.0.0
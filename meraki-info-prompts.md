# 🌐 Meraki Info - Project Specification

## 📋 Project Overview

**Meraki Info** is a comprehensive Golang command-line application designed to collect and analyze Cisco Meraki network information with production-grade security and reliability.

---

## 🎯 Core Functionality

### **📱 Application Commands** *(alphabetical order)*

| **Command** | **Description** | **Data Source** |
|-------------|-----------------|-----------------|
| `access` | Display organizations and networks available for API key | Dashboard API |
| `alerting` | Output all devices currently in alerting state | Device Status API |
| `down` | Output all devices that are currently offline | Device Status API |
| `licenses` | Output license information for organization | License API |
| `route-tables` | Output routing tables from appliances, switches, and stacks | Network API |

### **⚙️ Command Requirements**
- **Mandatory**: Exactly one command must be specified
- **Validation**: Display usage if no command provided
- **Ordering**: Commands listed alphabetically in help/usage

---

## 🔐 Authentication & Security

### **🔑 Authentication Methods**
- **Primary**: API Key authentication via `--apikey` flag or `MERAKI_APIKEY` environment variable
- **Enterprise**: OAuth2 production authentication for secure enterprise access
- **Security**: No sensitive data exposed in usage output

### **🛡️ Security Requirements**
- **Credential Protection**: API keys never displayed in usage (show "env MERAKI_APIKEY is set")
- **Error Handling**: No credentials in error messages or logs
- **Transport Security**: TLS encryption for all API communications
- **Access Control**: Honor organization and network-level permissions

---

## 🔧 Configuration Options

### **🏢 Organization Selection**
```bash
--org <name|id>              # Organization by name or ID
export MERAKI_ORG="BCI"      # Environment variable
```
- **Flexibility**: Case-insensitive matching by name or ID
- **Scope**: Process all organizations when used with `--all`

### **🌐 Network Selection**
```bash
--network <name|id>          # Network by name or ID
export MERAKI_NET="Main"     # Environment variable
```
- **Auto-behavior**: When omitted, automatically enables `--all` mode
- **Flexibility**: Case-insensitive matching by name or ID

### **🔑 API Key Configuration**
```bash
--apikey <key>               # API key for authentication
export MERAKI_APIKEY="key"   # Environment variable (recommended)
```
- **Security**: No default displayed in usage
- **Indication**: Show "env MERAKI_APIKEY is set" when configured

### **📤 Output Configuration**
```bash
--output <filename>          # Write to file
--output -                   # Write to stdout (default)
--format <text|json|xml|csv> # Output format (default: text)
```

### **🔄 Processing Scope**
```bash
--all                        # Enable consolidated processing
```
- **Organizations**: Process all accessible organizations (without `--org`)
- **Networks**: Process all networks (without `--network`)
- **Context**: Include org name, org ID, network name, and network ID

### **📝 Logging Configuration**
```bash
--loglevel <error|info|debug>  # Logging verbosity (default: error)
```

---

## 📊 Data Collection Specifications

### **🛣️ Route Tables**
- **Sources**: Security appliances (MX), switches, switch stacks
- **Content**: Static routes, metrics, next-hop information
- **Context**: Network topology and routing configuration

### **📜 License Information**
- **Details**: License types, editions, expiration dates
- **Assignment**: Device assignments and utilization tracking
- **Status**: Active, unused, expired, pending states

### **📊 Device Status**
- **Health**: Online/offline status with timestamps
- **Alerts**: Alert conditions, severity levels, descriptions
- **Identity**: Device serial numbers, models, MAC addresses

### **🔐 Access Information**
- **Organizations**: Available organizations with permissions
- **Networks**: Network inventory with access scope
- **Filtering**: Support `--org` parameter for organization filtering

---

## 📤 Output Requirements

### **📋 Supported Formats**
- **text**: Human-readable formatted output (default)
- **json**: Machine-readable JSON structure
- **xml**: Well-formed XML documents
- **csv**: Comma-separated values with headers

### **🎯 Consolidated Output Features**
- **Rich Context**: Organization name and ID for all records
- **Network Info**: Network name and ID for device records
- **Unified Structure**: Consistent format across all data types
- **Complete Traceability**: Full organizational context

### **❌ Error Handling**
- **Target**: All errors directed to `stderr`
- **Clarity**: Clear, actionable error messages
- **User-Friendly**: No technical stack traces for end users
- **Graceful**: No panics under any circumstances

---

## ⚡ Performance & Reliability

### **🚀 API Efficiency**
- **Rate Limiting**: Respect Meraki API limits (5 requests/second)
- **Optimization**: Minimize API calls to avoid rate limits
- **Batching**: Implement request batching where possible
- **Caching**: Smart caching of metadata

### **🔄 Error Recovery**
- **Graceful Degradation**: Continue on partial failures
- **Retry Logic**: Exponential backoff for transient errors
- **Timeout Handling**: Appropriate timeouts for operations
- **Clear Messages**: Meaningful error context and suggestions

### **💾 Resource Management**
- **Memory**: Efficient memory usage for large datasets
- **Connections**: Proper cleanup of HTTP connections
- **File System**: Handle file system errors gracefully
- **Network**: Resilient network error handling

---

## 🏗️ Code Quality Requirements

### **📁 Architecture**
- **Modular Design**: Well-structured, maintainable codebase
- **Clean Code**: Easy to understand and extend
- **Documentation**: Comprehensive comments and documentation
- **Best Practices**: Follow Go idioms and conventions

### **🧪 Testing**
- **Unit Tests**: Comprehensive test coverage for all packages
- **Test-Driven**: Write tests for all functionality
- **Coverage**: Achieve high test coverage metrics
- **Error Testing**: Test all error scenarios

### **📚 Documentation**
- **Code Comments**: Clear function and package documentation
- **Usage Examples**: Practical examples for all features
- **API Guide**: Complete API interaction documentation
- **Troubleshooting**: Common issues and solutions

---

## 🎯 Implementation Priorities

### **✅ Core Features**
1. **CLI Framework**: Command parsing and validation
2. **Authentication**: API key and OAuth2 support
3. **All Commands**: Implement all five commands
4. **Output Formats**: Support for text, JSON, XML, CSV
5. **Error Handling**: Graceful error handling without panics

### **🔄 Advanced Features**
1. **Consolidated Mode**: `--all` processing with full context
2. **Performance**: Rate limiting and retry logic
3. **Testing**: Comprehensive unit test suite
4. **Documentation**: Complete usage and API documentation

### **📋 Quality Assurance**
1. **Code Review**: Thorough code review process
2. **Testing**: All functionality tested
3. **Performance**: Efficient API usage
4. **Security**: Secure credential handling
5. **Documentation**: Clear, comprehensive documentation

---

## 🚀 Development Tasks

### **📝 Immediate Actions**
- [ ] Clean up and organize project documentation
- [ ] Regenerate professional `REQUIREMENTS.md`
- [ ] Perform comprehensive code review
- [ ] Document findings in `CODE_REVIEW.md`
- [ ] Sanitize all code examples
- [ ] Ensure alphabetical command ordering in usage

### **🔍 Quality Checks**
- [ ] Verify no panics in any code path
- [ ] Ensure all errors go to stderr
- [ ] Validate API key security measures
- [ ] Confirm rate limiting implementation
- [ ] Test all output formats
- [ ] Verify consolidated output includes all context

---

**📅 Last Updated**: July 17, 2025  
**🏷️ Document Version**: 1.0.0  
**👥 Maintainer**: BEHRConsulting
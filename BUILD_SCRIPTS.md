# Build Scripts for Windows

This directory contains Windows-specific build scripts for the Meraki Info application, providing native Windows development experience alongside the traditional Makefile approach.

## Available Scripts

### 1. PowerShell Script: `build.ps1` (Recommended for Simple Builds)

**Usage:**
```powershell
# Build for Windows (default)
.\build.ps1

# Build for all platforms
.\build.ps1 -All

# Clean and test before building
.\build.ps1 -Clean -Test

# Build for specific platform
.\build.ps1 -Target linux
.\build.ps1 -Target mac

# Show help
.\build.ps1 -Help
```

**Features:**
- Clean, simple output with progress indicators
- Cross-platform building (Windows, Linux, macOS)
- Automatic testing of Windows executable
- File size reporting
- Error handling with proper exit codes

### 2. Advanced PowerShell Script: `build-windows.ps1`

**Usage:**
```powershell
# Build for Windows (default)
.\build-windows.ps1

# Build for all platforms
.\build-windows.ps1 -All

# Clean and test before building
.\build-windows.ps1 -Clean -Test

# Build for specific platform
.\build-windows.ps1 -Target linux
.\build-windows.ps1 -Target mac

# Show help
.\build-windows.ps1 -Help
```

**Features:**
- Enhanced logging with detailed progress
- Comprehensive error handling
- Advanced testing capabilities
- Color-coded output with emojis
- Verbose build information

### 3. PowerShell Make System: `make.ps1` (Full Make Replacement)

**Usage:**
```powershell
# Build commands
.\make.ps1 build                # Current platform
.\make.ps1 build-windows        # Windows AMD64
.\make.ps1 build-linux          # Linux AMD64
.\make.ps1 build-linux-arm      # Linux ARM64
.\make.ps1 build-mac            # macOS Intel
.\make.ps1 build-mac-arm        # macOS Apple Silicon
.\make.ps1 build-all            # All platforms and architectures

# Development commands
.\make.ps1 test                 # Run tests
.\make.ps1 test-v               # Verbose tests
.\make.ps1 coverage             # Test coverage
.\make.ps1 clean                # Clean artifacts
.\make.ps1 deps                 # Update dependencies

# Utility commands
.\make.ps1 run                  # Build and run help
.\make.ps1 access               # Show API access info
.\make.ps1 install              # Install to GOPATH
.\make.ps1 help                 # Show all targets
```

**Features:**
- Full Makefile functionality in PowerShell
- ARM64 support for Apple Silicon and Linux
- Enhanced user experience with colors and emojis
- Comprehensive error handling
- Environment variable support

### 4. Batch Script: `make.bat`

**Usage:**
```cmd
# Works in Command Prompt
make.bat build
make.bat build-all
make.bat test
make.bat clean
make.bat help
```

**Features:**
- Compatible with Command Prompt and PowerShell
- Automatically calls the PowerShell make system
- No additional setup required

### 5. Global PowerShell Function: `PowerShell-Make-Function.ps1`

A PowerShell function that can be added to your profile for system-wide `make` command support.

**Setup:**
```powershell
# Add to PowerShell profile
notepad $PROFILE

# Add the contents of PowerShell-Make-Function.ps1
# Then reload profile
. $PROFILE

# Now use make globally
make build
make build-all
make test
```

## Build Outputs

| Platform | Architecture | Output File | Typical Size |
|----------|--------------|-------------|--------------|
| Windows | AMD64 | `meraki-info.exe` | 9.4 MB |
| Linux | AMD64 | `meraki-info-linux` | 9.3 MB |
| Linux | ARM64 | `meraki-info-linux-arm` | 8.8 MB |
| macOS | AMD64 (Intel) | `meraki-info-mac` | 9.3 MB |
| macOS | ARM64 (Apple Silicon) | `meraki-info-mac-arm` | 8.8 MB |

## Quick Start Guide

### For Most Users (Simple PowerShell)
```powershell
# Quick Windows build
.\build.ps1

# Build for all platforms
.\build.ps1 -All

# Clean build with tests
.\build.ps1 -Clean -Test
```

### For Make Users (PowerShell Make System)
```powershell
# Traditional make-like experience
.\make.ps1 build-all      # All platforms including ARM64
.\make.ps1 test           # Run tests
.\make.ps1 clean          # Clean artifacts
.\make.ps1 help           # Show all options
```

### For Command Prompt Users
```cmd
# Simple batch wrapper
make.bat build
make.bat test
make.bat clean
```

## Platform-Specific Features

### ARM64 Support
The build system now includes comprehensive ARM64 support:

- **Apple Silicon Macs**: Native ARM64 builds with `build-mac-arm`
- **ARM64 Linux**: Optimized for cloud instances with `build-linux-arm`
- **Automatic Detection**: Smart platform detection for current architecture
- **Size Optimization**: ARM64 binaries are typically 5% smaller

### Cross-Platform Compatibility
All scripts handle platform differences intelligently:

- **Executable Extensions**: `.exe` on Windows, no extension on Unix
- **Command Availability**: Graceful fallbacks for missing commands
- **Path Handling**: Proper path separators for each platform
- **Environment Variables**: Platform-aware environment handling

## Troubleshooting

### PowerShell Execution Policy
If you get execution policy errors:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Go Not Found
Ensure Go is installed and in your PATH:
```powershell
go version
```

### ARM64 Build Issues
Ensure you have Go 1.16+ for ARM64 support:
```powershell
go version
# Should show go1.16 or later
```

### Permission Issues
Run PowerShell as Administrator if you encounter permission issues.

## Integration with Traditional Tools

### Makefile Compatibility
These scripts provide the same functionality as Makefile targets:

| Makefile Target | PowerShell Equivalent | Description |
|----------------|----------------------|-------------|
| `make build` | `.\build.ps1` or `.\make.ps1 build` | Current platform |
| `make build-all` | `.\build.ps1 -All` or `.\make.ps1 build-all` | All platforms |
| `make build-mac-arm` | `.\make.ps1 build-mac-arm` | Apple Silicon |
| `make build-linux-arm` | `.\make.ps1 build-linux-arm` | Linux ARM64 |
| `make clean` | `.\build.ps1 -Clean` or `.\make.ps1 clean` | Clean artifacts |
| `make test` | `.\build.ps1 -Test` or `.\make.ps1 test` | Run tests |
| `make deps` | `.\make.ps1 deps` | Update dependencies |

### CI/CD Integration
All scripts provide proper exit codes for automation:

```powershell
# GitHub Actions example
- name: Build All Platforms
  run: .\make.ps1 build-all

# Azure DevOps example  
- powershell: .\build.ps1 -All -Test

# Jenkins example
powershell ".\make.ps1 build-all"
```

## Development Workflows

### Daily Development
```powershell
# Quick development cycle
.\make.ps1 clean          # Clean environment
.\make.ps1 test           # Ensure tests pass
.\make.ps1 build          # Build for current platform
.\make.ps1 run            # Test the executable
```

### Release Preparation
```powershell
# Full release build
.\make.ps1 clean          # Start clean
.\make.ps1 test           # Verify all tests
.\make.ps1 coverage       # Check test coverage
.\make.ps1 build-all      # Build all platforms
```

### Apple Silicon Development
```powershell
# Specific ARM64 workflows
.\make.ps1 build-mac-arm      # Native Apple Silicon
.\make.ps1 build-linux-arm    # ARM64 cloud deployment
```

## Performance Comparison

### Build Speed
- **PowerShell Scripts**: ~2-5 seconds per platform
- **Cross-compilation**: No performance penalty
- **ARM64 Builds**: Similar speed to AMD64

### Binary Optimization
- **ARM64 Binaries**: 5-10% smaller than AMD64
- **Native Performance**: ARM64 runs faster on Apple Silicon
- **Cloud Efficiency**: ARM64 Linux for better price/performance

## Advanced Usage

### Environment Variables
```powershell
# Set API key for access commands
$env:MERAKI_APIKEY = "your-api-key"
.\make.ps1 access

# Custom build configurations
$env:CGO_ENABLED = "0"        # Static builds
$env:GOFLAGS = "-ldflags=-s -w"  # Smaller binaries
```

### Custom Build Tags
```powershell
# Debug builds
$env:GOFLAGS = "-tags=debug"
.\make.ps1 build

# Production builds
$env:GOFLAGS = "-tags=production -ldflags=-s -w"
.\make.ps1 build-all
```

## Migration from Make

### Existing Make Users
1. **Keep using Makefile**: Enhanced with ARM64 support
2. **Try PowerShell Make**: `.\make.ps1` for enhanced experience
3. **Gradual Migration**: Use both systems side by side

### Windows-First Development
1. **Start with `build.ps1`**: Simple and effective
2. **Upgrade to `make.ps1`**: Full feature set
3. **Add to Profile**: Global `make` function

The PowerShell build system provides a comprehensive, Windows-native alternative to traditional make while maintaining full compatibility and adding modern features like ARM64 support and enhanced user experience.

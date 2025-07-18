# PowerShell Make Support for Meraki Info

This project provides comprehensive `make` support for PowerShell environments, delivering the same functionality as the traditional Makefile but optimized for Windows development with enhanced features and ARM64 support.

## 🎯 Overview

The PowerShell make system consists of multiple components that work together to provide a seamless build experience across all platforms and architectures, including modern ARM64 systems.

## 📁 Files Included

### 1. `make.ps1` - Main PowerShell Make Wrapper
A full-featured PowerShell script providing complete Makefile functionality with enhanced Windows support and ARM64 architecture support.

### 2. `make.bat` - Batch File Wrapper  
Enables running `make` commands from Command Prompt by automatically invoking the PowerShell script.

### 3. `PowerShell-Make-Function.ps1` - Global Make Function
A PowerShell function for system-wide `make` command support that can be added to your PowerShell profile.

### 4. `build.ps1` - Simple Build Script
Lightweight PowerShell script for basic building needs without the full make system complexity.

### 5. `build-windows.ps1` - Advanced Build Script
Feature-rich PowerShell script with detailed logging and comprehensive build options.

## 🚀 Quick Start

### Basic Usage
```powershell
# Show all available targets
.\make.ps1 help

# Build for current platform
.\make.ps1 build

# Build for all platforms and architectures
.\make.ps1 build-all

# Run tests
.\make.ps1 test

# Clean build artifacts
.\make.ps1 clean
```

### ARM64 Support
```powershell
# Build for Apple Silicon Macs
.\make.ps1 build-mac-arm

# Build for ARM64 Linux servers
.\make.ps1 build-linux-arm

# Build all architectures
.\make.ps1 build-all  # Includes ARM64 variants
```

### Using the Batch Wrapper
```cmd
# Works seamlessly in Command Prompt
make.bat build
make.bat build-all
make.bat test
make.bat clean
```

## 📋 Complete Target Reference

| Target | Description | Output File | Architecture |
|--------|-------------|-------------|--------------|
| `help` | Show available targets | - | - |
| `build` | Build for current platform | Platform-specific | Current |
| `build-windows` | Build for Windows | `meraki-info.exe` | AMD64 |
| `build-linux` | Build for Linux | `meraki-info-linux` | AMD64 |
| `build-linux-arm` | Build for Linux ARM64 | `meraki-info-linux-arm` | ARM64 |
| `build-mac` | Build for macOS Intel | `meraki-info-mac` | AMD64 |
| `build-mac-arm` | Build for macOS Apple Silicon | `meraki-info-mac-arm` | ARM64 |
| `build-all` | Build all platforms/architectures | All above | All |
| `test` | Run all tests | - | - |
| `test-v` | Run tests with verbose output | - | - |
| `coverage` | Run tests with coverage report | - | - |
| `clean` | Clean build artifacts | - | - |
| `run` | Build and run with help flag | - | - |
| `access` | Show API access information | - | - |
| `install` | Install to GOPATH/bin | - | - |
| `deps` | Update dependencies | - | - |

## 🔧 Setup Options

### Option 1: Project-Level Usage (Recommended)
Use the scripts directly in your project directory:
```powershell
.\make.ps1 build-all
```

### Option 2: Global Make Function
Add system-wide `make` command support:

1. **Edit your PowerShell profile:**
   ```powershell
   notepad $PROFILE
   ```

2. **Add the contents of `PowerShell-Make-Function.ps1`**

3. **Reload your profile:**
   ```powershell
   . $PROFILE
   ```

4. **Use `make` globally:**
   ```powershell
   cd any-project-with-make.ps1
   make build-all    # Automatically finds and uses make.ps1
   ```

### Option 3: PATH Integration (Advanced)
Add the project directory to your PATH environment variable for ultimate convenience.

## 🆚 Traditional Makefile Comparison

| Traditional Makefile | PowerShell Make | Status | Enhancement |
|---------------------|-----------------|--------|-------------|
| `make help` | `.\make.ps1 help` | ✅ Enhanced | Colored output, better formatting |
| `make build` | `.\make.ps1 build` | ✅ Enhanced | Smart platform detection |
| `make build-windows` | `.\make.ps1 build-windows` | ✅ Equivalent | Same functionality |
| `make build-linux` | `.\make.ps1 build-linux` | ✅ Equivalent | Same functionality |
| `make build-mac` | `.\make.ps1 build-mac` | ✅ Equivalent | Same functionality |
| - | `.\make.ps1 build-linux-arm` | ✅ New | ARM64 Linux support |
| - | `.\make.ps1 build-mac-arm` | ✅ New | Apple Silicon support |
| `make build-all` | `.\make.ps1 build-all` | ✅ Enhanced | Now includes ARM64 |
| `make test` | `.\make.ps1 test` | ✅ Enhanced | Better output formatting |
| `make test-v` | `.\make.ps1 test-v` | ✅ Equivalent | Same functionality |
| `make coverage` | `.\make.ps1 coverage` | ✅ Equivalent | Same functionality |
| `make clean` | `.\make.ps1 clean` | ✅ Enhanced | ARM64 artifact cleanup |
| `make run` | `.\make.ps1 run` | ✅ Enhanced | Smart executable detection |
| `make access` | `.\make.ps1 access` | ✅ Enhanced | Better error messages |
| `make install` | `.\make.ps1 install` | ✅ Equivalent | Same functionality |
| `make deps` | `.\make.ps1 deps` | ✅ Equivalent | Same functionality |

## ✨ PowerShell Enhancements

### 🎨 Enhanced User Experience
- **Colored Output**: Green for success, red for errors, yellow for warnings
- **Emoji Indicators**: 🔨 for building, ✅ for success, ❌ for errors
- **Progress Feedback**: Real-time build progress and status updates
- **File Size Reporting**: Shows binary sizes after successful builds

### 🛡️ Superior Error Handling
- **Detailed Error Messages**: Context-aware error reporting
- **Proper Exit Codes**: Full CI/CD integration support
- **Graceful Degradation**: Smart fallbacks for missing dependencies
- **Environment Validation**: Checks for required tools and variables

### 🔧 Windows Integration
- **Native PowerShell**: No external dependencies required
- **Command Prompt Compatible**: Works in both PowerShell and CMD
- **Smart Path Handling**: Handles Windows path conventions properly
- **Environment Variables**: Proper Windows environment support

### 📊 Advanced Features
- **ARM64 Support**: Complete support for Apple Silicon and ARM64 Linux
- **Smart Platform Detection**: Automatically detects current platform/architecture
- **Comprehensive Cleanup**: Removes all artifacts including ARM64 variants
- **Flexible Output**: Supports both stdout and file output modes

## 🌍 Platform & Architecture Matrix

| Platform | Architecture | Command | Output File | Size |
|----------|--------------|---------|-------------|------|
| Windows | AMD64 | `.\make.ps1 build-windows` | `meraki-info.exe` | ~9.4 MB |
| Linux | AMD64 | `.\make.ps1 build-linux` | `meraki-info-linux` | ~9.3 MB |
| Linux | ARM64 | `.\make.ps1 build-linux-arm` | `meraki-info-linux-arm` | ~8.8 MB |
| macOS | AMD64 (Intel) | `.\make.ps1 build-mac` | `meraki-info-mac` | ~9.3 MB |
| macOS | ARM64 (Apple Silicon) | `.\make.ps1 build-mac-arm` | `meraki-info-mac-arm` | ~8.8 MB |

**Note**: ARM64 binaries are typically 5-10% smaller due to more efficient instruction encoding.

## � Environment Variables

| Variable | Required For | Description | Example |
|----------|--------------|-------------|---------|
| `MERAKI_APIKEY` | `access` target | Meraki Dashboard API key | `your-api-key-here` |
| `GOOS` | Manual cross-compilation | Target operating system | `linux`, `windows`, `darwin` |
| `GOARCH` | Manual cross-compilation | Target architecture | `amd64`, `arm64` |

### Setting Environment Variables
```powershell
# Temporary (current session)
$env:MERAKI_APIKEY = "your-api-key-here"

# Permanent (user-level)
[Environment]::SetEnvironmentVariable("MERAKI_APIKEY", "your-api-key", "User")

# Permanent (system-level, requires admin)
[Environment]::SetEnvironmentVariable("MERAKI_APIKEY", "your-api-key", "Machine")
```

## 🚀 Usage Examples

### Development Workflow
```powershell
# Daily development cycle
.\make.ps1 clean              # Clean environment
.\make.ps1 test               # Run tests
.\make.ps1 build              # Build for current platform
.\make.ps1 run                # Test the executable

# Check API access
$env:MERAKI_APIKEY = "your-key"
.\make.ps1 access
```

### Release Preparation
```powershell
# Complete release workflow
.\make.ps1 clean              # Start clean
.\make.ps1 deps               # Update dependencies
.\make.ps1 test               # Ensure all tests pass
.\make.ps1 coverage           # Check test coverage
.\make.ps1 build-all          # Build all platforms/architectures

# Results in 5 binaries ready for distribution
```

### ARM64 Specific Workflows
```powershell
# Apple Silicon development
.\make.ps1 build-mac-arm      # Native Apple Silicon build
.\make.ps1 run                # Test on Apple Silicon

# ARM64 cloud deployment
.\make.ps1 build-linux-arm    # Build for ARM64 servers
# Deploy to AWS Graviton, Azure ARM, etc.
```

### CI/CD Integration
```powershell
# GitHub Actions example
.\make.ps1 deps
.\make.ps1 test
.\make.ps1 build-all

# Azure DevOps example
.\make.ps1 clean
.\make.ps1 test
.\make.ps1 coverage
.\make.ps1 build-all

# Jenkins example
.\make.ps1 build-all
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
```

## 🔍 Troubleshooting

### PowerShell Execution Policy
**Error**: "execution of scripts is disabled"
```powershell
# Solution: Set execution policy for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Missing Dependencies
**Error**: "'go' is not recognized"
```powershell
# Solution: Install Go and verify PATH
go version
# If not found, install Go from https://golang.org/dl/
```

### ARM64 Build Issues
**Error**: "unsupported GOOS/GOARCH pair"
```powershell
# Solution: Ensure Go 1.16+ for ARM64 support
go version
# Should show go1.16 or later
```

### Make Function Not Working
**Issue**: Global `make` command not found
```powershell
# Check profile exists
Test-Path $PROFILE

# Create profile if missing
New-Item -Path $PROFILE -Type File -Force

# Add function to profile
# Copy contents of PowerShell-Make-Function.ps1

# Reload profile
. $PROFILE

# Verify function
Get-Command make
```

### Permission Issues
**Error**: Access denied or permission errors
```powershell
# Solution: Run PowerShell as Administrator
# Or check file permissions
Get-Acl .\make.ps1
```

## 🔄 Migration Guide

### From Traditional Make
1. **Keep Existing Workflow**: Enhanced Makefile still works
2. **Try PowerShell Make**: Use `.\make.ps1` for enhanced experience
3. **Gradual Adoption**: Use both systems during transition
4. **Full Migration**: Eventually replace all make calls with PowerShell

### From Basic Build Scripts
1. **Start Simple**: Use `.\build.ps1` for basic needs
2. **Add Features**: Upgrade to `.\make.ps1` for full functionality
3. **Global Access**: Add function to PowerShell profile
4. **Team Adoption**: Share setup across development team

## 🤝 Contributing

### Adding New Targets
When extending the PowerShell make system:

1. **Add Target**: Include in switch statement in `make.ps1`
2. **Create Function**: Follow `Invoke-TargetName` naming pattern
3. **Update Help**: Add description in `Show-MakeHelp` function
4. **Test Thoroughly**: Verify on both PowerShell and Command Prompt
5. **Update Documentation**: Maintain this file and examples

### Best Practices
- **Use Approved Verbs**: Follow PowerShell verb guidelines
- **Error Handling**: Proper exit codes and error messages
- **User Feedback**: Clear progress indicators and success messages
- **Cross-Platform**: Test on different Windows versions and PowerShell editions

## 📞 Support

### Common Issues
1. **Check Documentation**: Review troubleshooting section
2. **Verify PowerShell Version**: `$PSVersionTable.PSVersion`
3. **Test Basic Scripts**: Try `.\build.ps1` first
4. **Check Environment**: Verify Go installation and PATH

### Getting Help
- **GitHub Issues**: Report bugs and feature requests
- **Documentation**: Comprehensive guides in repository
- **Examples**: Working examples in this file
- **Community**: PowerShell and Go community resources

---

## 🎯 Summary

The PowerShell make system provides a comprehensive, Windows-native alternative to traditional make while offering:

- **✅ Full Compatibility**: All Makefile functionality preserved
- **✅ Enhanced Experience**: Better UI, error handling, and feedback
- **✅ ARM64 Support**: Complete support for modern architectures
- **✅ Windows Integration**: Native PowerShell with Command Prompt compatibility
- **✅ Future-Proof**: Ready for evolving development ecosystems

Whether you're developing on Windows, deploying to ARM64 cloud instances, or building for Apple Silicon, the PowerShell make system provides the tools you need with an enhanced developer experience.

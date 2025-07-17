# PowerShell Make Support for Meraki Info

This project now includes comprehensive `make` support for PowerShell environments, providing the same functionality as the original Makefile but optimized for Windows development.

## 🎯 Files Added

### 1. `make.ps1` - Main PowerShell Make Wrapper
A full-featured PowerShell script that provides all Makefile functionality with enhanced Windows support.

### 2. `make.bat` - Batch File Wrapper  
Allows running `make` commands from Command Prompt by calling the PowerShell script.

### 3. `PowerShell-Make-Function.ps1` - Global Make Function
A PowerShell function that can be added to your profile for system-wide `make` command support.

## 🚀 Quick Start

### Basic Usage
```powershell
# Show help
.\make.ps1 help

# Build for Windows
.\make.ps1 build

# Build for all platforms
.\make.ps1 build-all

# Run tests
.\make.ps1 test

# Clean build artifacts
.\make.ps1 clean
```

### Using the Batch Wrapper
```cmd
# Works in Command Prompt too
make.bat build
make.bat test
make.bat clean
```

## 📋 Available Targets

| Target | Description | Example |
|--------|-------------|---------|
| `help` | Show available targets | `.\make.ps1 help` |
| `build` | Build for current platform (Windows) | `.\make.ps1 build` |
| `build-windows` | Build for Windows x64 | `.\make.ps1 build-windows` |
| `build-linux` | Build for Linux x64 | `.\make.ps1 build-linux` |
| `build-mac` | Build for macOS x64 | `.\make.ps1 build-mac` |
| `build-all` | Build for all platforms | `.\make.ps1 build-all` |
| `test` | Run all tests | `.\make.ps1 test` |
| `test-v` | Run tests with verbose output | `.\make.ps1 test-v` |
| `coverage` | Run tests with coverage report | `.\make.ps1 coverage` |
| `clean` | Clean build artifacts | `.\make.ps1 clean` |
| `run` | Build and run with help flag | `.\make.ps1 run` |
| `access` | Show API access information | `.\make.ps1 access` |
| `install` | Install to GOPATH/bin | `.\make.ps1 install` |
| `deps` | Update dependencies | `.\make.ps1 deps` |

## 🔧 Setup Options

### Option 1: Project-Level Usage (Recommended)
Just use the scripts directly in your project:
```powershell
.\make.ps1 build
```

### Option 2: Global Make Function
Add the function to your PowerShell profile for system-wide access:

1. **Edit your PowerShell profile:**
   ```powershell
   notepad $PROFILE
   ```

2. **Add the contents of `PowerShell-Make-Function.ps1` to your profile**

3. **Reload your profile:**
   ```powershell
   . $PROFILE
   ```

4. **Now you can use `make` anywhere:**
   ```powershell
   make build
   make test
   make clean
   ```

### Option 3: Add to PATH (Advanced)
For ultimate convenience, add the project directory to your PATH environment variable.

## 🆚 Comparison with Original Makefile

| Makefile Target | PowerShell Equivalent | Status |
|-----------------|----------------------|--------|
| `make help` | `.\make.ps1 help` | ✅ Enhanced |
| `make build` | `.\make.ps1 build` | ✅ Equivalent |
| `make build-windows` | `.\make.ps1 build-windows` | ✅ Equivalent |
| `make build-linux` | `.\make.ps1 build-linux` | ✅ Equivalent |
| `make build-mac` | `.\make.ps1 build-mac` | ✅ Equivalent |
| `make build-all` | `.\make.ps1 build-all` | ✅ Equivalent |
| `make test` | `.\make.ps1 test` | ✅ Equivalent |
| `make test-v` | `.\make.ps1 test-v` | ✅ Equivalent |
| `make coverage` | `.\make.ps1 coverage` | ✅ Equivalent |
| `make clean` | `.\make.ps1 clean` | ✅ Enhanced |
| `make run` | `.\make.ps1 run` | ✅ Equivalent |
| `make access` | `.\make.ps1 access` | ✅ Equivalent |
| `make install` | `.\make.ps1 install` | ✅ Equivalent |
| `make deps` | `.\make.ps1 deps` | ✅ Equivalent |

## ✨ PowerShell Enhancements

The PowerShell version includes several improvements over the original Makefile:

### 🎨 Enhanced Output
- **Colored output** with emojis for better readability
- **Progress indicators** for long-running operations
- **Clear success/error messages**

### 🛡️ Better Error Handling
- **Detailed error messages** with context
- **Proper exit codes** for CI/CD integration
- **Graceful failure handling**

### 🔧 Windows Integration
- **Native PowerShell integration**
- **Works in both PowerShell and Command Prompt**
- **No external dependencies** (no need to install make)

### 📊 Improved Feedback
- **File size reporting** for built executables
- **Build time information**
- **Detailed help with examples**

## 🌍 Environment Variables

| Variable | Required For | Description |
|----------|--------------|-------------|
| `MERAKI_APIKEY` | `access` target | Your Meraki Dashboard API key |

### Setting Environment Variables
```powershell
# Set for current session
$env:MERAKI_APIKEY = "your-api-key-here"

# Set permanently (requires admin)
[Environment]::SetEnvironmentVariable("MERAKI_APIKEY", "your-api-key", "User")
```

## 🚀 Examples

### Development Workflow
```powershell
# Clean build environment
.\make.ps1 clean

# Run tests to ensure code quality
.\make.ps1 test

# Build for all platforms
.\make.ps1 build-all

# Test the Windows executable
.\make.ps1 run
```

### CI/CD Integration
```powershell
# In your CI/CD pipeline
.\make.ps1 deps    # Update dependencies
.\make.ps1 test    # Run tests
.\make.ps1 build-all # Build all platforms
```

### API Testing
```powershell
# Set your API key
$env:MERAKI_APIKEY = "your-key-here"

# Check API access
.\make.ps1 access
```

## 🔍 Troubleshooting

### PowerShell Execution Policy
If you get execution policy errors:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Missing Dependencies
Ensure Go is installed and in your PATH:
```powershell
go version
```

### Make Function Not Working
If the global `make` function isn't working:
1. Check your profile exists: `Test-Path $PROFILE`
2. Reload your profile: `. $PROFILE`
3. Verify the function: `Get-Command make`

## 🤝 Contributing

When adding new targets to the PowerShell make system:

1. **Add the target** to the switch statement in `make.ps1`
2. **Create a function** following the `Invoke-TargetName` pattern
3. **Update the help** in `Show-MakeHelp` function
4. **Test thoroughly** on both PowerShell and Command Prompt
5. **Update documentation**

## 📞 Support

For issues specific to the PowerShell make implementation:
1. Check this documentation
2. Verify your PowerShell version: `$PSVersionTable.PSVersion`
3. Test with the original build scripts (`build.ps1`, `build-windows.ps1`)
4. Check the GitHub repository for updates

---

**Note**: The PowerShell make system is designed to be a drop-in replacement for the original Makefile while providing enhanced Windows support and better user experience.

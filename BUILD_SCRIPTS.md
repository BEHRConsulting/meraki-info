# Build Scripts for Windows

This directory contains Windows-specific build scripts for the Meraki Info application since `make` is not readily available in Windows PowerShell environments.

## Available Scripts

### 1. PowerShell Script: `build.ps1` (Recommended)

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
- Clean, simple output with emojis
- Cross-platform building (Windows, Linux, macOS)
- Automatic testing of Windows executable
- File size reporting
- Error handling with proper exit codes

### 2. Batch Script: `build-windows.bat`

**Usage:**
```cmd
# Build for Windows (default)
build-windows.bat

# Build for all platforms
build-windows.bat --all

# Clean and test before building
build-windows.bat --clean --test

# Build for specific platform
build-windows.bat --target linux
build-windows.bat --target mac

# Show help
build-windows.bat --help
```

**Features:**
- Compatible with Command Prompt and PowerShell
- Same functionality as PowerShell script
- Cross-platform building support
- Automatic testing and validation

### 3. Advanced PowerShell Script: `build-windows.ps1`

A more feature-rich version with detailed logging and enhanced functionality.

## Build Outputs

| Platform | Output File | Description |
|----------|-------------|-------------|
| Windows | `meraki-info.exe` | Windows executable |
| Linux | `meraki-info-linux` | Linux binary |
| macOS | `meraki-info-mac` | macOS binary |

## Quick Start

For most users, the simple PowerShell script is recommended:

```powershell
# Build for Windows
.\build.ps1

# Or build for all platforms
.\build.ps1 -All
```

## Troubleshooting

### PowerShell Execution Policy

If you get an execution policy error, run:
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Go Not Found

Ensure Go is installed and in your PATH:
```powershell
go version
```

### Permission Issues

Run PowerShell as Administrator if you encounter permission issues.

## Integration with Makefile

These scripts provide the same functionality as the Makefile targets:

| Makefile Target | PowerShell Equivalent | Batch Equivalent |
|----------------|----------------------|------------------|
| `make build-windows` | `.\build.ps1` | `build-windows.bat` |
| `make build-all` | `.\build.ps1 -All` | `build-windows.bat --all` |
| `make clean` | `.\build.ps1 -Clean` | `build-windows.bat --clean` |
| `make test` | `.\build.ps1 -Test` | `build-windows.bat --test` |

## Examples

```powershell
# Quick Windows build
.\build.ps1

# Full development workflow
.\build.ps1 -Clean -Test -All

# Just clean and build for Windows
.\build.ps1 -Clean

# Build for deployment to Linux server
.\build.ps1 -Target linux
```

<#
.SYNOPSIS
    Meraki Info Windows Build Script.

.DESCRIPTION
    PowerShell script to build the Meraki Info application for Windows.

    The longer sibling of build.ps1 with the same options: optionally clean
    and test, then cross-compile meraki-info for one platform (amd64) or
    for Windows, Linux and macOS together, smoke-test meraki-info.exe with
    --help, and list the binaries that were built.

.PARAMETER Target
    Platform to build when -All is not given: windows (default), linux or
    mac. Case-insensitive.

.PARAMETER All
    Build Windows, Linux and macOS (amd64) instead of a single -Target.

.PARAMETER Clean
    Remove previous meraki-info binaries and run `go clean` first.

.PARAMETER Test
    Run `go test ./...` before building; exit 1 if any test fails.

.PARAMETER Help
    Print a usage summary and exit.

.EXAMPLE
    .\build-windows.ps1

.EXAMPLE
    .\build-windows.ps1 -Target linux

.EXAMPLE
    .\build-windows.ps1 -Clean -Test -All

.INPUTS
    None. You cannot pipe objects to this script.

.OUTPUTS
    None. Progress is written to the host; binaries are written to the
    current directory.

.NOTES
    Exits 1 on a failed test run, an unknown target or a failed build.
#>

param(
    [string]$Target = "windows",
    [switch]$All = $false,
    [switch]$Clean = $false,
    [switch]$Test = $false,
    [switch]$Help = $false
)

# Function to display help
function Show-Help {
    <#
    .SYNOPSIS
        Prints this script's usage summary.

    .OUTPUTS
        None. Writes to the host.
    #>
    Write-Host "Meraki Info Windows Build Script" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage: .\build-windows.ps1 [OPTIONS]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "OPTIONS:" -ForegroundColor Yellow
    Write-Host "  -Target <platform>   Build for specific platform: windows, linux, mac (default: windows)"
    Write-Host "  -All                 Build for all platforms"
    Write-Host "  -Clean               Clean build artifacts before building"
    Write-Host "  -Test                Run tests before building"
    Write-Host "  -Help                Show this help message"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build-windows.ps1                    # Build for Windows"
    Write-Host "  .\build-windows.ps1 -Target linux      # Build for Linux"
    Write-Host "  .\build-windows.ps1 -All               # Build for all platforms"
    Write-Host "  .\build-windows.ps1 -Clean -Test       # Clean, test, then build"
    Write-Host ""
}

# Function to clean build artifacts
function Remove-BuildArtifacts {
    <#
    .SYNOPSIS
        Deletes previously built meraki-info binaries and runs `go clean`.

    .DESCRIPTION
        Removes meraki-info.exe, meraki-info-linux, meraki-info-mac and
        meraki-info from the current directory when present. ARM64
        binaries (meraki-info-*-arm) are not removed.

    .OUTPUTS
        None. Writes progress to the host.
    #>
    Write-Host "🧹 Cleaning build artifacts..." -ForegroundColor Yellow
    
    $artifacts = @(
        "meraki-info.exe",
        "meraki-info-linux",
        "meraki-info-mac",
        "meraki-info"
    )
    
    foreach ($artifact in $artifacts) {
        if (Test-Path $artifact) {
            Remove-Item $artifact -Force
            Write-Host "  Removed: $artifact" -ForegroundColor Gray
        }
    }
    
    # Run go clean
    go clean
    Write-Host "✅ Build artifacts cleaned" -ForegroundColor Green
}

# Function to run tests
function Invoke-Tests {
    <#
    .SYNOPSIS
        Runs `go test ./...`.

    .DESCRIPTION
        Exits the whole script with 1 if any test fails.

    .OUTPUTS
        None. Writes progress to the host.
    #>
    Write-Host "🧪 Running tests..." -ForegroundColor Yellow
    
    go test ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Tests failed!" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ All tests passed" -ForegroundColor Green
}

# Function to build for a specific platform
function Build-Platform {
    <#
    .SYNOPSIS
        Cross-compiles meraki-info for one platform.

    .DESCRIPTION
        Sets GOOS/GOARCH, runs `go build -o <OutputName> .`, reports the
        binary size, then clears GOOS/GOARCH. Exits the whole script with 1
        if the build fails.

    .PARAMETER Platform
        Display name used in messages, for example "Windows".

    .PARAMETER GOOS
        Value for GOOS: windows, linux or darwin.

    .PARAMETER GOARCH
        Value for GOARCH, for example amd64.

    .PARAMETER OutputName
        Output file name for the binary.

    .OUTPUTS
        None. Writes progress to the host.
    #>
    param(
        [string]$Platform,
        [string]$GOOS,
        [string]$GOARCH,
        [string]$OutputName
    )
    
    Write-Host "🔨 Building for $Platform..." -ForegroundColor Yellow
    
    # Set environment variables for cross-compilation
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    
    # Build the application
    $buildCommand = "go build -o $OutputName ."
    Write-Host "  Command: $buildCommand" -ForegroundColor Gray
    
    go build -o $OutputName .
    
    if ($LASTEXITCODE -eq 0) {
        $fileInfo = Get-Item $OutputName
        $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
        Write-Host "✅ $Platform build completed: $OutputName ($sizeMB MB)" -ForegroundColor Green
    } else {
        Write-Host "❌ $Platform build failed!" -ForegroundColor Red
        exit 1
    }
    
    # Reset environment variables
    Remove-Item env:GOOS -ErrorAction SilentlyContinue
    Remove-Item env:GOARCH -ErrorAction SilentlyContinue
}

# Function to test the Windows executable
function Test-WindowsExecutable {
    <#
    .SYNOPSIS
        Smoke-tests meraki-info.exe by running it with --help.

    .DESCRIPTION
        Does nothing if meraki-info.exe is absent. A failure is reported
        but does not stop the script.

    .OUTPUTS
        None. Writes the result to the host.
    #>
    if (Test-Path "meraki-info.exe") {
        Write-Host "🧪 Testing Windows executable..." -ForegroundColor Yellow
        
        $testOutput = .\meraki-info.exe --help 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Windows executable test passed" -ForegroundColor Green
        } else {
            Write-Host "❌ Windows executable test failed!" -ForegroundColor Red
            Write-Host $testOutput -ForegroundColor Red
        }
    }
}

# Main execution
if ($Help) {
    Show-Help
    exit 0
}

Write-Host "🚀 Starting Meraki Info build process..." -ForegroundColor Green
Write-Host ""

# Clean if requested
if ($Clean) {
    Remove-BuildArtifacts
    Write-Host ""
}

# Run tests if requested
if ($Test) {
    Invoke-Tests
    Write-Host ""
}

# Build based on target
if ($All) {
    Write-Host "Building for all platforms..." -ForegroundColor Cyan
    Build-Platform "Windows" "windows" "amd64" "meraki-info.exe"
    Build-Platform "Linux" "linux" "amd64" "meraki-info-linux"
    Build-Platform "macOS" "darwin" "amd64" "meraki-info-mac"
    
    # Test Windows executable if it was built
    Test-WindowsExecutable
} else {
    switch ($Target.ToLower()) {
        "windows" {
            Build-Platform "Windows" "windows" "amd64" "meraki-info.exe"
            Test-WindowsExecutable
        }
        "linux" {
            Build-Platform "Linux" "linux" "amd64" "meraki-info-linux"
        }
        "mac" {
            Build-Platform "macOS" "darwin" "amd64" "meraki-info-mac"
        }
        default {
            Write-Host "❌ Unknown target: $Target" -ForegroundColor Red
            Write-Host "Valid targets: windows, linux, mac" -ForegroundColor Yellow
            exit 1
        }
    }
}

Write-Host ""
Write-Host "🎉 Build process completed successfully!" -ForegroundColor Green

# Show built files
Write-Host ""
Write-Host "📦 Built files:" -ForegroundColor Cyan
Get-ChildItem meraki-info* | Where-Object { $_.Name -match "meraki-info.*\.(exe|$)" } | ForEach-Object {
    $sizeMB = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $($_.Name) ($sizeMB MB)" -ForegroundColor White
}

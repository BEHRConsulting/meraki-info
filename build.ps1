<#
.SYNOPSIS
    Simple Meraki Info Build Script for Windows.

.DESCRIPTION
    PowerShell script to build the Meraki Info application.

    Optionally cleans old binaries and runs `go test ./...`, then
    cross-compiles meraki-info for one platform (amd64 only) or for
    Windows, Linux and macOS together. If meraki-info.exe exists afterwards
    it is smoke-tested with --help. ARM64 builds are not offered here; use
    make.ps1 build-linux-arm / build-mac-arm.

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
    Print a short usage summary and exit.

.EXAMPLE
    .\build.ps1
    Builds meraki-info.exe for Windows amd64.

.EXAMPLE
    .\build.ps1 -All

.EXAMPLE
    .\build.ps1 -Clean -Test -Target linux

.INPUTS
    None. You cannot pipe objects to this script.

.OUTPUTS
    None. Progress is written to the host; binaries are written to the
    current directory.

.NOTES
    Exits 1 on a failed test run, an unknown target or a failed build.
    GOOS/GOARCH are set in the session environment for each build and
    removed afterwards.
#>

param(
    [string]$Target = "windows",
    [switch]$All,
    [switch]$Clean,
    [switch]$Test,
    [switch]$Help
)

if ($Help) {
    Write-Host "Meraki Info Build Script" -ForegroundColor Green
    Write-Host ""
    Write-Host "Usage: .\build.ps1 [OPTIONS]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "OPTIONS:" -ForegroundColor Yellow
    Write-Host "  -Target <platform>   Build for: windows, linux, mac (default: windows)"
    Write-Host "  -All                 Build for all platforms"
    Write-Host "  -Clean               Clean build artifacts first"
    Write-Host "  -Test                Run tests before building"
    Write-Host "  -Help                Show this help"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1                    # Build for Windows"
    Write-Host "  .\build.ps1 -All               # Build for all platforms"
    Write-Host "  .\build.ps1 -Clean -Test       # Clean, test, then build"
    exit 0
}

Write-Host "🚀 Meraki Info Build Process" -ForegroundColor Green
Write-Host ""

# Clean artifacts
if ($Clean) {
    Write-Host "🧹 Cleaning..." -ForegroundColor Yellow
    Remove-Item -Path "meraki-info*.exe", "meraki-info-linux", "meraki-info-mac", "meraki-info" -ErrorAction SilentlyContinue
    go clean
    Write-Host "✅ Cleaned" -ForegroundColor Green
    Write-Host ""
}

# Run tests
if ($Test) {
    Write-Host "🧪 Testing..." -ForegroundColor Yellow
    go test ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Tests failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Tests passed" -ForegroundColor Green
    Write-Host ""
}

# Build function
function BuildApp($platform, $goos, $goarch, $output) {
    <#
    .SYNOPSIS
        Cross-compiles meraki-info for one platform.

    .DESCRIPTION
        Sets GOOS/GOARCH, runs `go build -o <output> .`, reports the binary
        size, then clears GOOS/GOARCH. Exits the whole script with 1 if the
        build fails.

    .PARAMETER platform
        Display name used in messages, for example "Windows".

    .PARAMETER goos
        Value for GOOS: windows, linux or darwin.

    .PARAMETER goarch
        Value for GOARCH, for example amd64.

    .PARAMETER output
        Output file name for the binary.

    .OUTPUTS
        None. Writes progress to the host.
    #>
    Write-Host "🔨 Building $platform..." -ForegroundColor Yellow
    $env:GOOS = $goos
    $env:GOARCH = $goarch
    go build -o $output .
    if ($LASTEXITCODE -eq 0) {
        $size = (Get-Item $output).Length / 1MB
        $sizeMB = [math]::Round($size, 2)
        Write-Host ("✅ {0}: {1} ({2} MB)" -f $platform, $output, $sizeMB) -ForegroundColor Green
    } else {
        Write-Host "❌ $platform build failed!" -ForegroundColor Red
        exit 1
    }
    Remove-Item env:GOOS, env:GOARCH -ErrorAction SilentlyContinue
}

# Build targets
if ($All) {
    BuildApp "Windows" "windows" "amd64" "meraki-info.exe"
    BuildApp "Linux" "linux" "amd64" "meraki-info-linux"
    BuildApp "macOS" "darwin" "amd64" "meraki-info-mac"
} else {
    switch ($Target.ToLower()) {
        "windows" { BuildApp "Windows" "windows" "amd64" "meraki-info.exe" }
        "linux"   { BuildApp "Linux" "linux" "amd64" "meraki-info-linux" }
        "mac"     { BuildApp "macOS" "darwin" "amd64" "meraki-info-mac" }
        default   { 
            Write-Host "❌ Unknown target: $Target" -ForegroundColor Red
            Write-Host "Valid: windows, linux, mac" -ForegroundColor Yellow
            exit 1
        }
    }
}

# Test Windows executable if built
if (Test-Path "meraki-info.exe") {
    Write-Host "🧪 Testing executable..." -ForegroundColor Yellow
    .\meraki-info.exe --help | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Executable works" -ForegroundColor Green
    } else {
        Write-Host "❌ Executable test failed" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "🎉 Build complete!" -ForegroundColor Green

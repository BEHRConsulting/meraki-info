# Simple Meraki Info Build Script for Windows
# PowerShell script to build the Meraki Info application

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

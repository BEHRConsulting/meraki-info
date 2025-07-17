# Meraki Info Windows Build Script
# PowerShell script to build the Meraki Info application for Windows

param(
    [string]$Target = "windows",
    [switch]$All = $false,
    [switch]$Clean = $false,
    [switch]$Test = $false,
    [switch]$Help = $false
)

# Function to display help
function Show-Help {
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

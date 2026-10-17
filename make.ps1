<#
.SYNOPSIS
    PowerShell Make Wrapper for Meraki Info.

.DESCRIPTION
    This script provides make-like functionality for PowerShell environments
    Usage: .\make.ps1 <target> [arguments]

    It mirrors the Makefile's targets for Windows users without make.
    Single-platform amd64 builds are delegated to build.ps1; the ARM64
    targets are built here directly.

    Targets: help, build, build-windows, build-linux, build-linux-arm,
    build-mac, build-mac-arm, build-all, test, test-v, coverage, clean, run,
    access, install, deps.

.PARAMETER Target
    The target to run. Defaults to help. Case-insensitive.

.PARAMETER Arguments
    Extra arguments. Accepted for make-style invocation but not used by any
    target.

.EXAMPLE
    .\make.ps1 build-all

.EXAMPLE
    $env:MERAKI_APIKEY = 'your-key'; .\make.ps1 access

.INPUTS
    None. You cannot pipe objects to this script.

.OUTPUTS
    None. Progress is written to the host.

.NOTES
    Exits 1 for an unknown target, and with the failing command's exit code
    when a build, test or install fails.
#>

param(
    [Parameter(Position=0)]
    [string]$Target = "help",
    
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Arguments = @()
)

# Colors for output
$Colors = @{
    Success = 'Green'
    Warning = 'Yellow'
    Error = 'Red'
    Info = 'Cyan'
    Target = 'Magenta'
}

function Write-MakeOutput {
    <#
    .SYNOPSIS
        Writes a message to the host in one of the script's named colors.

    .PARAMETER Message
        The text to write.

    .PARAMETER Color
        A key of $Colors (Success, Warning, Error, Info, Target). Any other
        value, including the default 'White', is written in white.

    .OUTPUTS
        None. Writes to the host.
    #>
    param(
        [string]$Message,
        [string]$Color = 'White'
    )
    
    $colorValue = $Colors[$Color]
    if (-not $colorValue) {
        $colorValue = 'White'
    }
    Write-Host $Message -ForegroundColor $colorValue
}

function Show-MakeHelp {
    <#
    .SYNOPSIS
        Prints the list of targets, examples and environment variables.

    .OUTPUTS
        None. Writes to the host.
    #>
    Write-MakeOutput "Meraki Info - PowerShell Make Wrapper" -Color Success
    Write-MakeOutput ""
    Write-MakeOutput "Available targets:" -Color Info
    Write-MakeOutput "  help          Show this help message" -Color Target
    Write-MakeOutput "  build         Build the application for current platform" -Color Target
    Write-MakeOutput "  build-windows Build for Windows (x64)" -Color Target
    Write-MakeOutput "  build-linux   Build for Linux (x64)" -Color Target
    Write-MakeOutput "  build-linux-arm Build for Linux (ARM64)" -Color Target
    Write-MakeOutput "  build-mac     Build for macOS (x64 - Intel)" -Color Target
    Write-MakeOutput "  build-mac-arm Build for macOS (ARM64 - Apple Silicon)" -Color Target
    Write-MakeOutput "  build-all     Build for all platforms and architectures" -Color Target
    Write-MakeOutput "  test          Run all tests" -Color Target
    Write-MakeOutput "  test-v        Run tests with verbose output" -Color Target
    Write-MakeOutput "  coverage      Run tests with coverage report" -Color Target
    Write-MakeOutput "  clean         Clean build artifacts" -Color Target
    Write-MakeOutput "  run           Build and run with help flag" -Color Target
    Write-MakeOutput "  access        Show available organizations and networks" -Color Target
    Write-MakeOutput "  install       Install the application to GOPATH/bin" -Color Target
    Write-MakeOutput "  deps          Download and organize dependencies" -Color Target
    Write-MakeOutput ""
    Write-MakeOutput "Examples:" -Color Info
    Write-MakeOutput "  .\make.ps1 build         # Build for current platform"
    Write-MakeOutput "  .\make.ps1 build-all     # Build for all platforms"
    Write-MakeOutput "  .\make.ps1 test          # Run tests"
    Write-MakeOutput "  .\make.ps1 clean         # Clean build artifacts"
    Write-MakeOutput ""
    Write-MakeOutput "Environment Variables:" -Color Info
    Write-MakeOutput "  MERAKI_APIKEY    Required for 'access' target"
    Write-MakeOutput ""
}

function Invoke-BuildTarget {
    <#
    .SYNOPSIS
        Builds meraki-info for one platform or for all of them.

    .DESCRIPTION
        windows, linux and mac delegate to build.ps1 -Target; linux-arm and
        mac-arm are built here with GOOS/GOARCH set; all builds every
        platform. Any other value runs build.ps1 with its defaults (Windows).
        After a delegated build, exits the script with that build's exit
        code if it failed.

    .PARAMETER Platform
        windows (default), linux, linux-arm, mac, mac-arm or all.

    .OUTPUTS
        None. Writes progress to the host.
    #>
    param([string]$Platform = "windows")
    
    Write-MakeOutput "🔨 Building for $Platform..." -Color Info
    
    switch ($Platform.ToLower()) {
        "windows" {
            & .\build.ps1 -Target windows
        }
        "linux" {
            & .\build.ps1 -Target linux
        }
        "linux-arm" {
            Write-MakeOutput "🔨 Building for Linux ARM64..." -Color Info
            $env:GOOS = "linux"
            $env:GOARCH = "arm64"
            go build -o meraki-info-linux-arm .
            if ($LASTEXITCODE -eq 0) {
                Write-MakeOutput "✅ Linux ARM64 build completed: meraki-info-linux-arm" -Color Success
            }
            Remove-Item env:GOOS, env:GOARCH -ErrorAction SilentlyContinue
            return
        }
        "mac" {
            & .\build.ps1 -Target mac
        }
        "mac-arm" {
            Write-MakeOutput "🔨 Building for macOS ARM64 (Apple Silicon)..." -Color Info
            $env:GOOS = "darwin"
            $env:GOARCH = "arm64"
            go build -o meraki-info-mac-arm .
            if ($LASTEXITCODE -eq 0) {
                Write-MakeOutput "✅ macOS ARM64 build completed: meraki-info-mac-arm" -Color Success
            }
            Remove-Item env:GOOS, env:GOARCH -ErrorAction SilentlyContinue
            return
        }
        "all" {
            Write-MakeOutput "Building for all platforms and architectures..." -Color Info
            & .\build.ps1 -Target windows
            & .\build.ps1 -Target linux
            Invoke-BuildTarget "linux-arm"
            & .\build.ps1 -Target mac
            Invoke-BuildTarget "mac-arm"
            Write-MakeOutput "🎉 All platform builds completed!" -Color Success
            return
        }
        default {
            & .\build.ps1
        }
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-MakeOutput "❌ Build failed!" -Color Error
        exit $LASTEXITCODE
    }
}

function Invoke-TestTarget {
    <#
    .SYNOPSIS
        Runs the Go tests.

    .PARAMETER Verbose
        Run `go test -v ./...`.

    .PARAMETER Coverage
        Run `go test -cover ./...`. Takes precedence over -Verbose.

    .OUTPUTS
        None. Exits the script with the test exit code on failure.
    #>
    param([switch]$Verbose, [switch]$Coverage)
    
    if ($Coverage) {
        Write-MakeOutput "🧪 Running tests with coverage..." -Color Info
        go test -cover ./...
    } elseif ($Verbose) {
        Write-MakeOutput "🧪 Running tests (verbose)..." -Color Info
        go test -v ./...
    } else {
        Write-MakeOutput "🧪 Running tests..." -Color Info
        go test ./...
    }
    
    if ($LASTEXITCODE -ne 0) {
        Write-MakeOutput "❌ Tests failed!" -Color Error
        exit $LASTEXITCODE
    } else {
        Write-MakeOutput "✅ Tests passed!" -Color Success
    }
}

function Invoke-CleanTarget {
    <#
    .SYNOPSIS
        Deletes every built meraki-info binary and runs `go clean`.

    .OUTPUTS
        None. Writes each removed file to the host.
    #>
    Write-MakeOutput "🧹 Cleaning build artifacts..." -Color Info
    
    # Remove built executables
    $artifacts = @(
        "meraki-info.exe",
        "meraki-info-linux", 
        "meraki-info-linux-arm",
        "meraki-info-mac",
        "meraki-info-mac-arm",
        "meraki-info"
    )
    
    foreach ($artifact in $artifacts) {
        if (Test-Path $artifact) {
            Remove-Item $artifact -Force
            Write-MakeOutput "  Removed: $artifact" -Color Warning
        }
    }
    
    # Run go clean
    go clean
    Write-MakeOutput "✅ Build artifacts cleaned" -Color Success
}

function Invoke-RunTarget {
    <#
    .SYNOPSIS
        Builds meraki-info and runs it with --help.

    .OUTPUTS
        The program's usage text. Exits 1 if no binary was produced.
    #>
    Write-MakeOutput "🚀 Building and running..." -Color Info
    Invoke-BuildTarget
    if (Test-Path "meraki-info.exe") {
        .\meraki-info.exe --help
    } elseif (Test-Path "meraki-info") {
        .\meraki-info --help
    } else {
        Write-MakeOutput "❌ No executable found to run" -Color Error
        exit 1
    }
}

function Invoke-AccessTarget {
    <#
    .SYNOPSIS
        Builds meraki-info and runs its access command.

    .DESCRIPTION
        Requires MERAKI_APIKEY in the environment; exits 1 without it.
        meraki-info reads the key from the environment, so it is not passed
        on the command line (where it would also have to precede the
        command). This makes a live call to the Meraki Dashboard API.

    .OUTPUTS
        The access report written by meraki-info.
    #>
    if (-not $env:MERAKI_APIKEY) {
        Write-MakeOutput "❌ Error: MERAKI_APIKEY environment variable is required" -Color Error
        Write-MakeOutput "Usage: `$env:MERAKI_APIKEY='your-key'; .\make.ps1 access" -Color Info
        exit 1
    }
    
    Write-MakeOutput "🔍 Showing access information..." -Color Info
    Invoke-BuildTarget
    
    if (Test-Path "meraki-info.exe") {
        .\meraki-info.exe access
    } elseif (Test-Path "meraki-info") {
        .\meraki-info access
    } else {
        Write-MakeOutput "❌ No executable found" -Color Error
        exit 1
    }
}

function Invoke-InstallTarget {
    <#
    .SYNOPSIS
        Runs `go install .` to install meraki-info into GOPATH/bin.

    .OUTPUTS
        None. Exits the script with go's exit code on failure.
    #>
    Write-MakeOutput "📦 Installing to GOPATH/bin..." -Color Info
    go install .
    if ($LASTEXITCODE -eq 0) {
        Write-MakeOutput "✅ Installation completed" -Color Success
    } else {
        Write-MakeOutput "❌ Installation failed" -Color Error
        exit $LASTEXITCODE
    }
}

function Invoke-DepsTarget {
    <#
    .SYNOPSIS
        Runs `go mod download` and `go mod tidy`.

    .DESCRIPTION
        Only the exit code of `go mod tidy` is checked.

    .OUTPUTS
        None. Exits the script with go's exit code on failure.
    #>
    Write-MakeOutput "📋 Managing dependencies..." -Color Info
    go mod download
    go mod tidy
    if ($LASTEXITCODE -eq 0) {
        Write-MakeOutput "✅ Dependencies updated" -Color Success
    } else {
        Write-MakeOutput "❌ Dependency management failed" -Color Error
        exit $LASTEXITCODE
    }
}

# Main execution
switch ($Target.ToLower()) {
    "help" { 
        Show-MakeHelp 
    }
    "build" { 
        Invoke-BuildTarget 
    }
    "build-windows" { 
        Invoke-BuildTarget "windows" 
    }
    "build-linux" { 
        Invoke-BuildTarget "linux" 
    }
    "build-linux-arm" { 
        Invoke-BuildTarget "linux-arm" 
    }
    "build-mac" { 
        Invoke-BuildTarget "mac" 
    }
    "build-mac-arm" { 
        Invoke-BuildTarget "mac-arm" 
    }
    "build-all" { 
        Invoke-BuildTarget "all" 
    }
    "test" { 
        Invoke-TestTarget 
    }
    "test-v" { 
        Invoke-TestTarget -Verbose 
    }
    "coverage" { 
        Invoke-TestTarget -Coverage 
    }
    "clean" { 
        Invoke-CleanTarget 
    }
    "run" { 
        Invoke-RunTarget 
    }
    "access" { 
        Invoke-AccessTarget 
    }
    "install" { 
        Invoke-InstallTarget 
    }
    "deps" { 
        Invoke-DepsTarget 
    }
    default {
        Write-MakeOutput "❌ Unknown target: $Target" -Color Error
        Write-MakeOutput "Run '.\make.ps1 help' to see available targets" -Color Info
        exit 1
    }
}

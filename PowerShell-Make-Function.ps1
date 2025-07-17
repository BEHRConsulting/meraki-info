# PowerShell Make Function
# Add this to your PowerShell profile to enable 'make' command globally
# 
# To add to profile:
# 1. Run: notepad $PROFILE
# 2. Add the content of this file to your profile
# 3. Restart PowerShell or run: . $PROFILE

function make {
    param(
        [Parameter(Position=0)]
        [string]$Target = "help",
        
        [Parameter(ValueFromRemainingArguments=$true)]
        [string[]]$Arguments = @()
    )
    
    # Check if we're in a directory with make.ps1
    if (Test-Path ".\make.ps1") {
        & .\make.ps1 $Target @Arguments
        return
    }
    
    # Check if we're in a directory with a Makefile and try to find make.exe
    if (Test-Path ".\Makefile") {
        # Try to find make in common locations
        $makePaths = @(
            "make",
            "C:\Program Files\Git\usr\bin\make.exe",
            "C:\msys64\usr\bin\make.exe",
            "C:\tools\msys64\usr\bin\make.exe",
            "${env:ProgramFiles}\Git\usr\bin\make.exe",
            "${env:ProgramFiles(x86)}\Git\usr\bin\make.exe"
        )
        
        foreach ($makePath in $makePaths) {
            if (Get-Command $makePath -ErrorAction SilentlyContinue) {
                & $makePath $Target @Arguments
                return
            }
        }
        
        Write-Warning "Found Makefile but no make executable. Consider using .\make.ps1 instead."
        Write-Host "Available: .\make.ps1 $Target" -ForegroundColor Yellow
        return
    }
    
    # Fallback message
    Write-Warning "No make.ps1 or Makefile found in current directory."
    Write-Host "For PowerShell make functionality, ensure make.ps1 is present." -ForegroundColor Yellow
}

# Export the function if this script is dot-sourced
Export-ModuleMember -Function make

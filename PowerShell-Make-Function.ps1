<#
.SYNOPSIS
    PowerShell Make Function.

.DESCRIPTION
    Add this to your PowerShell profile to enable 'make' command globally

    To add to profile:
    1. Run: notepad $PROFILE
    2. Add the content of this file to your profile
    3. Restart PowerShell or run: . $PROFILE

    Defines a global `make` function that runs .\make.ps1 when the current
    directory has one, otherwise a real make.exe when there is a Makefile.

.EXAMPLE
    . .\PowerShell-Make-Function.ps1
    make build

.INPUTS
    None.

.OUTPUTS
    None. Defines the make function in the current scope.

.NOTES
    The final Export-ModuleMember line only works when this file is loaded
    as a module; when the file is dot-sourced or pasted into a profile,
    PowerShell reports an error for that line, but the function is still
    defined.
#>


function make {
    <#
    .SYNOPSIS
        Runs a make target with make.ps1 or a real make.exe.

    .DESCRIPTION
        If .\make.ps1 exists in the current directory, runs it with the
        target and arguments. Otherwise, if a Makefile exists, runs the
        first make executable found on PATH or in common Git for Windows /
        MSYS2 locations. Otherwise prints a warning.

    .PARAMETER Target
        The make target to run. Defaults to help.

    .PARAMETER Arguments
        Any further arguments, passed through unchanged.

    .EXAMPLE
        make build-all

    .OUTPUTS
        Whatever the invoked make writes.
    #>
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

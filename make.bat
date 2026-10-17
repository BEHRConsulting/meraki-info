@echo off
REM PowerShell Make Wrapper Batch File
REM This allows you to run 'make' from Command Prompt and PowerShell
REM
REM Usage: make.bat <target> [arguments]
REM Passes everything to make.ps1 (next to this file) with the execution
REM policy bypassed for that one run. See make.ps1 for the targets.

REM Check if make.ps1 exists in current directory
if exist "%~dp0make.ps1" (
    powershell -ExecutionPolicy Bypass -File "%~dp0make.ps1" %*
) else (
    echo Error: make.ps1 not found in current directory
    echo Please ensure make.ps1 is present for PowerShell make functionality
    exit /b 1
)

@echo off
REM PowerShell Make Wrapper Batch File
REM This allows you to run 'make' from Command Prompt and PowerShell

REM Check if make.ps1 exists in current directory
if exist "%~dp0make.ps1" (
    powershell -ExecutionPolicy Bypass -File "%~dp0make.ps1" %*
) else (
    echo Error: make.ps1 not found in current directory
    echo Please ensure make.ps1 is present for PowerShell make functionality
    exit /b 1
)

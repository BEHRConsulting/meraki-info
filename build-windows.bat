@echo off
REM Meraki Info Windows Build Script
REM Batch file to build the Meraki Info application for Windows

setlocal EnableDelayedExpansion

REM Parse command line arguments
set "TARGET=windows"
set "BUILD_ALL=false"
set "CLEAN=false"
set "RUN_TESTS=false"
set "SHOW_HELP=false"

:parse_args
if "%1"=="" goto :main
if /i "%1"=="--help" set "SHOW_HELP=true"
if /i "%1"=="-h" set "SHOW_HELP=true"
if /i "%1"=="--all" set "BUILD_ALL=true"
if /i "%1"=="--clean" set "CLEAN=true"
if /i "%1"=="--test" set "RUN_TESTS=true"
if /i "%1"=="--target" (
    shift
    set "TARGET=%1"
)
shift
goto :parse_args

:main
if "%SHOW_HELP%"=="true" goto :show_help

echo.
echo 🚀 Starting Meraki Info build process...
echo.

REM Clean if requested
if "%CLEAN%"=="true" call :clean_artifacts

REM Run tests if requested
if "%RUN_TESTS%"=="true" call :run_tests

REM Build based on target
if "%BUILD_ALL%"=="true" (
    call :build_all
) else (
    call :build_single %TARGET%
)

echo.
echo 🎉 Build process completed successfully!
echo.

REM Show built files
echo 📦 Built files:
for %%f in (meraki-info*.exe meraki-info-linux meraki-info-mac) do (
    if exist "%%f" (
        for %%s in ("%%f") do (
            set /a "size_mb=%%~zs / 1048576"
            echo   %%f ^(!size_mb! MB^)
        )
    )
)

goto :end

:show_help
echo Meraki Info Windows Build Script
echo.
echo Usage: build-windows.bat [OPTIONS]
echo.
echo OPTIONS:
echo   --target ^<platform^>   Build for specific platform: windows, linux, mac ^(default: windows^)
echo   --all                 Build for all platforms
echo   --clean               Clean build artifacts before building
echo   --test                Run tests before building
echo   --help, -h            Show this help message
echo.
echo Examples:
echo   build-windows.bat                    # Build for Windows
echo   build-windows.bat --target linux    # Build for Linux
echo   build-windows.bat --all             # Build for all platforms
echo   build-windows.bat --clean --test    # Clean, test, then build
echo.
goto :end

:clean_artifacts
echo 🧹 Cleaning build artifacts...
if exist "meraki-info.exe" (
    del "meraki-info.exe"
    echo   Removed: meraki-info.exe
)
if exist "meraki-info-linux" (
    del "meraki-info-linux"
    echo   Removed: meraki-info-linux
)
if exist "meraki-info-mac" (
    del "meraki-info-mac"
    echo   Removed: meraki-info-mac
)
if exist "meraki-info" (
    del "meraki-info"
    echo   Removed: meraki-info
)
go clean
echo ✅ Build artifacts cleaned
echo.
goto :eof

:run_tests
echo 🧪 Running tests...
go test ./...
if errorlevel 1 (
    echo ❌ Tests failed!
    exit /b 1
)
echo ✅ All tests passed
echo.
goto :eof

:build_single
set "platform=%1"
if /i "%platform%"=="windows" (
    call :build_windows
    call :test_windows_exe
) else if /i "%platform%"=="linux" (
    call :build_linux
) else if /i "%platform%"=="mac" (
    call :build_mac
) else (
    echo ❌ Unknown target: %platform%
    echo Valid targets: windows, linux, mac
    exit /b 1
)
goto :eof

:build_all
echo Building for all platforms...
call :build_windows
call :build_linux
call :build_mac
call :test_windows_exe
goto :eof

:build_windows
echo 🔨 Building for Windows...
set GOOS=windows
set GOARCH=amd64
go build -o meraki-info.exe .
if errorlevel 1 (
    echo ❌ Windows build failed!
    exit /b 1
)
echo ✅ Windows build completed: meraki-info.exe
goto :eof

:build_linux
echo 🔨 Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -o meraki-info-linux .
if errorlevel 1 (
    echo ❌ Linux build failed!
    exit /b 1
)
echo ✅ Linux build completed: meraki-info-linux
goto :eof

:build_mac
echo 🔨 Building for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -o meraki-info-mac .
if errorlevel 1 (
    echo ❌ macOS build failed!
    exit /b 1
)
echo ✅ macOS build completed: meraki-info-mac
goto :eof

:test_windows_exe
if exist "meraki-info.exe" (
    echo 🧪 Testing Windows executable...
    meraki-info.exe --help >nul 2>&1
    if errorlevel 1 (
        echo ❌ Windows executable test failed!
    ) else (
        echo ✅ Windows executable test passed
    )
)
goto :eof

:end
endlocal

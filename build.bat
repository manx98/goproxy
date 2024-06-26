@echo off
chcp 65001 >nul
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to set codepage to UTF8
    exit /b %ERRORLEVEL%
)
echo 正在编译goproxy_windows_amd64...
set GOOS=windows
set GOARCH=amd64
go build -o dist/goproxy_windows_amd64.exe cmd/goproxy/main.go
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to build goproxy_windows_amd64
    exit /b %ERRORLEVEL%
)
echo 正在编译goproxy_linux_amd64...
set GOOS=linux
set GOARCH=amd64
go build -o dist/goproxy_linux_amd64 cmd/goproxy/main.go
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to build goproxy_linux_amd64
    exit /b %ERRORLEVEL%
)
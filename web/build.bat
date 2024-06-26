@echo off

chcp 65001 >nul
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to set codepage to UTF8
    exit /b %ERRORLEVEL%
)

echo 切换到 src 目录
cd src
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to change directory to src
    exit /b %ERRORLEVEL%
)

echo 安装pnpm
call npm install -g pnpm
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to install dependencies
    exit /b %ERRORLEVEL%
)

echo 安装依赖
call pnpm install
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to install dependencies
    exit /b %ERRORLEVEL%
)

set DIST_OUT_DIR=../dist

echo 编译项目
call pnpm run build --emptyOutDir
IF %ERRORLEVEL% NEQ 0 (
    echo Failed to build the project
    exit /b %ERRORLEVEL%
)

echo Build succeeded
exit /b 0

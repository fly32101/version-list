@echo off
REM Go版本管理器环境变量设置脚本（Windows）

chcp 936 >nul 2>&1  REM 切换到GBK编码（Windows默认中文编码）
setlocal enabledelayedexpansion

REM 设置环境变量
set "GOROOT=%USERPROFILE%\.go-version\current"
set "GOPATH=%USERPROFILE%\go"
set "GOBIN=%GOROOT%\bin"

echo 正在设置Go环境变量...
echo.

REM 检查是否以管理员权限运行
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo 警告：需要管理员权限才能永久设置系统环境变量
    echo 请右键以管理员身份运行此脚本
    echo.
    echo 或者手动设置以下环境变量：
    echo GOROOT=%GOROOT%
    echo GOPATH=%GOPATH%
    echo 并将 %GOBIN% 添加到PATH中
    pause
    exit /b 1
)

REM 设置系统环境变量
setx GOROOT "%GOROOT%"
setx GOPATH "%GOPATH%"

REM 获取当前PATH
for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v PATH 2^>nul') do (
    set "currentPath=%%B"
)

REM 检查是否已经包含GOBIN
echo %currentPath% | findstr /i "%GOBIN%" >nul
if %errorlevel% neq 0 (
    REM 添加GOBIN到PATH
    set "newPath=%currentPath%;%GOBIN%"
    setx PATH "%newPath%"
    echo PATH已更新，包含：%GOBIN%
) else (
    echo PATH已包含：%GOBIN%
)

echo.
echo 环境变量设置完成！
echo 请重新打开命令提示符或PowerShell窗口以使更改生效
pause

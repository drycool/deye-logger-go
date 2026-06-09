@echo off
chcp 65001 >nul
title Deye Logger Go (DEBUG)

echo =========================================
echo   Deye 6kW Logger (Go) — DEBUG MODE
echo =========================================
echo.

if not exist data mkdir data

echo Запуск с --debug и интервалом 10с...
echo.
.\deye-logger.exe --config config\deye.yml -i 10 --debug %*

if %ERRORLEVEL% neq 0 (
    echo.
    echo [!] Ошибка. Код: %ERRORLEVEL%
    pause
)

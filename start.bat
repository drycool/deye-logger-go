@echo off
chcp 65001 >nul
title Deye Logger Go

echo =========================================
echo   Deye 6kW Logger (Go)
echo =========================================
echo.

if not exist config\deye.yml (
    echo [!] Конфиг не найден: config\deye.yml
    echo [i] Копирую из deye.example.yml...
    copy config\deye.example.yml config\deye.yml >nul
    echo [✓] Создан config\deye.yml — проверь настройки!
    echo.
)

if not exist data mkdir data

echo Запуск: .\deye-logger.exe --config config\deye.yml -i 30
echo.
.\deye-logger.exe --config config\deye.yml -i 30 %*

if %ERRORLEVEL% neq 0 (
    echo.
    echo [!] Ошибка запуска. Код: %ERRORLEVEL%
    pause
)

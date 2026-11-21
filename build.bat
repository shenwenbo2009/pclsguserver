@echo off
setlocal

echo === Building frontend ===

echo --- Installing shared deps ---
cd frontend\shared
call npm install --silent
if errorlevel 1 exit /b 1

echo --- Building admin ---
cd ..\admin
call npm install --silent
if errorlevel 1 exit /b 1
call npm run build
if errorlevel 1 exit /b 1

echo --- Building market ---
cd ..\market
call npm install --silent
if errorlevel 1 exit /b 1
call npm run build
if errorlevel 1 exit /b 1

echo === Building Go binary ===
cd ..\..
go build -o pclsguserver.exe .\cmd\server
if errorlevel 1 exit /b 1

echo === Done ===
echo Binary: .\pclsguserver.exe
echo Run: .\pclsguserver.exe -config config.yaml
@echo off
setlocal

echo Wrestling - build

where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Go is not installed.
    exit /b 1
)

where npm >nul 2>nul
if %errorlevel% neq 0 (
    echo Node.js/npm is not installed.
    exit /b 1
)

if not exist server-go\bin mkdir server-go\bin
cd server-go
go test ./...
if %errorlevel% neq 0 exit /b 1
go vet ./...
if %errorlevel% neq 0 exit /b 1
go build -o bin\server.exe .\cmd\server
if %errorlevel% neq 0 exit /b 1
cd ..

if not exist node_modules call npm install
call npm run build
if %errorlevel% neq 0 exit /b 1

echo.
echo Build completed successfully.
endlocal

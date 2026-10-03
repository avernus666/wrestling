@echo off
setlocal
where go >nul 2>nul || (echo Go is not installed.& exit /b 1)
where npm >nul 2>nul || (echo Node.js/npm is not installed.& exit /b 1)
cd server-go
go test -race -coverprofile=coverage.out ./...
if %errorlevel% neq 0 exit /b 1
go vet ./...
if %errorlevel% neq 0 exit /b 1
cd ..
if not exist node_modules call npm install
call npm test -- --runInBand
if %errorlevel% neq 0 exit /b 1
call npm run build
if %errorlevel% neq 0 exit /b 1
echo All tests passed.
endlocal

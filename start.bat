@echo off
if "%~1"=="" (
    echo Please enter the port, e.g.: .\start.bat 8080
    exit /b 1
)

set APP_PORT=%1
docker-compose up --build -d
echo The application was launched on http://localhost:%APP_PORT%
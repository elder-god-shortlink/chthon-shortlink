# Script hoàn chỉnh để chạy tất cả Chthon ShortLink Services
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "   Chthon ShortLink - Full System" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Kiểm tra file .env
if (!(Test-Path ".env")) {
    Write-Host "X File .env not found!" -ForegroundColor Red
    Write-Host "Please copy .env.example to .env and configure your settings" -ForegroundColor Yellow
    exit 1
}

Write-Host "√ Found .env configuration file" -ForegroundColor Green

# Đọc cấu hình từ .env  
$envContent = Get-Content ".env"
$config = @{}
foreach ($line in $envContent) {
    if ($line -match "^([^#=]+)=(.*)$") {
        $config[$matches[1]] = $matches[2]
    }
}

$POSTGRES_HOST = $config["POSTGRES_HOST"]
$POSTGRES_PORT = $config["POSTGRES_PORT"] 
$REDIS_HOST = $config["REDIS_HOST"]
$REDIS_PORT = $config["REDIS_PORT"]

Write-Host ""
Write-Host "Configuration loaded:" -ForegroundColor Yellow
Write-Host "- PostgreSQL: ${POSTGRES_HOST}:${POSTGRES_PORT}" -ForegroundColor White
Write-Host "- Redis: ${REDIS_HOST}:${REDIS_PORT}" -ForegroundColor White
Write-Host ""

Write-Host "Testing infrastructure connections..." -ForegroundColor Yellow

# Test PostgreSQL
try {
    $connection = New-Object System.Net.Sockets.TcpClient
    $connection.Connect($POSTGRES_HOST, [int]$POSTGRES_PORT)
    Write-Host "√ PostgreSQL: SUCCESS" -ForegroundColor Green
    $connection.Close()
} catch {
    Write-Host "X PostgreSQL: FAILED - $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Please ensure PostgreSQL is running and database 'shortlink' exists" -ForegroundColor Yellow
    Write-Host "Run: .\create-database.ps1 for database setup instructions" -ForegroundColor Yellow
    exit 1
}

# Test Redis
try {
    $connection = New-Object System.Net.Sockets.TcpClient
    $connection.Connect($REDIS_HOST, [int]$REDIS_PORT)
    Write-Host "√ Redis: SUCCESS" -ForegroundColor Green
    $connection.Close()
} catch {
    Write-Host "X Redis: FAILED - $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Infrastructure checks passed! Starting services..." -ForegroundColor Green
Write-Host ""

# Kill existing processes
Write-Host "Stopping any existing services..." -ForegroundColor Yellow
Get-Process | Where-Object {$_.ProcessName -like "*gateway*" -or $_.ProcessName -like "*shortlink*" -or $_.ProcessName -like "*redirect*" -or $_.ProcessName -like "*analytics*" -or $_.ProcessName -like "*user-management*"} | Stop-Process -Force -ErrorAction SilentlyContinue

Start-Sleep -Seconds 2

# Kiểm tra binaries
$services = @(
    "bin/api-gateway.exe",
    "bin/shortlink-service.exe", 
    "bin/redirect-service.exe",
    "bin/analytics-service.exe",
    "bin/user-management-service.exe"
)

foreach ($service in $services) {
    if (!(Test-Path $service)) {
        Write-Host "X $service not found! Please run build first." -ForegroundColor Red
        exit 1
    }
}

Write-Host "√ All service binaries found" -ForegroundColor Green
Write-Host ""

# Start services với delay
Write-Host "Starting services in sequence..." -ForegroundColor Cyan
Write-Host ""

Write-Host "1. Starting API Gateway (port 8080)..." -ForegroundColor Yellow
Start-Process -FilePath ".\bin\api-gateway.exe" -WindowStyle Normal
Start-Sleep -Seconds 3

Write-Host "2. Starting Shortlink Service (port 8081)..." -ForegroundColor Yellow  
Start-Process -FilePath ".\bin\shortlink-service.exe" -WindowStyle Normal
Start-Sleep -Seconds 3

Write-Host "3. Starting Redirect Service (port 8082)..." -ForegroundColor Yellow
Start-Process -FilePath ".\bin\redirect-service.exe" -WindowStyle Normal
Start-Sleep -Seconds 3

Write-Host "4. Starting User Management Service (port 8084)..." -ForegroundColor Yellow
Start-Process -FilePath ".\bin\user-management-service.exe" -WindowStyle Normal
Start-Sleep -Seconds 3

Write-Host "5. Starting Analytics Service (port 8083)..." -ForegroundColor Yellow  
Start-Process -FilePath ".\bin\analytics-service.exe" -WindowStyle Normal
Start-Sleep -Seconds 3

Write-Host ""
Write-Host "==================================" -ForegroundColor Green
Write-Host "   All Services Started!" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Green
Write-Host ""

# Test health check
Write-Host "Testing system health..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 10
    Write-Host "√ Health Check: SUCCESS" -ForegroundColor Green
    Write-Host ""
    Write-Host "Service Status:" -ForegroundColor White
    foreach ($service in $response.services.PSObject.Properties) {
        $status = if ($service.Value) { "√ ONLINE" } else { "X OFFLINE" }
        $color = if ($service.Value) { "Green" } else { "Red" }
        Write-Host "- $($service.Name): $status" -ForegroundColor $color
    }
} catch {
    Write-Host "! Health Check: Some services may still be starting..." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Available Endpoints:" -ForegroundColor White
Write-Host "- API Gateway:    http://localhost:8080" -ForegroundColor Yellow
Write-Host "- Health Check:   http://localhost:8080/health" -ForegroundColor Yellow
Write-Host "- API Docs:       http://localhost:8080/docs/swagger" -ForegroundColor Yellow
Write-Host ""
Write-Host "Service Ports:" -ForegroundColor White
Write-Host "- API Gateway:         :8080" -ForegroundColor Yellow
Write-Host "- Shortlink Service:   :8081" -ForegroundColor Yellow  
Write-Host "- Redirect Service:    :8082" -ForegroundColor Yellow
Write-Host "- Analytics Service:   :8083" -ForegroundColor Yellow
Write-Host "- User Management:     :8084" -ForegroundColor Yellow
Write-Host ""
Write-Host "Infrastructure:" -ForegroundColor White
Write-Host "- PostgreSQL: ${POSTGRES_HOST}:${POSTGRES_PORT}" -ForegroundColor Yellow
Write-Host "- Redis: ${REDIS_HOST}:${REDIS_PORT}" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press Ctrl+C to stop monitoring..." -ForegroundColor Gray
Write-Host "Opening health check in browser..." -ForegroundColor Cyan

Start-Process "http://localhost:8080/health"

# Monitoring loop
while ($true) {
    Start-Sleep -Seconds 10
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 5 -ErrorAction Stop
        $onlineCount = ($response.services.PSObject.Properties | Where-Object { $_.Value }).Count
        $totalCount = $response.services.PSObject.Properties.Count
        Write-Host "$(Get-Date -Format 'HH:mm:ss') - Services: $onlineCount/$totalCount online" -ForegroundColor Green
    } catch {
        Write-Host "$(Get-Date -Format 'HH:mm:ss') - System health check failed" -ForegroundColor Red
    }
}

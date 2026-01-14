# demo.ps1 - NanoLambda Real World Demo

Write-Host "Starting NanoLambda Demo: Image Resizer" -ForegroundColor Cyan
Write-Host "----------------------------------------"

# 1. deploy the function
Write-Host "Step 1: Deploying 'image-resizer' function..." -ForegroundColor Yellow
.\nanolambda.exe deploy examples/image-resizer

if ($LASTEXITCODE -ne 0) {
    Write-Host "Deployment failed!" -ForegroundColor Red
    exit
}

Write-Host "`nFunction deployed! (Wait a moment for registration...)" -ForegroundColor Green
Start-Sleep -Seconds 2

# 2. first invocation (cold start)
Write-Host "`nStep 2: Triggering COLD START (This simulates the first user arriving)" -ForegroundColor Cyan
Write-Host "   Processing a Go Gopher image (100x100)..."

$timer = [System.Diagnostics.Stopwatch]::StartNew()

try {
    $response = Invoke-RestMethod -Method Post -Uri "http://localhost:8080/function/image-resizer" -Body '{"width": 100, "height": 100}' -ContentType "application/json"
    $timer.Stop()
    
    Write-Host "`n   Response Time: $($timer.Elapsed.TotalSeconds.ToString("0.00")) seconds" -ForegroundColor Magenta
    Write-Host "   Output:" -ForegroundColor White
    $response | Format-List
} catch {
    Write-Host "Invocation failed! Is the Gateway running?" -ForegroundColor Red
    Write-Host "   Run '.
gateway.exe' in another terminal first." -ForegroundColor Gray
    exit
}

# 3. second invocation (hot start)
Write-Host "`nStep 3: Triggering HOT START (Second user arrives immediately)" -ForegroundColor Cyan
$timer = [System.Diagnostics.Stopwatch]::StartNew()

try {
    $response = Invoke-RestMethod -Method Post -Uri "http://localhost:8080/function/image-resizer" -Body '{"width": 50, "height": 50}' -ContentType "application/json"
    $timer.Stop()
    
    Write-Host "`n   Response Time: $($timer.Elapsed.TotalSeconds.ToString("0.00")) seconds" -ForegroundColor Magenta
    Write-Host "   Output:" -ForegroundColor White
    $response | Format-List
} catch {
    Write-Host "Invocation failed!" -ForegroundColor Red
}

Write-Host "`n----------------------------------------"
Write-Host "Conclusion:" -ForegroundColor Yellow
Write-Host "   - The Cold Start was slower because it had to spin up a Docker container."
Write-Host "   - The Hot Start was instant because the container was ready."
Write-Host "   - In a real scenario, the AI Prophet (running in background) would predict"
Write-Host "     traffic spikes and ensure even the FIRST user gets a Hot Start!"
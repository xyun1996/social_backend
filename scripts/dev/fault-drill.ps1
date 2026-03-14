Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "Fault drill helper"
Write-Host "1. Confirm services and dependencies are up."
go run ./scripts/dev/cmd/check_local_durable_status
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "2. Manually restart Redis or MySQL now, then press Enter to continue."
Read-Host | Out-Null

Write-Host "3. Re-check durable status after recovery."
go run ./scripts/dev/cmd/check_local_durable_status

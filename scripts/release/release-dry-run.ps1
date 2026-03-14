Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host "Running release dry run..."
go test ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

powershell -ExecutionPolicy Bypass -File ./scripts/dev/dev-check.ps1
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

go run ./scripts/dev/cmd/check_contract_inventory
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Release dry run completed."
Write-Host "Next manual checks:"
Write-Host "  1. Verify migration order and rollback notes."
Write-Host "  2. Confirm APP_INTERNAL_TOKEN and OPS_API_TOKEN are set in target env."
Write-Host "  3. Link the release note to active runbooks and on-call contacts."

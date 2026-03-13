param(
    [switch]$RequireBuf,
    [switch]$RunLocalDurableGate
)

$ErrorActionPreference = "Stop"

Push-Location (Resolve-Path (Join-Path $PSScriptRoot "..\.."))
try {
    go test ./...

    $protoCheckArgs = @("-ExecutionPolicy", "Bypass", "-File", ".\scripts\dev\proto-check.ps1")
    if ($RequireBuf) {
        $protoCheckArgs += @("-RequireBuf", "-RunBufLint")
    }
    powershell @protoCheckArgs
    if ($LASTEXITCODE -ne 0) {
        throw "proto-check failed with exit code $LASTEXITCODE"
    }

    go run ./scripts/dev/cmd/check_contract_inventory
    if ($LASTEXITCODE -ne 0) {
        throw "check_contract_inventory failed with exit code $LASTEXITCODE"
    }

    if ($RunLocalDurableGate) {
        go run ./scripts/dev/cmd/check_local_durable_status
        if ($LASTEXITCODE -ne 0) {
            throw "check_local_durable_status failed with exit code $LASTEXITCODE"
        }
    }
}
finally {
    Pop-Location
}

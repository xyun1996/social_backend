param(
    [switch]$RequireBuf
)

$ErrorActionPreference = "Stop"

Push-Location (Resolve-Path (Join-Path $PSScriptRoot "..\.."))
try {
    go test ./api/proto

    $buf = Get-Command buf -ErrorAction SilentlyContinue
    if ($null -eq $buf) {
        if ($RequireBuf) {
            throw "buf is not installed or not on PATH"
        }
        Write-Host "buf is not installed or not on PATH; skipped buf lint"
        return
    }

    buf lint
}
finally {
    Pop-Location
}

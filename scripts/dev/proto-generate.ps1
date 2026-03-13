param(
    [switch]$LintOnly
)

$ErrorActionPreference = "Stop"

function Require-Buf {
    if (-not (Get-Command buf -ErrorAction SilentlyContinue)) {
        throw "buf is not installed or not on PATH"
    }
}

Push-Location (Resolve-Path (Join-Path $PSScriptRoot "..\.."))
try {
    Require-Buf

    if ($LintOnly) {
        buf lint
        return
    }

    New-Item -ItemType Directory -Force -Path ".\gen\proto\go" | Out-Null
    buf lint
    buf generate
}
finally {
    Pop-Location
}

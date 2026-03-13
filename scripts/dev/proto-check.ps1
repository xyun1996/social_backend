param(
    [switch]$RequireBuf,
    [switch]$RunBufLint
)

$ErrorActionPreference = "Stop"

function Resolve-Buf {
    $command = Get-Command buf -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }

    $gopath = (& go env GOPATH).Trim()
    if ($gopath) {
        $candidate = Join-Path $gopath "bin\buf.exe"
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    return $null
}

function Invoke-Checked {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Executable,
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$Arguments
    )

    & $Executable @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$Executable failed with exit code $LASTEXITCODE"
    }
}

Push-Location (Resolve-Path (Join-Path $PSScriptRoot "..\.."))
try {
    go test ./api/proto

    if ($RequireBuf) {
        $RunBufLint = $true
    }
    if (-not $RunBufLint) {
        Write-Host "buf lint skipped; use -RunBufLint or -RequireBuf for stricter proto checks"
        return
    }

    $buf = Resolve-Buf
    if ($null -eq $buf) {
        if ($RequireBuf) {
            throw "buf is not installed or not on PATH"
        }
        Write-Host "buf is not installed or not on PATH; skipped buf lint"
        return
    }

    Invoke-Checked $buf lint
}
finally {
    Pop-Location
}

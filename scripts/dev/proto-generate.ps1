param(
    [switch]$LintOnly,
    [switch]$SkipLint
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

    throw "buf is not installed or not on PATH"
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
    $buf = Resolve-Buf

    if ($LintOnly) {
        Invoke-Checked $buf lint
        return
    }

    New-Item -ItemType Directory -Force -Path ".\gen\proto\go" | Out-Null
    Get-ChildItem ".\gen\proto\go" -Force -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force
    New-Item -ItemType Directory -Force -Path ".\gen\proto\go" | Out-Null
    if (-not $SkipLint) {
        Invoke-Checked $buf lint
    }
    Invoke-Checked $buf generate
}
finally {
    Pop-Location
}

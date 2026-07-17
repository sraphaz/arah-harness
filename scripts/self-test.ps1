#Requires -Version 5.1
<#
.SYNOPSIS
  Teste de fumaça do ARAH Harness — roda em CI e localmente.
#>
$ErrorActionPreference = 'Stop'
$HarnessRoot = Split-Path $PSScriptRoot -Parent
$Cli = Join-Path $HarnessRoot 'cli/arah.ps1'
$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("arah-selftest-" + [guid]::NewGuid().ToString('n').Substring(0, 8))

Write-Host "ARAH self-test → $Tmp"
New-Item -ItemType Directory -Path $Tmp -Force | Out-Null
git -C $Tmp init -q

# Minimal app surface so discover has something to observe
New-Item -ItemType Directory -Path (Join-Path $Tmp 'backend') -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $Tmp 'frontend') -Force | Out-Null
Set-Content -Path (Join-Path $Tmp 'backend/main.go') -Value 'package main' -Encoding UTF8
Set-Content -Path (Join-Path $Tmp 'go.mod') -Value "module selftest`n`ngo 1.22" -Encoding UTF8

try {
    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli init -Target $Tmp -ProjectName selftest -Force
    if ($LASTEXITCODE -ne 0) { throw "init failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli domain sync -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "domain sync failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli discover -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "discover failed" }
    if (-not (Test-Path (Join-Path $Tmp 'docs/_meta/discovery.proposed.yaml'))) {
        throw "discovery.proposed.yaml missing"
    }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli organism bootstrap -Target $Tmp -Force
    if ($LASTEXITCODE -ne 0) { throw "organism bootstrap failed" }
    if (-not (Test-Path (Join-Path $Tmp 'docs/_meta/organism.manifest.yaml'))) {
        throw "organism.manifest.yaml missing"
    }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli organism signal -Target $Tmp `
        -From orchestrator -To backend -Type attract -Topic delivery
    if ($LASTEXITCODE -ne 0) { throw "organism signal failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli evolve -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "evolve failed" }
    if (-not (Test-Path (Join-Path $Tmp 'docs/_meta/evolution.proposed.yaml'))) {
        throw "evolution.proposed.yaml missing"
    }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli export-graph -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "export-graph failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $Tmp 'scripts/harness/validate-agent-graph.ps1')
    if ($LASTEXITCODE -ne 0) { throw "validate-agent-graph failed" }

    # regenerate without re-update (kernel already installed); skip doctor harness path noise
    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli regenerate -Target $Tmp -SkipDoctor -Force
    if ($LASTEXITCODE -ne 0) { throw "regenerate failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli doctor -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "doctor failed" }

    Write-Host "self-test: OK"
    exit 0
} finally {
    Remove-Item $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}

#Requires -Version 5.1
<#
.SYNOPSIS
  Teste de fumaça do ARAH Harness — roda em CI e localmente.
#>
$ErrorActionPreference = 'Stop'
$HarnessRoot = Split-Path $PSScriptRoot -Parent
$Cli = Join-Path $HarnessRoot 'cli'
$Cli = Join-Path $Cli 'arah.ps1'
$PwshExe = if (Get-Command pwsh -ErrorAction SilentlyContinue) { 'pwsh' } else { 'powershell' }
$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("arah-selftest-" + [guid]::NewGuid().ToString('n').Substring(0, 8))

function Join-RepoPath {
    param([string]$Base, [string[]]$Parts)
    $path = $Base
    foreach ($p in $Parts) { $path = Join-Path $path $p }
    return $path
}

Write-Host "ARAH self-test → $Tmp"
New-Item -ItemType Directory -Path $Tmp -Force | Out-Null
git -C $Tmp init -q

# Minimal app surface so discover has something to observe
New-Item -ItemType Directory -Path (Join-Path $Tmp 'backend') -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $Tmp 'frontend') -Force | Out-Null
Set-Content -Path (Join-Path (Join-Path $Tmp 'backend') 'main.go') -Value 'package main' -Encoding UTF8
Set-Content -Path (Join-Path $Tmp 'go.mod') -Value "module selftest`n`ngo 1.22" -Encoding UTF8

try {
    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli init -Target $Tmp -ProjectName selftest -Force
    if ($LASTEXITCODE -ne 0) { throw "init failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli domain sync -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "domain sync failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli discover -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "discover failed" }
    $discovery = Join-RepoPath $Tmp @('docs', '_meta', 'discovery.proposed.yaml')
    if (-not (Test-Path -LiteralPath $discovery)) {
        throw "discovery.proposed.yaml missing"
    }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli organism bootstrap -Target $Tmp -Force
    if ($LASTEXITCODE -ne 0) { throw "organism bootstrap failed" }
    $organism = Join-RepoPath $Tmp @('docs', '_meta', 'organism.manifest.yaml')
    if (-not (Test-Path -LiteralPath $organism)) {
        throw "organism.manifest.yaml missing"
    }

    $signalBus = Join-RepoPath $Tmp @('scripts', 'agents', 'signal-bus.ps1')
    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $signalBus `
        -From orchestrator -SignalTo backend -SignalType attract -Topic delivery
    if ($LASTEXITCODE -ne 0) { throw "organism signal failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli evolve -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "evolve failed" }
    $evolution = Join-RepoPath $Tmp @('docs', '_meta', 'evolution.proposed.yaml')
    if (-not (Test-Path -LiteralPath $evolution)) {
        throw "evolution.proposed.yaml missing"
    }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli metrics rollup -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "metrics rollup failed" }
    $summary = Join-RepoPath $Tmp @('.arah', 'observability', 'summary.yaml')
    if (-not (Test-Path -LiteralPath $summary)) {
        throw "metrics summary.yaml missing"
    }
    $summaryRaw = Get-Content -LiteralPath $summary -Raw
    if ($summaryRaw -notmatch 'schema:\s*arah-harness/metrics-summary') {
        throw "metrics summary missing schema arah-harness/metrics-summary"
    }
    if ($summaryRaw -notmatch 'semaphore:') {
        throw "metrics summary missing semaphore"
    }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli metrics report -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "metrics report failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli export-graph -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "export-graph failed" }

    $validateGraph = Join-RepoPath $Tmp @('scripts', 'harness', 'validate-agent-graph.ps1')
    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $validateGraph -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "validate-agent-graph failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli regenerate -Target $Tmp -SkipDoctor -Force
    if ($LASTEXITCODE -ne 0) { throw "regenerate failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli doctor -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "doctor failed" }

    Write-Host "self-test: OK"
    exit 0
} finally {
    Remove-Item $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}

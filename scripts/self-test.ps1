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
        -From orchestrator -SignalTo backend -SignalType attract -Topic delivery `
        -Payload '{"note":"ok","api_key":"SUPERSECRETVALUE12345"}'
    if ($LASTEXITCODE -ne 0) { throw "organism signal failed" }

    $pendingDir = Join-RepoPath $Tmp @('.arah', 'local', 'bus', 'pending')
    $pendingFiles = @(Get-ChildItem -LiteralPath $pendingDir -Filter '*.json' -ErrorAction SilentlyContinue)
    if ($pendingFiles.Count -lt 1) { throw "signal pending file missing under .arah/local/bus/pending" }
    $sigRaw = Get-Content -LiteralPath $pendingFiles[0].FullName -Raw
    if ($sigRaw -notmatch '"v"\s*:') { throw "signal missing wire field v" }
    if ($sigRaw -match 'SUPERSECRETVALUE12345') { throw "secret scrubbing failed — raw secret on disk" }
    if ($sigRaw -notmatch 'REDACTED') { throw "secret scrubbing failed — expected REDACTED marker" }

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

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli compact -Target $Tmp -Kind bus
    if ($LASTEXITCODE -ne 0) { throw "compact failed" }
    $archiveDir = Join-RepoPath $Tmp @('.arah', 'local', 'bus', 'archive')
    $archives = @(Get-ChildItem -LiteralPath $archiveDir -Filter '*.jsonl' -ErrorAction SilentlyContinue)
    if ($archives.Count -lt 1) { throw "compact did not produce archive jsonl" }
    $pendingAfter = @(Get-ChildItem -LiteralPath $pendingDir -Filter '*.json' -ErrorAction SilentlyContinue)
    if ($pendingAfter.Count -ne 0) { throw "compact left pending files behind" }

    # Seed legacy jsonl and migrate
    $legacyBus = Join-RepoPath $Tmp @('.arah', 'bus')
    New-Item -ItemType Directory -Path $legacyBus -Force | Out-Null
    Set-Content -Path (Join-Path $legacyBus 'signals.jsonl') -Value '{"ts":"2026-01-01T00:00:00Z","type":"status","from":"legacy"}' -Encoding UTF8
    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli migrate-state -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "migrate-state failed" }
    $runsRoot = Join-RepoPath $Tmp @('docs', '_meta', 'runs')
    $runSummaries = @(Get-ChildItem -LiteralPath $runsRoot -Recurse -Filter 'summary.json' -ErrorAction SilentlyContinue)
    if ($runSummaries.Count -lt 1) { throw "migrate-state did not write cold summary" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli export-graph -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "export-graph failed" }

    $validateGraph = Join-RepoPath $Tmp @('scripts', 'harness', 'validate-agent-graph.ps1')
    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $validateGraph -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "validate-agent-graph failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli regenerate -Target $Tmp -SkipDoctor -Force
    if ($LASTEXITCODE -ne 0) { throw "regenerate failed" }

    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli doctor -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "doctor failed" }

    # Minimal install smoke
    $TmpMin = Join-Path ([System.IO.Path]::GetTempPath()) ("arah-selftest-min-" + [guid]::NewGuid().ToString('n').Substring(0, 8))
    New-Item -ItemType Directory -Path $TmpMin -Force | Out-Null
    git -C $TmpMin init -q
    try {
        & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli install -Target $TmpMin -ProjectName minitest -Force -Minimal
        if ($LASTEXITCODE -ne 0) { throw "install -Minimal failed" }
        $cfg = Get-Content (Join-Path $TmpMin 'arah.config.yaml') -Raw
        if ($cfg -notmatch 'enabled:\s*false') { throw "minimal install missing organism.enabled=false" }
        $gi = Get-Content (Join-Path $TmpMin '.gitignore') -Raw
        if ($gi -notmatch '\.arah/local/') { throw "minimal install missing .arah/local/ gitignore" }
        & $PwshExe -NoProfile -ExecutionPolicy Bypass -File $Cli hooks install -Target $TmpMin -Force
        if ($LASTEXITCODE -ne 0) { throw "hooks install failed" }
        if (-not (Test-Path (Join-Path (Join-Path $TmpMin '.git') (Join-Path 'hooks' 'pre-commit')))) {
            throw "pre-commit hook missing"
        }
    } finally {
        Remove-Item $TmpMin -Recurse -Force -ErrorAction SilentlyContinue
    }

    # capabilities.yaml present in harness
    $caps = Join-Path $HarnessRoot 'capabilities.yaml'
    if (-not (Test-Path -LiteralPath $caps)) { throw "capabilities.yaml missing" }

    # Execution Control Protocol — distribution + scenarios
    Write-Host "=== Execution Control ==="
    if (-not (Test-Path (Join-Path $Tmp 'scripts/agents/execute-task.ps1'))) {
        throw "execute-task.ps1 missing after init"
    }
    if (-not (Test-Path (Join-Path $Tmp '.cursor/rules/arah-execution-control.mdc'))) {
        throw "cursor execution-control rule missing after init"
    }
    if (-not (Test-Path (Join-Path $Tmp 'schemas/arah-harness/execution-contract.schema.yaml'))) {
        throw "execution-contract schema missing after init"
    }
    $cfgEcp = Get-Content (Join-Path $Tmp 'arah.config.yaml') -Raw
    if ($cfgEcp -notmatch '(?m)^execution_control:') { throw "execution_control missing in consumer config" }
    & $PwshExe -NoProfile -ExecutionPolicy Bypass -File (Join-Path $HarnessRoot 'scripts/harness/test-execution-control.ps1')
    if ($LASTEXITCODE -ne 0) { throw "test-execution-control failed" }

    Write-Host "self-test: OK"
    exit 0
} finally {
    Remove-Item $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}

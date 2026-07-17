#Requires -Version 5.1
<#
.SYNOPSIS
  Homeostase do biocomponente — regenera o harness no repositório consumidor.
.DESCRIPTION
  Pipeline unificado:
    1) update kernel (opcional, via -UpdateKernel / harness path)
    2) domain sync
    3) discover
    4) organism bootstrap
    5) evolve
    6) export-graph
    7) doctor
  Sugestões ficam em docs/_meta/*.proposed.yaml para o repositório evoluir.
.EXAMPLE
  ./regenerate-harness.ps1
  ./regenerate-harness.ps1 -UpdateKernel -HarnessRoot C:\arah-harness -Force
#>
param(
    [string]$HarnessRoot = '',
    [switch]$UpdateKernel,
    [switch]$Force,
    [switch]$ApplyDiscovery,
    [switch]$SkipDoctor,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'
$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$Agents = $PSScriptRoot

Write-Host "regenerate: organism homeostasis → $Root"

function Invoke-Step {
    param([string]$Name, [scriptblock]$Block)
    Write-Host ""
    Write-Host "==> $Name"
    if ($DryRun) {
        Write-Host "[dry-run] $Name"
        return
    }
    & $Block
    if ($LASTEXITCODE -and $LASTEXITCODE -ne 0) {
        throw "regenerate step failed: $Name (exit $LASTEXITCODE)"
    }
}

# 1) Kernel update (optional — requires harness clone)
if ($UpdateKernel) {
    Invoke-Step 'update kernel' {
        if (-not $HarnessRoot) {
            if ($env:ARAH_HARNESS_PATH) { $HarnessRoot = $env:ARAH_HARNESS_PATH }
        }
        if (-not $HarnessRoot) {
            Write-Warning 'UpdateKernel skipped: set -HarnessRoot or ARAH_HARNESS_PATH'
            return
        }
        $cli = Join-Path $HarnessRoot 'cli/arah.ps1'
        if (-not (Test-Path $cli)) { throw "Harness CLI not found: $cli" }
        & powershell -NoProfile -ExecutionPolicy Bypass -File $cli update -Target $Root -Force:$Force
    }
}

# 2) Domain sync
Invoke-Step 'domain sync' {
    $script = Join-Path $Agents 'domain-sync.ps1'
    if (-not (Test-Path $script)) { throw 'domain-sync.ps1 missing — run arah init/update first' }
    & powershell -NoProfile -ExecutionPolicy Bypass -File $script @($(if ($DryRun) { '-DryRun' }))
}

# 3) Discover
Invoke-Step 'discover' {
    $script = Join-Path $Agents 'discover-repo.ps1'
    $args = @()
    if ($ApplyDiscovery) { $args += '-Apply' }
    if ($DryRun) { $args += '-DryRun' }
    & powershell -NoProfile -ExecutionPolicy Bypass -File $script @args
}

# 4) Organism bootstrap
Invoke-Step 'organism bootstrap' {
    $script = Join-Path $Agents 'organism-bootstrap.ps1'
    $args = @()
    if ($Force) { $args += '-Force' }
    if ($DryRun) { $args += '-DryRun' }
    & powershell -NoProfile -ExecutionPolicy Bypass -File $script @args
}

# 5) Evolve
Invoke-Step 'evolve' {
    $script = Join-Path $Agents 'evolve-harness.ps1'
    & powershell -NoProfile -ExecutionPolicy Bypass -File $script @($(if ($DryRun) { '-DryRun' }))
}

# 6) Export graph
Invoke-Step 'export-graph' {
    $script = Join-Path $Agents 'export-agent-graph.ps1'
    if (-not (Test-Path $script)) {
        Write-Warning 'export-agent-graph.ps1 missing — skip'
        return
    }
    Push-Location $Root
    try { & powershell -NoProfile -ExecutionPolicy Bypass -File $script } finally { Pop-Location }
}

# 7) Doctor
if (-not $SkipDoctor) {
    Invoke-Step 'doctor' {
        $validate = Join-Path $Agents 'validate-manifests.ps1'
        if (Test-Path $validate) {
            & powershell -NoProfile -ExecutionPolicy Bypass -File $validate
        }
        if ($HarnessRoot -or $env:ARAH_HARNESS_PATH) {
            $hr = if ($HarnessRoot) { $HarnessRoot } else { $env:ARAH_HARNESS_PATH }
            $doctor = Join-Path $hr 'cli/doctor.ps1'
            if (Test-Path $doctor) {
                & powershell -NoProfile -ExecutionPolicy Bypass -File $doctor -Target $Root
            }
        }
    }
}

Write-Host ""
Write-Host 'regenerate: complete'
Write-Host '  artifacts:'
Write-Host '    docs/_meta/discovery.proposed.yaml'
Write-Host '    docs/_meta/organism.manifest.yaml'
Write-Host '    docs/_meta/evolution.proposed.yaml'
Write-Host '    docs/_meta/agent-graph.generated.json'
Write-Host '  review proposals → PR → human merge'

$record = Join-Path $Agents 'record-agent-event.ps1'
if ((Test-Path $record) -and -not $DryRun) {
    & $record -AgentId orchestrator -Action 'regenerate.homeostasis' -Outcome ok -AutonomyLevel activate -Details 'pipeline=ok' 2>$null
}

exit 0

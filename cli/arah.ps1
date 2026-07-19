#Requires -Version 5.1
<#
.SYNOPSIS
  ARAH Harness CLI — init, update, doctor, discover, organism, evolve, regenerate, compact, migrate-state, hooks
#>
param(
    [Parameter(Position = 0)]
    [ValidateSet(
        'init', 'install', 'update', 'doctor', 'sync-check', 'domain',
        'export-graph', 'validate-runtime', 'discover', 'organism',
        'evolve', 'regenerate', 'compact', 'migrate-state', 'hooks', 'help'
    )]
    [string]$Command = 'help',

    [Parameter(Position = 1)]
    [ValidateSet('sync', 'bootstrap', 'status', 'signal', 'install', '')]
    [string]$SubCommand = '',

    [string]$Target = '',
    [string]$ProjectName = '',
    [switch]$Force,
    [switch]$DryRun,
    [switch]$Apply,
    [switch]$UpdateKernel,
    [switch]$ApplyDiscovery,
    [switch]$SkipDoctor,
    [switch]$Minimal,
    [ValidateSet('all', 'bus', 'audit', '')]
    [string]$Kind = '',
    [int]$RetainDays = 90,

    # signal-bus passthrough (SignalTo avoids -To/-Topic prefix ambiguity)
    [string]$From = '',
    [string]$SignalTo = '*',
    [ValidateSet('attract', 'consult', 'propose', 'acknowledge', 'coalesce', 'evolve', 'status', '')]
    [string]$SignalType = '',
    [string]$Topic = 'general',
    [string]$Payload = ''
)

$ErrorActionPreference = 'Stop'
$CliDir = $PSScriptRoot
$HarnessRoot = Split-Path $CliDir -Parent
$targetPath = if ($Target) { $Target } else { (Get-Location).Path }
if (Test-Path -LiteralPath $targetPath) {
    $targetPath = (Resolve-Path -LiteralPath $targetPath).Path
}

function Get-TargetScript {
    param([string]$Rel)
    $parts = $Rel -split '[\\/]+'
    $path = $targetPath
    foreach ($p in $parts) {
        if ($p) { $path = Join-Path $path $p }
    }
    if (-not (Test-Path -LiteralPath $path)) {
        # Fall back to harness root scripts (for compact/migrate before/without full install)
        $fallback = $HarnessRoot
        foreach ($p in $parts) {
            if ($p) { $fallback = Join-Path $fallback $p }
        }
        if (Test-Path -LiteralPath $fallback) { return $fallback }
        Write-Error "$Rel not found — run arah init/update first"
        exit 1
    }
    return $path
}

function Invoke-TargetScript {
    param(
        [string]$ScriptPath,
        [object[]]$ScriptArgs = @()
    )
    Push-Location $targetPath
    try {
        & $ScriptPath @ScriptArgs
        if (-not $?) { exit 1 }
        if ($null -ne $LASTEXITCODE -and $LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    } finally {
        Pop-Location
    }
}

switch ($Command) {
    'init' {
        $splat = @{ Target = $targetPath; Force = $Force; Minimal = $Minimal }
        if ($ProjectName) { $splat.ProjectName = $ProjectName }
        & (Join-Path $CliDir 'init.ps1') @splat
    }
    'install' {
        $splat = @{ Target = $targetPath; Force = $Force; Minimal = $Minimal }
        if ($ProjectName) { $splat.ProjectName = $ProjectName }
        & (Join-Path $CliDir 'install.ps1') @splat
    }
    'update' {
        & (Join-Path $CliDir 'update.ps1') -Target $targetPath -Force:$Force
    }
    'doctor' {
        & (Join-Path $CliDir 'doctor.ps1') -Target $targetPath
    }
    'sync-check' {
        & (Join-Path $CliDir 'sync-check.ps1') -Target $targetPath
    }
    'domain' {
        if ($SubCommand -ne 'sync') {
            Write-Error "Use: arah domain sync"
            exit 1
        }
        $script = Get-TargetScript 'scripts/agents/domain-sync.ps1'
        $invokeArgs = @()
        if ($DryRun) { $invokeArgs += '-DryRun' }
        Invoke-TargetScript -ScriptPath $script -ScriptArgs $invokeArgs
    }
    'export-graph' {
        $script = Get-TargetScript 'scripts/agents/export-agent-graph.ps1'
        Invoke-TargetScript -ScriptPath $script
    }
    'validate-runtime' {
        $script = Get-TargetScript 'scripts/harness/validate-solution-choreography.ps1'
        Invoke-TargetScript -ScriptPath $script
    }
    'discover' {
        $script = Get-TargetScript 'scripts/agents/discover-repo.ps1'
        $invokeArgs = @()
        if ($Apply) { $invokeArgs += '-Apply' }
        if ($DryRun) { $invokeArgs += '-DryRun' }
        Invoke-TargetScript -ScriptPath $script -ScriptArgs $invokeArgs
    }
    'organism' {
        switch ($SubCommand) {
            'bootstrap' {
                $script = Get-TargetScript 'scripts/agents/organism-bootstrap.ps1'
                $invokeArgs = @()
                if ($Force) { $invokeArgs += '-Force' }
                if ($DryRun) { $invokeArgs += '-DryRun' }
                Invoke-TargetScript -ScriptPath $script -ScriptArgs $invokeArgs
            }
            'status' {
                $manifest = Join-Path (Join-Path $targetPath 'docs') (Join-Path '_meta' 'organism.manifest.yaml')
                $state = Join-Path (Join-Path $targetPath '.arah') (Join-Path 'organism' 'state.json')
                if (Test-Path -LiteralPath $state) { Get-Content -LiteralPath $state }
                elseif (Test-Path -LiteralPath $manifest) {
                    Write-Host "organism: manifest present → $manifest"
                    Get-Content -LiteralPath $manifest | Select-Object -First 40
                }
                else {
                    Write-Host 'organism: not bootstrapped — run arah organism bootstrap'
                    exit 1
                }
            }
            'signal' {
                $script = Get-TargetScript 'scripts/agents/signal-bus.ps1'
                if (-not $From -or -not $SignalType) {
                    Write-Error 'Use: arah organism signal -From <cell> -SignalType <attract|consult|propose|...> [-SignalTo ...] [-Topic ...] [-Payload json]'
                    exit 1
                }
                $signalArgs = @('-From', $From, '-SignalTo', $SignalTo, '-SignalType', $SignalType, '-Topic', $Topic)
                if ($Payload) { $signalArgs += @('-Payload', $Payload) }
                Invoke-TargetScript -ScriptPath $script -ScriptArgs $signalArgs
            }
            default {
                Write-Error 'Use: arah organism bootstrap|status|signal'
                exit 1
            }
        }
    }
    'evolve' {
        $script = Get-TargetScript 'scripts/agents/evolve-harness.ps1'
        $invokeArgs = @()
        if ($Apply) { $invokeArgs += '-Apply' }
        if ($DryRun) { $invokeArgs += '-DryRun' }
        Invoke-TargetScript -ScriptPath $script -ScriptArgs $invokeArgs
    }
    'regenerate' {
        & (Join-Path $CliDir 'regenerate.ps1') -Target $targetPath -Force:$Force `
            -UpdateKernel:$UpdateKernel -ApplyDiscovery:$ApplyDiscovery `
            -SkipDoctor:$SkipDoctor -DryRun:$DryRun
    }
    'compact' {
        $script = Get-TargetScript 'scripts/agents/compact-state.ps1'
        $k = if ($Kind) { $Kind } else { 'all' }
        $invokeArgs = @('-Kind', $k, '-RetainDays', $RetainDays)
        if ($DryRun) { $invokeArgs += '-DryRun' }
        Invoke-TargetScript -ScriptPath $script -ScriptArgs $invokeArgs
    }
    'migrate-state' {
        $script = Get-TargetScript 'scripts/agents/migrate-state.ps1'
        $invokeArgs = @()
        if ($DryRun) { $invokeArgs += '-DryRun' }
        Invoke-TargetScript -ScriptPath $script -ScriptArgs $invokeArgs
    }
    'hooks' {
        if ($SubCommand -ne 'install') {
            Write-Error 'Use: arah hooks install [-Target path] [-Force]'
            exit 1
        }
        $script = Join-Path (Join-Path (Join-Path $HarnessRoot 'scripts') 'agents') 'install-hooks.ps1'
        if (-not (Test-Path -LiteralPath $script)) {
            $script = Get-TargetScript 'scripts/agents/install-hooks.ps1'
        }
        $invokeArgs = @('-Target', $targetPath)
        if ($Force) { $invokeArgs += '-Force' }
        & $script @invokeArgs
        if ($null -ne $LASTEXITCODE -and $LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    }
    default {
        Write-Host @"
ARAH Harness CLI — TechOrganism

  powershell -File cli/arah.ps1 install [-Target path] [-ProjectName name] [-Force] [-Minimal]
  powershell -File cli/arah.ps1 init [-Target path] [-ProjectName name] [-Force] [-Minimal]
  powershell -File cli/arah.ps1 update [-Target path] [-Force]
  powershell -File cli/arah.ps1 doctor [-Target path]
  powershell -File cli/arah.ps1 sync-check [-Target path]
  powershell -File cli/arah.ps1 domain sync [-Target path] [-DryRun]
  powershell -File cli/arah.ps1 export-graph [-Target path]
  powershell -File cli/arah.ps1 validate-runtime [-Target path]

  # TechOrganism
  powershell -File cli/arah.ps1 discover [-Target path] [-Apply] [-DryRun]
  powershell -File cli/arah.ps1 organism bootstrap [-Target path] [-Force]
  powershell -File cli/arah.ps1 organism status [-Target path]
  powershell -File cli/arah.ps1 organism signal -From cell -SignalType attract|consult|propose|... [-SignalTo ...] [-Topic ...]
  powershell -File cli/arah.ps1 evolve [-Target path] [-Apply] [-DryRun]
  powershell -File cli/arah.ps1 regenerate [-Target path] [-UpdateKernel] [-Force] [-ApplyDiscovery]

  # State model (hot/cold)
  powershell -File cli/arah.ps1 compact [-Target path] [-Kind all|bus|audit] [-RetainDays 90] [-DryRun]
  powershell -File cli/arah.ps1 migrate-state [-Target path] [-DryRun]
  powershell -File cli/arah.ps1 hooks install [-Target path] [-Force]
"@
    }
}

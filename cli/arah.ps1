#Requires -Version 5.1
<#
.SYNOPSIS
  ARAH Harness CLI — init, update, doctor, discover, organism, evolve, regenerate
#>
param(
    [Parameter(Position = 0)]
    [ValidateSet(
        'init', 'install', 'update', 'doctor', 'sync-check', 'domain',
        'export-graph', 'validate-runtime', 'discover', 'organism',
        'evolve', 'regenerate', 'help'
    )]
    [string]$Command = 'help',

    [Parameter(Position = 1)]
    [ValidateSet('sync', 'bootstrap', 'status', 'signal', '')]
    [string]$SubCommand = '',

    [string]$Target = '',
    [string]$ProjectName = '',
    [switch]$Force,
    [switch]$DryRun,
    [switch]$Apply,
    [switch]$UpdateKernel,
    [switch]$ApplyDiscovery,
    [switch]$SkipDoctor,

    # signal-bus passthrough
    [string]$From = '',
    [string]$To = '*',
    [ValidateSet('attract', 'consult', 'propose', 'acknowledge', 'coalesce', 'evolve', 'status', '')]
    [string]$Type = '',
    [string]$Topic = 'general',
    [string]$Payload = ''
)

$CliDir = $PSScriptRoot
$HarnessRoot = Split-Path $CliDir -Parent
$targetPath = if ($Target) { $Target } else { (Get-Location).Path }

function Get-TargetScript {
    param([string]$Rel)
    $path = Join-Path $targetPath $Rel
    if (-not (Test-Path $path)) {
        Write-Error "$Rel not found — run arah init/update first"
        exit 1
    }
    return $path
}

switch ($Command) {
    'init' {
        $args = @{ Target = $targetPath; Force = $Force }
        if ($ProjectName) { $args.ProjectName = $ProjectName }
        & (Join-Path $CliDir 'init.ps1') @args
    }
    'install' {
        $args = @{ Target = $targetPath; Force = $Force }
        if ($ProjectName) { $args.ProjectName = $ProjectName }
        & (Join-Path $CliDir 'install.ps1') @args
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
        & $script @($(if ($DryRun) { '-DryRun' }))
    }
    'export-graph' {
        $script = Get-TargetScript 'scripts/agents/export-agent-graph.ps1'
        Push-Location $targetPath
        try { & $script } finally { Pop-Location }
    }
    'validate-runtime' {
        $script = Get-TargetScript 'scripts/harness/validate-solution-choreography.ps1'
        Push-Location $targetPath
        try { & $script } finally { Pop-Location }
    }
    'discover' {
        $script = Get-TargetScript 'scripts/agents/discover-repo.ps1'
        $args = @()
        if ($Apply) { $args += '-Apply' }
        if ($DryRun) { $args += '-DryRun' }
        Push-Location $targetPath
        try { & $script @args } finally { Pop-Location }
    }
    'organism' {
        switch ($SubCommand) {
            'bootstrap' {
                $script = Get-TargetScript 'scripts/agents/organism-bootstrap.ps1'
                $args = @()
                if ($Force) { $args += '-Force' }
                if ($DryRun) { $args += '-DryRun' }
                Push-Location $targetPath
                try { & $script @args } finally { Pop-Location }
            }
            'status' {
                $manifest = Join-Path $targetPath 'docs/_meta/organism.manifest.yaml'
                $state = Join-Path $targetPath '.arah/organism/state.json'
                if (Test-Path $state) { Get-Content $state }
                elseif (Test-Path $manifest) {
                    Write-Host "organism: manifest present → $manifest"
                    Get-Content $manifest | Select-Object -First 40
                }
                else {
                    Write-Host 'organism: not bootstrapped — run arah organism bootstrap'
                    exit 1
                }
            }
            'signal' {
                $script = Get-TargetScript 'scripts/agents/signal-bus.ps1'
                if (-not $From -or -not $Type) {
                    Write-Error 'Use: arah organism signal -From <cell> -Type <attract|consult|propose|...> [-To ...] [-Topic ...] [-Payload json]'
                    exit 1
                }
                Push-Location $targetPath
                try {
                    & $script -From $From -To $To -Type $Type -Topic $Topic -Payload $Payload
                } finally { Pop-Location }
            }
            default {
                Write-Error 'Use: arah organism bootstrap|status|signal'
                exit 1
            }
        }
    }
    'evolve' {
        $script = Get-TargetScript 'scripts/agents/evolve-harness.ps1'
        $args = @()
        if ($Apply) { $args += '-Apply' }
        if ($DryRun) { $args += '-DryRun' }
        Push-Location $targetPath
        try { & $script @args } finally { Pop-Location }
    }
    'regenerate' {
        & (Join-Path $CliDir 'regenerate.ps1') -Target $targetPath -Force:$Force `
            -UpdateKernel:$UpdateKernel -ApplyDiscovery:$ApplyDiscovery `
            -SkipDoctor:$SkipDoctor -DryRun:$DryRun
    }
    default {
        Write-Host @"
ARAH Harness CLI — biocomponente

  powershell -File cli/arah.ps1 install [-Target path] [-ProjectName name] [-Force]
  powershell -File cli/arah.ps1 init [-Target path] [-ProjectName name] [-Force]
  powershell -File cli/arah.ps1 update [-Target path] [-Force]
  powershell -File cli/arah.ps1 doctor [-Target path]
  powershell -File cli/arah.ps1 sync-check [-Target path]
  powershell -File cli/arah.ps1 domain sync [-Target path] [-DryRun]
  powershell -File cli/arah.ps1 export-graph [-Target path]
  powershell -File cli/arah.ps1 validate-runtime [-Target path]

  # Biocomponente
  powershell -File cli/arah.ps1 discover [-Target path] [-Apply] [-DryRun]
  powershell -File cli/arah.ps1 organism bootstrap [-Target path] [-Force]
  powershell -File cli/arah.ps1 organism status [-Target path]
  powershell -File cli/arah.ps1 organism signal -From cell -Type attract|consult|propose|... [-To ...] [-Topic ...]
  powershell -File cli/arah.ps1 evolve [-Target path] [-Apply] [-DryRun]
  powershell -File cli/arah.ps1 regenerate [-Target path] [-UpdateKernel] [-Force] [-ApplyDiscovery]
"@
    }
}

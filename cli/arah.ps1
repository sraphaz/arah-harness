#Requires -Version 5.1
<#
.SYNOPSIS
  ARAH Harness CLI — init, update, doctor, sync-check, domain sync, export-graph
#>
param(
    [Parameter(Position = 0)]
    [ValidateSet('init', 'install', 'update', 'doctor', 'sync-check', 'domain', 'export-graph', 'validate-runtime', 'help')]
    [string]$Command = 'help',
    [Parameter(Position = 1)]
    [ValidateSet('sync')]
    [string]$SubCommand = '',
    [string]$Target = '',
    [string]$ProjectName = '',
    [switch]$Force,
    [switch]$DryRun
)

$CliDir = $PSScriptRoot
$HarnessRoot = Split-Path $CliDir -Parent
$targetPath = if ($Target) { $Target } else { (Get-Location).Path }

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
        $script = Join-Path $targetPath 'scripts/agents/domain-sync.ps1'
        if (-not (Test-Path $script)) {
            Write-Error "domain-sync not found — run arah init first"
            exit 1
        }
        & $script @($(if ($DryRun) { '-DryRun' }))
    }
    'export-graph' {
        $script = Join-Path $targetPath 'scripts/agents/export-agent-graph.ps1'
        if (-not (Test-Path $script)) {
            Write-Error "export-agent-graph not found — run arah init first"
            exit 1
        }
        Push-Location $targetPath
        try { & $script } finally { Pop-Location }
    }
    'validate-runtime' {
        $script = Join-Path $targetPath 'scripts/harness/validate-solution-choreography.ps1'
        if (-not (Test-Path $script)) {
            Write-Error "validate-solution-choreography not found — run arah update"
            exit 1
        }
        Push-Location $targetPath
        try { & $script } finally { Pop-Location }
    }
    default {
        Write-Host @"
ARAH Harness CLI

  powershell -File cli/arah.ps1 install [-Target path] [-ProjectName name] [-Force]
  powershell -File cli/arah.ps1 init [-Target path] [-ProjectName name] [-Force]
  powershell -File cli/arah.ps1 update [-Target path] [-Force]
  powershell -File cli/arah.ps1 doctor [-Target path]
  powershell -File cli/arah.ps1 sync-check [-Target path]
  powershell -File cli/arah.ps1 domain sync [-Target path] [-DryRun]
  powershell -File cli/arah.ps1 export-graph [-Target path]
  powershell -File cli/arah.ps1 validate-runtime [-Target path]
"@
    }
}

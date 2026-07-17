#Requires -Version 5.1
<#
.SYNOPSIS
  CLI wrapper: regenera biocomponente no Target (update opcional + pipeline local).
#>
param(
    [string]$Target = (Get-Location).Path,
    [switch]$Force,
    [switch]$UpdateKernel,
    [switch]$ApplyDiscovery,
    [switch]$SkipDoctor,
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'
$HarnessRoot = Split-Path $PSScriptRoot -Parent
$Target = (Resolve-Path $Target).Path

$script = Join-Path $Target 'scripts/agents/regenerate-harness.ps1'
if (-not (Test-Path $script)) {
    # Bootstrap: ensure kernel scripts exist first
    Write-Host 'regenerate: local script missing — applying kernel update first'
    & (Join-Path $PSScriptRoot 'update.ps1') -Target $Target -Force:$Force
    $script = Join-Path $Target 'scripts/agents/regenerate-harness.ps1'
    if (-not (Test-Path $script)) {
        Write-Error 'regenerate-harness.ps1 still missing after update — harness too old?'
        exit 1
    }
}

$args = @{
    HarnessRoot = $HarnessRoot
    UpdateKernel = $UpdateKernel
    Force = $Force
    ApplyDiscovery = $ApplyDiscovery
    SkipDoctor = $SkipDoctor
    DryRun = $DryRun
}
& $script @args

#Requires -Version 5.1
param(
    [string]$Target = (Get-Location).Path,
    [switch]$Force
)
$ErrorActionPreference = 'Stop'
& (Join-Path $PSScriptRoot 'init.ps1') -Target $Target -Force:$Force
Write-Host "update: re-applied kernel (customize via arah.config.yaml and .agents/domain/)"

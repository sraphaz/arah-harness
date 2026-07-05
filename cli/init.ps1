#Requires -Version 5.1
<#
.SYNOPSIS
  Instala o kernel ARAH em um repositório-alvo.
.PARAMETER Target
  Caminho do repositório (default: diretório atual).
.PARAMETER ProjectName
  Nome do projeto para AGENTS.md e arah.config.yaml.
.PARAMETER Force
  Sobrescreve arquivos do kernel existentes.
#>
param(
    [string]$Target = (Get-Location).Path,
    [string]$ProjectName = '',
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$HarnessRoot = Split-Path $PSScriptRoot -Parent
$KernelRoot = Join-Path $HarnessRoot 'kernel'
$TemplatesRoot = Join-Path $HarnessRoot 'templates'
$Version = '0.2.0'

if (-not (Test-Path $KernelRoot)) {
    Write-Error "Kernel not found at $KernelRoot"
    exit 1
}

$Target = (Resolve-Path $Target).Path
if (-not $ProjectName) {
    $ProjectName = Split-Path $Target -Leaf
}

Write-Host "ARAH init → $Target (project: $ProjectName)"

function Copy-Tree {
    param([string]$From, [string]$To, [switch]$Overwrite)
    if (-not (Test-Path $From)) { return }
    Get-ChildItem $From -Recurse -File | ForEach-Object {
        $rel = $_.FullName.Substring($From.Length).TrimStart('\', '/')
        $dest = Join-Path $To $rel
        $destDir = Split-Path $dest -Parent
        if (-not (Test-Path $destDir)) { New-Item -ItemType Directory -Path $destDir -Force | Out-Null }
        if ((Test-Path $dest) -and -not $Overwrite) {
            Write-Host "  skip (exists): $rel"
        } else {
            Copy-Item $_.FullName $dest -Force
            Write-Host "  installed: $rel"
        }
    }
}

# Kernel → target repo root
Copy-Tree -From (Join-Path $KernelRoot '.agents') -To (Join-Path $Target '.agents') -Overwrite:$Force
Copy-Tree -From (Join-Path $KernelRoot '.skills') -To (Join-Path $Target '.skills') -Overwrite:$Force
Copy-Tree -From (Join-Path $KernelRoot '.cursor') -To (Join-Path $Target '.cursor') -Overwrite:$Force
Copy-Tree -From (Join-Path $KernelRoot 'scripts') -To (Join-Path $Target 'scripts') -Overwrite:$Force

# Templates
$configTpl = Join-Path $TemplatesRoot 'arah.config.yaml'
$configDest = Join-Path $Target 'arah.config.yaml'
if (-not (Test-Path $configDest) -or $Force) {
    $cfg = Get-Content $configTpl -Raw
    $cfg = $cfg -replace '\{\{PROJECT_NAME\}\}', $ProjectName
    Set-Content -Path $configDest -Value $cfg -Encoding UTF8
    Write-Host "  installed: arah.config.yaml"
}

$agentsTpl = Join-Path $TemplatesRoot 'AGENTS.md.tpl'
$agentsDest = Join-Path $Target 'AGENTS.md'
if (-not (Test-Path $agentsDest) -or $Force) {
    $md = Get-Content $agentsTpl -Raw
    $md = $md -replace '\{\{PROJECT_NAME\}\}', $ProjectName
    $md = $md -replace '\{\{HARNESS_VERSION\}\}', $Version
    Set-Content -Path $agentsDest -Value $md -Encoding UTF8
    Write-Host "  installed: AGENTS.md"
}

$specTpl = Join-Path $TemplatesRoot 'docs\specs\_template.spec.yaml'
$specDest = Join-Path $Target 'docs\specs\_template.spec.yaml'
if (-not (Test-Path (Split-Path $specDest -Parent))) {
    New-Item -ItemType Directory -Path (Split-Path $specDest -Parent) -Force | Out-Null
}
if (-not (Test-Path $specDest) -or $Force) {
    Copy-Item $specTpl $specDest -Force
    Write-Host "  installed: docs/specs/_template.spec.yaml"
}

# docs/_meta for agent graph
$metaDir = Join-Path $Target 'docs/_meta'
if (-not (Test-Path $metaDir)) {
    New-Item -ItemType Directory -Path $metaDir -Force | Out-Null
    Write-Host "  installed: docs/_meta/"
}

# Domain/specialist dirs
foreach ($sub in @('domain', 'specialists')) {
    $d = Join-Path $Target ".agents/$sub"
    if (-not (Test-Path $d)) {
        New-Item -ItemType Directory -Path $d -Force | Out-Null
        Write-Host "  installed: .agents/$sub/"
    }
}

# GitHub workflow
$wfTpl = Join-Path $TemplatesRoot 'github/workflows/agents-validate.yml'
$wfDest = Join-Path $Target '.github/workflows/agents-validate.yml'
if ((Test-Path $wfTpl) -and ((-not (Test-Path $wfDest)) -or $Force)) {
    $wfDir = Split-Path $wfDest -Parent
    if (-not (Test-Path $wfDir)) { New-Item -ItemType Directory -Path $wfDir -Force | Out-Null }
    Copy-Item $wfTpl $wfDest -Force
    Write-Host "  installed: .github/workflows/agents-validate.yml"
}

# Pin file
$pin = @"
harness: arah-harness
version: $Version
installed: $(Get-Date -Format 'yyyy-MM-dd')
"@
Set-Content -Path (Join-Path $Target '.arah-version') -Value $pin -Encoding UTF8

Write-Host ""
Write-Host "ARAH init complete. Next:"
Write-Host "  1. Edit arah.config.yaml (tests, domains)"
Write-Host "  2. powershell -File path/to/arah-harness/cli/arah.ps1 domain sync"
Write-Host "  3. ./scripts/agents/validate-manifests.ps1 && arah export-graph"

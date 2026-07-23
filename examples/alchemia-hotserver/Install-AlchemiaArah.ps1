#Requires -Version 5.1
<#
.SYNOPSIS
  Instala ARAH Harness + overlay Alchemia HotServer no repositório do jogo.

.EXAMPLE
  powershell -ExecutionPolicy Bypass -File .\Install-AlchemiaArah.ps1 -Target "D:\SERVIDOR NO D"
#>
param(
  [string]$Target = "",
  [string]$HarnessPath = "",
  [switch]$SkipHarnessClone,
  [switch]$SkipDoctor
)

$ErrorActionPreference = "Stop"
$PackRoot = $PSScriptRoot
$Overlay = Join-Path $PackRoot "overlay"

function Write-Step([string]$msg) {
  Write-Host ""
  Write-Host "==> $msg" -ForegroundColor Cyan
}

function Ensure-Dir([string]$path) {
  if (-not (Test-Path -LiteralPath $path)) {
    New-Item -ItemType Directory -Path $path -Force | Out-Null
  }
}

function Copy-OverlayTree([string]$src, [string]$dst) {
  if (-not (Test-Path -LiteralPath $src)) {
    throw "Overlay incompleto: $src"
  }
  Ensure-Dir $dst
  Get-ChildItem -LiteralPath $src -Force | ForEach-Object {
    $destItem = Join-Path $dst $_.Name
    if ($_.PSIsContainer) {
      Ensure-Dir $destItem
      Copy-OverlayTree $_.FullName $destItem
    } else {
      Copy-Item -LiteralPath $_.FullName -Destination $destItem -Force
    }
  }
}

Write-Host "============================================" -ForegroundColor Green
Write-Host " Alchemia HotServer — ARAH Pack Installer" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green

if ([string]::IsNullOrWhiteSpace($Target)) {
  $Target = Read-Host "Caminho do repositorio do jogo (ex: D:\SERVIDOR NO D)"
}
$Target = $Target.Trim().Trim('"')
if (-not (Test-Path -LiteralPath $Target)) {
  throw "Target nao existe: $Target"
}

Write-Step "Target = $Target"

# --- Harness ---
if ([string]::IsNullOrWhiteSpace($HarnessPath)) {
  if ($env:ARAH_HARNESS_PATH -and (Test-Path -LiteralPath $env:ARAH_HARNESS_PATH)) {
    $HarnessPath = $env:ARAH_HARNESS_PATH
  } else {
    $HarnessPath = "D:\arah-harness"
  }
}

if (-not (Test-Path -LiteralPath (Join-Path $HarnessPath "cli\arah.ps1"))) {
  if ($SkipHarnessClone) {
    throw "Harness nao encontrado em $HarnessPath e -SkipHarnessClone foi passado."
  }
  Write-Step "Clonando arah-harness em $HarnessPath"
  $parent = Split-Path -Parent $HarnessPath
  Ensure-Dir $parent
  if (Test-Path -LiteralPath $HarnessPath) {
    throw "Pasta $HarnessPath existe mas nao contem cli\arah.ps1"
  }
  git clone https://github.com/sraphaz/arah-harness.git $HarnessPath
} else {
  Write-Step "Harness encontrado em $HarnessPath"
}

$env:ARAH_HARNESS_PATH = $HarnessPath
$ArahCli = Join-Path $HarnessPath "cli\arah.ps1"

# --- arah install (brownfield) ---
Write-Step "arah install (kernel + templates; nao sobrescreve AGENTS.md existente)"
& powershell -ExecutionPolicy Bypass -File $ArahCli install -Target $Target -ProjectName "alchemia-hotserver"

# --- Overlay Alchemia ---
Write-Step "Aplicando overlay Alchemia"
Copy-OverlayTree $Overlay $Target

# --- AGENTS.md merge ---
Write-Step "Mesclando AGENTS.md"
$AgentsTarget = Join-Path $Target "AGENTS.md"
$AgentsPack = Join-Path $Overlay "AGENTS.md"
$CodexBackup = Join-Path $Target "coisas do codex\backups"
$Stamp = Get-Date -Format "yyyyMMdd-HHmmss"

if (Test-Path -LiteralPath $AgentsTarget) {
  $existing = Get-Content -LiteralPath $AgentsTarget -Raw -Encoding UTF8
  if ($existing -notmatch "ARAH Harness") {
    if (Test-Path -LiteralPath (Join-Path $Target "coisas do codex")) {
      Ensure-Dir $CodexBackup
      Copy-Item -LiteralPath $AgentsTarget -Destination (Join-Path $CodexBackup "AGENTS.md.$Stamp.bak") -Force
      Write-Host "Backup AGENTS.md -> coisas do codex\backups\AGENTS.md.$Stamp.bak"
    } else {
      Copy-Item -LiteralPath $AgentsTarget -Destination "$AgentsTarget.$Stamp.bak" -Force
      Write-Host "Backup AGENTS.md -> AGENTS.md.$Stamp.bak"
    }
    # Pack AGENTS.md already contains Alchemia rules + ARAH section
    Copy-Item -LiteralPath $AgentsPack -Destination $AgentsTarget -Force
  } else {
    Write-Host "AGENTS.md ja contem secao ARAH — preservando arquivo atual e gravando referencia em AGENTS.ARAH.REFERENCE.md"
    Copy-Item -LiteralPath $AgentsPack -Destination (Join-Path $Target "AGENTS.ARAH.REFERENCE.md") -Force
  }
} else {
  Copy-Item -LiteralPath $AgentsPack -Destination $AgentsTarget -Force
}

# --- domain sync / validate / doctor ---
Write-Step "domain sync"
try {
  & powershell -ExecutionPolicy Bypass -File $ArahCli domain sync -Target $Target
} catch {
  Write-Warning "domain sync falhou (overlay de domains ja vem no pack): $_"
}

Write-Step "validate-manifests"
$Validate = Join-Path $Target "scripts\agents\validate-manifests.ps1"
if (Test-Path -LiteralPath $Validate) {
  try {
    & powershell -ExecutionPolicy Bypass -File $Validate
  } catch {
    Write-Warning "validate-manifests: $_"
  }
} else {
  Write-Warning "validate-manifests.ps1 nao encontrado — rode apos arah install completo"
}

if (-not $SkipDoctor) {
  Write-Step "doctor"
  try {
    & powershell -ExecutionPolicy Bypass -File $ArahCli doctor -Target $Target
  } catch {
    Write-Warning "doctor: $_"
  }
}

try {
  & powershell -ExecutionPolicy Bypass -File $ArahCli export-graph -Target $Target
} catch {
  Write-Warning "export-graph: $_"
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Green
Write-Host " Instalacao concluida" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Green
Write-Host "Proximo: abra CHECKLIST.md e valide no Cursor."
Write-Host "Target:  $Target"
Write-Host "Harness: $HarnessPath"
Write-Host ""
Write-Host "Skills Alchemia:"
Write-Host "  ./scripts/agents/invoke-skill.ps1 -Skill lua-validate"
Write-Host "  ./scripts/agents/invoke-skill.ps1 -Skill add-spell"
Write-Host "  ./scripts/agents/invoke-skill.ps1 -Skill balance-pass"
Write-Host "  ./scripts/agents/invoke-skill.ps1 -Skill power-beast-touch"

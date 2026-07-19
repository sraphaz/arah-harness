#Requires -Version 5.1
<#
.SYNOPSIS
  Instala ARAH Harness em um repositório-alvo (init + doctor + próximos passos).
.EXAMPLE
  cd C:\path\to\meu-projeto
  powershell -File $env:USERPROFILE\arah-harness\cli\install.ps1 -ProjectName meu-projeto
#>
param(
    [string]$Target = (Get-Location).Path,
    [string]$ProjectName = '',
    [switch]$Force,
    [switch]$Minimal
)

$ErrorActionPreference = 'Stop'
$CliDir = $PSScriptRoot
$HarnessRoot = Split-Path $CliDir -Parent

if (-not $ProjectName) {
    $ProjectName = Split-Path (Resolve-Path $Target).Path -Leaf
}

Write-Host ""
Write-Host "=== ARAH Harness install ==="
Write-Host "Harness: $HarnessRoot"
Write-Host "Target:  $Target"
Write-Host "Project: $ProjectName"
Write-Host ""

& (Join-Path $CliDir 'init.ps1') -Target $Target -ProjectName $ProjectName -Force:$Force -Minimal:$Minimal

Write-Host ""
Write-Host "=== Validating installation ==="
& (Join-Path $CliDir 'doctor.ps1') -Target $Target
if ($LASTEXITCODE -ne 0) {
    Write-Error 'doctor failed after init'
    exit 1
}

Write-Host ""
Write-Host "=== Next steps ==="
if ($Minimal) {
    Write-Host "  Mode: MINIMAL (manifests + gates; organism deferred)"
    Write-Host "  1. Edit arah.config.yaml (tests, domains[])"
    Write-Host "  2. powershell -File `"$HarnessRoot\cli\arah.ps1`" domain sync -Target `"$Target`""
    Write-Host "  3. powershell -File `"$HarnessRoot\cli\arah.ps1`" hooks install -Target `"$Target`""
    Write-Host "  4. Upgrade to full TechOrganism:"
    Write-Host "       powershell -File `"$HarnessRoot\cli\arah.ps1`" regenerate -Target `"$Target`" -UpdateKernel"
    Write-Host "       powershell -File `"$HarnessRoot\cli\arah.ps1`" discover -Target `"$Target`""
    Write-Host "       powershell -File `"$HarnessRoot\cli\arah.ps1`" organism bootstrap -Target `"$Target`""
} else {
    Write-Host "  1. Edit arah.config.yaml (tests, domains[])"
    Write-Host "  2. powershell -File `"$HarnessRoot\cli\arah.ps1`" domain sync -Target `"$Target`""
    Write-Host "  3. Optional: create .agents/choreography.<project>.yaml for path overlays"
    Write-Host "  4. powershell -File `"$HarnessRoot\cli\arah.ps1`" export-graph -Target `"$Target`""
    Write-Host "  5. git add .agents .skills scripts .cursor arah.config.yaml .arah-version .github"
}
Write-Host ""
Write-Host "Docs: $HarnessRoot\docs\INSTALL.md"
Write-Host ""

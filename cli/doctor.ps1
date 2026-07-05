#Requires -Version 5.1
param([string]$Target = (Get-Location).Path)
$ErrorActionPreference = 'Stop'
$Target = (Resolve-Path $Target).Path

Write-Host "ARAH doctor — $Target"
$ok = $true

foreach ($path in @('AGENTS.md', 'arah.config.yaml', '.agents/choreography.yaml', '.skills', 'scripts/agents/validate-manifests.ps1')) {
    $full = Join-Path $Target $path
    if (Test-Path $full) {
        Write-Host "  [ok] $path"
    } else {
        Write-Host "  [MISSING] $path"
        $ok = $false
    }
}

if ($ok) {
    Push-Location $Target
    try {
        & (Join-Path $Target 'scripts/agents/validate-manifests.ps1')
        if ($LASTEXITCODE -ne 0 -and $null -ne $LASTEXITCODE) { $ok = $false }
    } catch {
        $ok = $false
    } finally {
        Pop-Location
    }
}

if ($ok) { Write-Host "doctor: OK"; exit 0 }
Write-Host "doctor: FAIL — run arah init"
exit 1

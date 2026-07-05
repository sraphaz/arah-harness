#Requires -Version 5.1
param([string]$Target = (Get-Location).Path)
$ErrorActionPreference = 'Stop'
$HarnessRoot = Split-Path $PSScriptRoot -Parent
$Target = (Resolve-Path $Target).Path

$drift = @()
$kernelFiles = Get-ChildItem (Join-Path $HarnessRoot 'kernel') -Recurse -File
foreach ($kf in $kernelFiles) {
    $rel = $kf.FullName.Substring((Join-Path $HarnessRoot 'kernel').Length).TrimStart('\', '/')
    $dest = Join-Path $Target $rel
    if (-not (Test-Path $dest)) {
        $drift += "missing: $rel"
        continue
    }
    $h = (Get-FileHash $kf.FullName).Hash
    $t = (Get-FileHash $dest).Hash
    if ($h -ne $t) { $drift += "modified: $rel" }
}

if ($drift.Count -eq 0) {
    Write-Host "sync-check: OK — kernel in sync"
    exit 0
}

Write-Host "sync-check: DRIFT detected ($($drift.Count) files)"
$drift | Select-Object -First 20 | ForEach-Object { Write-Host "  $_" }
if ($drift.Count -gt 20) { Write-Host "  ... and $($drift.Count - 20) more" }
Write-Host "Run: arah update -Force to re-apply kernel"
exit 1

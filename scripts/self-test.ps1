#Requires -Version 5.1
<#
.SYNOPSIS
  Teste de fumaça do ARAH Harness — roda em CI e localmente.
#>
$ErrorActionPreference = 'Stop'
$HarnessRoot = Split-Path $PSScriptRoot -Parent
$Cli = Join-Path $HarnessRoot 'cli\arah.ps1'
$Tmp = Join-Path $env:TEMP ("arah-selftest-" + [guid]::NewGuid().ToString('n').Substring(0, 8))

Write-Host "ARAH self-test → $Tmp"
New-Item -ItemType Directory -Path $Tmp -Force | Out-Null
git -C $Tmp init -q

try {
    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli init -Target $Tmp -ProjectName selftest -Force
    if ($LASTEXITCODE -ne 0) { throw "init failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli domain sync -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "domain sync failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli export-graph -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "export-graph failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $Tmp 'scripts\harness\validate-agent-graph.ps1')
    if ($LASTEXITCODE -ne 0) { throw "validate-agent-graph failed" }

    & powershell -NoProfile -ExecutionPolicy Bypass -File $Cli doctor -Target $Tmp
    if ($LASTEXITCODE -ne 0) { throw "doctor failed" }

    Write-Host "self-test: OK"
    exit 0
} finally {
    Remove-Item $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}

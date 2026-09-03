#Requires -Version 5.1
param(
  [switch]$Start
)

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$Bat = Join-Path $Root "client_run\build_otclient_release.bat"

Write-Host "=== build-client (Alchemia) ==="
Write-Host "1) tasklist /FO CSV | findstr /I `"cmake ninja cl.exe link.exe msbuild devenv otclient`""
Write-Host "2) Nao use Ninja em PowerShell cru (crypt32.lib / SDK)."
Write-Host "3) Script correto: client_run\build_otclient_release.bat"
Write-Host "4) Background + log; avisar a cada 5 min."
Write-Host "5) Apos OK: copiar exe do usuario COM backup; PDBs/logs -> coisas do codex"

if (-not (Test-Path -LiteralPath $Bat)) {
  Write-Warning "Bat nao encontrado: $Bat"
  exit 0
}

if (-not $Start) {
  Write-Host ""
  Write-Host "Dry-run. Para iniciar de verdade:"
  Write-Host "  powershell -File .\scripts\agents\alchemia\build-client.ps1 -Start"
  exit 0
}

$LogDir = Join-Path $Root "coisas do codex\build-logs"
if (-not (Test-Path -LiteralPath (Split-Path $LogDir))) {
  $LogDir = Join-Path $Root "docs\arah\build-logs"
}
New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
$Log = Join-Path $LogDir ("otclient-build-" + (Get-Date -Format "yyyyMMdd-HHmmss") + ".log")
Write-Host "Iniciando build em background. Log: $Log"
Start-Process -FilePath "cmd.exe" -ArgumentList "/c `"$Bat`" > `"$Log`" 2>&1" -WorkingDirectory (Split-Path $Bat) -WindowStyle Minimized
Write-Host "Acompanhe o log. Nao feche processos de build sem motivo."
exit 0

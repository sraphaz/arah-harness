#Requires -Version 5.1
param(
  [string[]]$Paths = @()
)

$ErrorActionPreference = "Stop"

$needsCpp = $false
$needsReload = $false

if ($Paths.Count -eq 0) {
  Write-Host "Uso: informe -Paths com arquivos alterados (relativos ou absolutos)."
  Write-Host "Exemplo: -Paths 'canary-3.4.1/data/scripts/spells/x.lua','client_run/modules/.../y.lua'"
}

foreach ($p in $Paths) {
  $n = $p.ToLowerInvariant()
  if ($n -match '\\src\\|\/src\/|/cmake|\.cpp$|\.h$|\.hpp$|cmakelists') {
    $needsCpp = $true
  }
  if ($n -match '\.lua$|\.otui$|\.xml$|\.otml$') {
    $needsReload = $true
  }
}

Write-Host "=== hot-reload-check ==="
if ($needsCpp) {
  Write-Host "BUILD C++ necessario. Use client_run\build_otclient_release.bat (ou pipeline de server)." -ForegroundColor Yellow
  Write-Host "Nao chame Ninja em PowerShell cru. Build longo em BACKGROUND."
} elseif ($needsReload) {
  Write-Host "Sem build C++. Basta RELOAD/RESTART do processo (servidor e/ou cliente)." -ForegroundColor Green
} else {
  Write-Host "Nao foi possivel classificar. Revise paths e AGENTS.md." -ForegroundColor DarkYellow
}
exit 0

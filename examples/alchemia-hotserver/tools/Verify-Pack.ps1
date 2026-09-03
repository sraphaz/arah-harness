#Requires -Version 5.1
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$required = @(
  "LEIA-ME.md",
  "CHECKLIST.md",
  "Install-AlchemiaArah.ps1",
  "overlay\arah.config.yaml",
  "overlay\AGENTS.md",
  "overlay\.agents\choreography.alchemia.yaml",
  "overlay\.skills\add-spell.skill.yaml",
  "overlay\scripts\agents\alchemia\lua-validate.ps1",
  "overlay\docs\arah\ALCHEMIA_ARAH.md"
)
$missing = @()
foreach ($r in $required) {
  $p = Join-Path $Root $r
  if (-not (Test-Path -LiteralPath $p)) { $missing += $r }
}
$domains = @(Get-ChildItem (Join-Path $Root "overlay\.agents\domain\*.agent.yaml") -ErrorAction SilentlyContinue)
$skills = @(Get-ChildItem (Join-Path $Root "overlay\.skills\*.skill.yaml") -ErrorAction SilentlyContinue)
Write-Host "domains=$($domains.Count) skills=$($skills.Count)"
if ($missing.Count -gt 0) {
  Write-Host "MISSING:" -ForegroundColor Red
  $missing | ForEach-Object { Write-Host " - $_" }
  exit 1
}
if ($domains.Count -lt 8) { Write-Error "Expected >=8 domain agents"; exit 1 }
if ($skills.Count -lt 8) { Write-Error "Expected >=8 skills"; exit 1 }
Write-Host "Pack OK" -ForegroundColor Green
exit 0

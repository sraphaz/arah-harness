#Requires -Version 5.1
param(
  [ValidateSet("server", "client", "all")]
  [string]$Area = "all",
  [string[]]$Files = @()
)

$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$LuaJIT = Join-Path $Root "coisas do codex\tools\luajit_canary\luajit.exe"

if (-not (Test-Path -LiteralPath $LuaJIT)) {
  Write-Warning "LuaJIT nao encontrado em: $LuaJIT"
  Write-Warning "Ajuste o caminho ou instale conforme AGENTS.md. Saindo com aviso (exit 0) para nao bloquear ambientes sem tool."
  exit 0
}

function Get-DefaultFiles([string]$area) {
  $list = @()
  if ($area -eq "server" -or $area -eq "all") {
    $list += @(
      "canary-3.4.1\data\scripts\spells\party\alchemia_party_buffs.lua",
      "canary-3.4.1\data\scripts\spells\attack\alchemia_mid_bursts.lua",
      "canary-3.4.1\data\scripts\custom_skills.lua",
      "canary-3.4.1\data\scripts\creaturescripts\others\power_beasts.lua"
    )
  }
  if ($area -eq "client" -or $area -eq "all") {
    $list += @(
      "client_run\modules\game_cyclopedia\tab\magicalArchives\magicalArchives.lua"
    )
  }
  return $list
}

if ($Files.Count -eq 0) {
  $Files = Get-DefaultFiles $Area
}

$failed = 0
$checked = 0
foreach ($rel in $Files) {
  $path = Join-Path $Root $rel
  if (-not (Test-Path -LiteralPath $path)) {
    Write-Host "SKIP (ausente): $rel" -ForegroundColor DarkYellow
    continue
  }
  $checked++
  $luaPath = $path -replace '\\', '/'
  $expr = "local f,err=loadfile([[$luaPath]]); if not f then error(err) end; print('OK ' .. [[$luaPath]])"
  & $LuaJIT -e $expr
  if ($LASTEXITCODE -ne 0) {
    Write-Host "FAIL: $rel" -ForegroundColor Red
    $failed++
  } else {
    Write-Host "OK: $rel" -ForegroundColor Green
  }
}

Write-Host "lua-validate: checked=$checked failed=$failed area=$Area"
if ($checked -eq 0) {
  Write-Warning "Nenhum arquivo encontrado para validar. Paths do pack podem diferir do seu tree."
  exit 0
}
if ($failed -gt 0) { exit 1 }
exit 0

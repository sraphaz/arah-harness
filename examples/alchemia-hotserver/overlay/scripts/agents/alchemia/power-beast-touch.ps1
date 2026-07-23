#Requires -Version 5.1
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$OutDir = Join-Path $Root "coisas do codex\arah-checklists"
if (-not (Test-Path -LiteralPath $OutDir)) {
  $OutDir = Join-Path $Root "docs\arah\checklists"
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
$Stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$Out = Join-Path $OutDir "power-beast-touch-$Stamp.md"

@"
# Checklist — power-beast-touch

Gerado em $Stamp.
Arquivo principal: ``canary-3.4.1/data/scripts/creaturescripts/others/power_beasts.lua``

## Regras (AGENTS.md)
- [ ] Selecao da beast = instancia (sprite, familiar, walking effect) — NAO ativa avatar
- [ ] Ativacao real: Animagis / Fusio Bestialis / legs
- [ ] ``USE_NATIVE_AVATAR_FIELD_ITEMS`` considerado
- [ ] Se field persistente visual: itens SEM dano nativo (evitar dano duplicado)
- [ ] Teste: seleção sem dano; ativação com VFX esperado; desativar limpa field

## Validacao
- [ ] lua-validate no script
- [ ] Dominio power-beasts no parecer do PR
"@ | Set-Content -LiteralPath $Out -Encoding UTF8

Write-Host "Checklist criado: $Out" -ForegroundColor Green
exit 0

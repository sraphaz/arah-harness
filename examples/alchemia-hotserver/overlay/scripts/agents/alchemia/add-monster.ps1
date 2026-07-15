#Requires -Version 5.1
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$OutDir = Join-Path $Root "coisas do codex\arah-checklists"
if (-not (Test-Path -LiteralPath $OutDir)) {
  $OutDir = Join-Path $Root "docs\arah\checklists"
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
$Stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$Out = Join-Path $OutDir "add-monster-$Stamp.md"

@"
# Checklist — add-monster (Alchemia)

Gerado em $Stamp.

## Intent
- [ ] Nome / race id:
- [ ] Tier / area do mapa:
- [ ] Experience / HP / armor alvos:
- [ ] Referencia de balance (Handbook ou planilha):

## Server
- [ ] XML em ``canary-3.4.1/data/monster/...``
- [ ] Loot table revisada (sem drop quebrado)
- [ ] Spells do monster (chances/interval) testados
- [ ] Spawn em ``data/world`` ou spawns.xml se aplicavel
- [ ] Creaturescripts so se necessario

## Validacao
- [ ] XML bem formado
- [ ] Spawn referencia o monster correto
- [ ] Dominio ``monsters-spawns`` consultado no PR
"@ | Set-Content -LiteralPath $Out -Encoding UTF8

Write-Host "Checklist criado: $Out" -ForegroundColor Green
exit 0

#Requires -Version 5.1
<#
  Checklist + template para nova magia Alchemia.
  Nao inventa numeros de balance — exige Handbook / decisao humana.
#>
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$OutDir = Join-Path $Root "coisas do codex\arah-checklists"
if (-not (Test-Path -LiteralPath $OutDir)) {
  # fallback se pasta Codex ainda nao existe
  $OutDir = Join-Path $Root "docs\arah\checklists"
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
$Stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$Out = Join-Path $OutDir "add-spell-$Stamp.md"

@"
# Checklist — add-spell (Alchemia)

Gerado em $Stamp. Preencher antes do PR.

## Intent
- [ ] Nome da magia (words / id):
- [ ] Vocacao(oes):
- [ ] Tipo: attack / healing / support / party / outro
- [ ] Elemento:
- [ ] Area / range:
- [ ] Referencia Handbook (capitulo/secao):

## Server (Canary)
- [ ] Script Lua em ``canary-3.4.1/data/scripts/spells/...``
- [ ] Registro XML/spells se aplicavel
- [ ] Se party buff: alinhar com ``alchemia_party_buffs.lua`` + ``custom_skills.lua``
- [ ] Se mid burst: padrao 3 hits, efeito no **ultimo** hit
- [ ] ``lua-validate`` no arquivo novo

## Client (OTClient)
- [ ] Override em Magical Archive (elemento, area, tipo, descricao, preview)
  - ``client_run/modules/game_cyclopedia/tab/magicalArchives/magicalArchives.lua``
  - OTUI se necessario: ``magicalArchives.otui``
- [ ] Sem rebuild C++ (so reload)

## Validacao
- [ ] Teste in-game (ou log) descrito
- [ ] Sem dano/heal absurdo sem planilha/decisao
- [ ] Entrada append no Handbook se sistema novo

## Domínios ARAH a consultar
- combat-magic
- client-ux
"@ | Set-Content -LiteralPath $Out -Encoding UTF8

Write-Host "Checklist criado: $Out" -ForegroundColor Green
Write-Host "Abra o arquivo, preencha, implemente, rode lua-validate, abra PR."
Write-Host ""
Write-Host "Paths tipicos:"
Write-Host "  canary-3.4.1\data\scripts\spells\"
Write-Host "  client_run\modules\game_cyclopedia\tab\magicalArchives\"
exit 0

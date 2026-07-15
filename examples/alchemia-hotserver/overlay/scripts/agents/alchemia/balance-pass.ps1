#Requires -Version 5.1
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$OutDir = Join-Path $Root "coisas do codex\arah-checklists"
if (-not (Test-Path -LiteralPath $OutDir)) {
  $OutDir = Join-Path $Root "docs\arah\checklists"
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
$Stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$Out = Join-Path $OutDir "balance-pass-$Stamp.md"
$Handbook = Join-Path $Root "coisas do codex\ALCHEMIA_ENGINEERING_HANDBOOK_V2_0.txt"

$handbookNote = if (Test-Path -LiteralPath $Handbook) {
  "Handbook encontrado: $Handbook — ler capitulo 19 antes de items.xml."
} else {
  "Handbook NAO encontrado no caminho esperado. Localize e leia o cap. 19 antes de continuar."
}

@"
# Gate — balance-pass (Alchemia)

Gerado em $Stamp.

## Aviso
$handbookNote

## Obrigatorios
- [ ] Li o capitulo 19 do Handbook (ou secao especifica de items)
- [ ] Planilha tecnica gerada/atualizada em ``coisas do codex``
- [ ] Decisao mais recente no Manual/Handbook registrada se conflitar
- [ ] Diff de ``items.xml`` / spells justificado linha a linha no PR
- [ ] Casos de teste (vocab + nivel alvo) descritos

## Proibido
- [ ] Nao "tunar de olho" dano/loot em massa sem planilha
- [ ] Nao deixar planilha so no chat — arquivar em ``coisas do codex``

## Dominios
- items-economy
- ops-codex
- combat-magic (se spells)
"@ | Set-Content -LiteralPath $Out -Encoding UTF8

Write-Host "Gate criado: $Out" -ForegroundColor Green
Write-Host $handbookNote
exit 0

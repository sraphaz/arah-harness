#Requires -Version 5.1
$ErrorActionPreference = "Stop"
$Root = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$OutDir = Join-Path $Root "coisas do codex\arah-checklists"
if (-not (Test-Path -LiteralPath $OutDir)) {
  $OutDir = Join-Path $Root "docs\arah\checklists"
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null
$Stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$Out = Join-Path $OutDir "codex-append-$Stamp.md"

@"
# Template append — Handbook / Manual Mestre

Gerado em $Stamp. Copiar para o fim do Handbook ou Manual (NAO apagar historico).

## Entrada
- Data:
- Autor/agente:
- Titulo curto:

### Sintoma
(o que o jogador/dev viu)

### Causa
(root cause)

### Arquivos
- path1
- path2

### Validacao
- comando/teste:
- resultado:

### Pendencias
- [ ]

### Decisao
(se conflitar com entrada antiga: esta vence por ser mais recente/especifica? sim/nao + por quê)
"@ | Set-Content -LiteralPath $Out -Encoding UTF8

Write-Host "Template criado: $Out" -ForegroundColor Green
Write-Host "Cole no Handbook/Manual Mestre sem apagar historico."
exit 0

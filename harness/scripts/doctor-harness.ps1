<#
.SYNOPSIS
  Diagnóstico de um repo governado: o que está instalado, o que falta, o que divergiu.
  Não altera nada — só relata (modo leitura).
.EXAMPLE
  ./doctor-harness.ps1 -Target ../meu-repo
  ./doctor-harness.ps1 -Target ../meu-repo -Report .harness/doctor-latest.md
.NOTES
  destino: arah-harness/scripts/doctor-harness.ps1
  CRITÉRIOS DE ACEITE:
  - repo sem harness-profile.yaml → status FAILING com instrução de instalação
  - template gerenciado ausente → DRIFT listando o arquivo
  - bloco gerenciado alterado localmente → DRIFT com diff resumido
  - specs inválidas (via validate-specs) → FAILING
  - tudo ok → status OK; com -Report, escreve relatório md datado
  - exit code: 0 (ok), 2 (drift), 1 (failing) — consumível por CI e pelo Workspace
#>
param(
  [Parameter(Mandatory = $true)] [string] $Target,
  [string] $Report
)

$ErrorActionPreference = 'Stop'
$HarnessRoot = Split-Path $PSScriptRoot -Parent
$issues = @{ failing = @(); drift = @(); info = @() }

# --- 1. harness-profile.yaml presente e legível ---
$profileFile = Join-Path $Target 'harness-profile.yaml'
if (-not (Test-Path $profileFile)) {
  Write-Host "FAILING: harness-profile.yaml ausente. Rode install-harness.ps1 -Target $Target -Profile <p>."
  exit 1
}
$raw = Get-Content $profileFile -Raw
$profile = if ($raw -match '(?m)^\s*profile:\s*(\S+)') { $Matches[1] } else { $null }
$version = if ($raw -match '(?m)^\s*version:\s*(\S+)') { $Matches[1] } else { $null }
if (-not $profile) { $issues.failing += 'harness-profile.yaml sem campo profile' }

# --- 2. templates gerenciados presentes? ---
# TODO(yaml): ler managed_blocks de verdade; convenção mínima abaixo
$expected = @('AGENTS.md', 'docs/governance/DEFINITION_OF_DONE.md', 'docs/specs/README.md')
foreach ($e in $expected) {
  if (-not (Test-Path (Join-Path $Target $e))) { $issues.drift += "ausente: $e" }
}

# --- 3. blocos gerenciados divergiram do template da versão instalada? ---
# TODO: comparar conteúdo entre marcadores com templates/ da versão $version
$issues.info += "comparação de blocos gerenciados: TODO (versão instalada: $version)"

# --- 4. specs válidas? (reusa validate-specs — paridade) ---
& "$PSScriptRoot/validate-specs.ps1" -Target $Target *> $null
if ($LASTEXITCODE -ne 0) { $issues.failing += 'validate-specs falhou (rode-o para detalhes)' }

# --- 5. agent graph válido? ---
& "$PSScriptRoot/validate-agent-graph.ps1" -Target $Target *> $null
if ($LASTEXITCODE -ne 0) { $issues.failing += 'validate-agent-graph falhou (rode-o para detalhes)' }

# --- veredito ---
$status = if ($issues.failing.Count) { 'failing' } elseif ($issues.drift.Count) { 'drift' } else { 'ok' }
Write-Host "DOCTOR [$status] profile=$profile version=$version"
$issues.failing | ForEach-Object { Write-Host "  FAILING $_" }
$issues.drift   | ForEach-Object { Write-Host "  DRIFT   $_" }
$issues.info    | ForEach-Object { Write-Host "  info    $_" }

if ($Report) {
  $md = @("# Doctor — $(Split-Path $Target -Leaf)", "", "- status: **$status**",
          "- profile: $profile @ $version", "- em: $(Get-Date -Format o)", "",
          ($issues.failing + $issues.drift | ForEach-Object { "- [ ] $_" })) -join "`n"
  New-Item -ItemType Directory -Force -Path (Split-Path (Join-Path $Target $Report)) | Out-Null
  Set-Content -Path (Join-Path $Target $Report) -Value $md
}

exit @{ ok = 0; drift = 2; failing = 1 }[$status]

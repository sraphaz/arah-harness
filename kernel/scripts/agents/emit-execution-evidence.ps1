#Requires -Version 5.1
<#
.SYNOPSIS
  Emite um documento ExecutionEvidence v1 (contrato da federação).

.DESCRIPTION
  Skill de bootstrap para consumidores do ARAH Harness. Não inventa campos —
  só os do schema schemas/contracts/execution-evidence-1.0.0.schema.yaml.
  Hawk (e outros produtos) devem chamar este contrato, não um formato próprio.

.EXAMPLE
  pwsh scripts/agents/emit-execution-evidence.ps1 -DryRun -RequestId 01J4YAZ0FJ8Q4V9WYD3S6BTPZ7 -CorrelationId 01J4YAY7Q2M3N4P5R6S7T8V9W0 -Outcome completed -ProjectId iautos
#>
param(
    [string]$RequestId = $env:ARAH_EVIDENCE_REQUEST_ID,
    [string]$CorrelationId = $env:ARAH_EVIDENCE_CORRELATION_ID,
    [ValidateSet('completed', 'partial', 'failed', 'blocked', 'cancelled')]
    [string]$Outcome = $env:ARAH_EVIDENCE_OUTCOME,
    [string]$ProjectId = $(if ($env:ARAH_EVIDENCE_PROJECT_ID) { $env:ARAH_EVIDENCE_PROJECT_ID } else { 'unknown' }),
    [string]$DemandId,
    [string]$EvidenceId,
    [string]$ProducedBy = 'agent:arah-harness@emit-execution-evidence',
    [string]$Notes,
    [string]$PrUrl,
    [int]$PrNumber = 0,
    [string]$OutDir = 'docs/execution-evidence',
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'
$Crockford = '0123456789ABCDEFGHJKMNPQRSTVWXYZ'

function Test-Ulid([string]$Value) {
    return $Value -match '^[0-9A-HJKMNP-TV-Z]{26}$'
}

function New-Ulid {
    $t = [DateTimeOffset]::UtcNow.ToUnixTimeMilliseconds()
    $chars = New-Object char[] 10
    for ($i = 9; $i -ge 0; $i--) {
        $chars[$i] = $Crockford[$t % 32]
        $t = [math]::Floor($t / 32)
    }
    $rand = -join ((1..16) | ForEach-Object { $Crockford[(Get-Random -Maximum 32)] })
    return (-join $chars) + $rand
}

function ConvertTo-YamlScalar($Value) {
    if ($null -eq $Value) { return 'null' }
    $text = [string]$Value
    if ($text -match '[:#\n]' -or $text -match '^\s|\s$') {
        return ($text | ConvertTo-Json -Compress)
    }
    return $text
}

if (-not $RequestId -or -not $CorrelationId -or -not $Outcome) {
    Write-Host @'
emit-execution-evidence — ExecutionEvidence v1 (ARAH Harness)

  -RequestId <ulid>        obrigatório
  -CorrelationId <ulid>    obrigatório
  -Outcome completed|partial|failed|blocked|cancelled
  -ProjectId <slug>
  -DryRun                  valida e imprime; não grava

Skill: invoke-skill.ps1 -Skill emit-execution-evidence
(use env ARAH_EVIDENCE_REQUEST_ID / _CORRELATION_ID / _OUTCOME)
'@
    exit 2
}

if (-not (Test-Ulid $RequestId)) { Write-Error "request_id não é ULID: $RequestId"; exit 1 }
if (-not (Test-Ulid $CorrelationId)) { Write-Error "correlation_id não é ULID: $CorrelationId"; exit 1 }

if (-not $EvidenceId) { $EvidenceId = New-Ulid }
$EvidenceId = $EvidenceId.ToUpperInvariant()
if (-not (Test-Ulid $EvidenceId)) { Write-Error "evidence_id não é ULID: $EvidenceId"; exit 1 }

$producedAt = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffZ')
$demandLine = if ($null -eq $DemandId -or $DemandId -eq '') { 'null' } else { ConvertTo-YamlScalar $DemandId }

$evidenceBlock = "  {}"
if ($PrUrl -or $PrNumber -gt 0) {
    $n = if ($PrNumber -gt 0) { $PrNumber } else { 'null' }
    $u = if ($PrUrl) { ConvertTo-YamlScalar $PrUrl } else { 'null' }
    $evidenceBlock = @"
  pr:
    number: $n
    url: $u
    state: open
    merged_by: null
"@
}

if ($Outcome -eq 'completed' -and $evidenceBlock.Trim() -eq '{}') {
    Write-Error 'outcome completed exige ao menos um campo em evidence (ex.: -PrUrl)'
    exit 1
}

$notesLine = if ($Notes) { "notes: $(ConvertTo-YamlScalar $Notes)" } else { $null }

$yaml = @"
schema_version: 1.0.0
evidence_id: $EvidenceId
request_id: $RequestId
correlation_id: $CorrelationId
project_id: $(ConvertTo-YamlScalar $ProjectId)
demand_id: $demandLine
spec_id: null
task_id: task_$EvidenceId
run_id: run_$EvidenceId
outcome: $Outcome
produced_by: $(ConvertTo-YamlScalar $ProducedBy)
produced_at: $producedAt
evidence:
$evidenceBlock
$notesLine
"@.TrimEnd() + "`n"

if ($DryRun) {
    Write-Output $yaml
    Write-Host "dry-run ok evidence_id=$EvidenceId"
    exit 0
}

$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$destDir = Join-Path $Root $OutDir
New-Item -ItemType Directory -Force -Path $destDir | Out-Null
$dest = Join-Path $destDir "$EvidenceId.yaml"
Set-Content -Path $dest -Value $yaml -Encoding utf8
Write-Host "wrote $dest"
exit 0

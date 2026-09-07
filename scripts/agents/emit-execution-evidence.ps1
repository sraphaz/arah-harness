#Requires -Version 5.1
<#
.SYNOPSIS
  Emite um documento ExecutionEvidence v1 (contrato da federação).

.DESCRIPTION
  Skill de bootstrap para consumidores do ARAH Harness. Não inventa campos —
  só os do schema schemas/contracts/execution-evidence-1.0.0.schema.yaml.
  Hawk (e outros produtos) devem chamar este contrato, não um formato próprio.
  task_id / run_id são os da execução reportada (não derivados do evidence_id).

.EXAMPLE
  pwsh scripts/agents/emit-execution-evidence.ps1 -DryRun -RequestId 01J4YAZ0FJ8Q4V9WYD3S6BTPZ7 -CorrelationId 01J4YAY7Q2M3N4P5R6S7T8V9W0 -Outcome completed -ProjectId iautos -TaskId task-20260811-174210000-iautos -RunId run-20260811-174210000-iautos -PrUrl https://github.com/sraphaz/iautos/pull/210
#>
param(
    [string]$RequestId = $env:ARAH_EVIDENCE_REQUEST_ID,
    [string]$CorrelationId = $env:ARAH_EVIDENCE_CORRELATION_ID,
    [ValidateSet('completed', 'partial', 'failed', 'blocked', 'cancelled')]
    [string]$Outcome = $env:ARAH_EVIDENCE_OUTCOME,
    [string]$ProjectId = $(if ($env:ARAH_EVIDENCE_PROJECT_ID) { $env:ARAH_EVIDENCE_PROJECT_ID } else { 'unknown' }),
    [string]$DemandId = $env:ARAH_EVIDENCE_DEMAND_ID,
    [string]$SpecId = $env:ARAH_EVIDENCE_SPEC_ID,
    [string]$TaskId = $env:ARAH_EVIDENCE_TASK_ID,
    [string]$RunId = $env:ARAH_EVIDENCE_RUN_ID,
    [string]$EvidenceId = $env:ARAH_EVIDENCE_EVIDENCE_ID,
    [string]$ProducedBy = $(if ($env:ARAH_EVIDENCE_PRODUCED_BY) { $env:ARAH_EVIDENCE_PRODUCED_BY } else { 'agent:arah-harness@emit-execution-evidence' }),
    [string]$Notes = $env:ARAH_EVIDENCE_NOTES,
    [string]$PrUrl = $env:ARAH_EVIDENCE_PR_URL,
    [int]$PrNumber = 0,
    [string]$PrState = $env:ARAH_EVIDENCE_PR_STATE,
    [string]$MergedBy = $env:ARAH_EVIDENCE_MERGED_BY,
    [string]$OutDir = $(if ($env:ARAH_EVIDENCE_OUT_DIR) { $env:ARAH_EVIDENCE_OUT_DIR } else { 'docs/execution-evidence' }),
    [switch]$DryRun
)

$ErrorActionPreference = 'Stop'
$Crockford = '0123456789ABCDEFGHJKMNPQRSTVWXYZ'
$TaskIdPattern = '^task-[A-Za-z0-9._-]+$'
$RunIdPattern = '^run-[A-Za-z0-9._-]+$'

if ($PrNumber -le 0 -and $env:ARAH_EVIDENCE_PR_NUMBER) {
    $PrNumber = [int]$env:ARAH_EVIDENCE_PR_NUMBER
}
if (-not $DryRun -and $env:ARAH_EVIDENCE_DRY_RUN -match '^(1|true|yes)$') {
    $DryRun = $true
}

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
    return ([string]$Value | ConvertTo-Json -Compress)
}

if (-not $RequestId -or -not $CorrelationId -or -not $Outcome -or -not $TaskId -or -not $RunId) {
    Write-Host @'
emit-execution-evidence — ExecutionEvidence v1 (ARAH Harness)

  -RequestId <ulid>        obrigatório
  -CorrelationId <ulid>    obrigatório
  -Outcome completed|partial|failed|blocked|cancelled
  -TaskId task-…           obrigatório (execution-contract)
  -RunId run-…             obrigatório (derivado da task no kernel)
  -ProjectId <slug>
  -PrUrl / -PrNumber / -PrState / -Notes
  -DryRun                  valida e imprime; não grava

Skill: invoke-skill.ps1 -Skill emit-execution-evidence
(env ARAH_EVIDENCE_REQUEST_ID / _CORRELATION_ID / _OUTCOME / _TASK_ID / _RUN_ID
 / _PR_URL / _PR_NUMBER / _PR_STATE / _NOTES / _DEMAND_ID / _SPEC_ID / _DRY_RUN)
'@
    exit 2
}

if (-not (Test-Ulid $RequestId)) { Write-Error "request_id não é ULID: $RequestId"; exit 1 }
if (-not (Test-Ulid $CorrelationId)) { Write-Error "correlation_id não é ULID: $CorrelationId"; exit 1 }
if ($TaskId -notmatch $TaskIdPattern) {
    Write-Error "task_id deve casar ${TaskIdPattern} (execution-contract): $TaskId"
    exit 1
}
if ($RunId -notmatch $RunIdPattern) {
    Write-Error "run_id deve casar ${RunIdPattern}: $RunId"
    exit 1
}

if (-not $EvidenceId) { $EvidenceId = New-Ulid }
$EvidenceId = $EvidenceId.ToUpperInvariant()
if (-not (Test-Ulid $EvidenceId)) { Write-Error "evidence_id não é ULID: $EvidenceId"; exit 1 }

$producedAt = [DateTime]::UtcNow.ToString('yyyy-MM-ddTHH:mm:ss.fffZ')
$demandLine = if ($null -eq $DemandId -or $DemandId -eq '') { 'null' } else { ConvertTo-YamlScalar $DemandId }
$specLine = if ($null -eq $SpecId -or $SpecId -eq '') { 'null' } else { ConvertTo-YamlScalar $SpecId }

$evidenceBlock = "  {}"
if ($PrUrl -or $PrNumber -gt 0) {
    if (-not $PrState) { $PrState = 'open' }
    if ($PrState -notin @('open', 'merged', 'closed')) {
        Write-Error "evidence.pr.state inválido: $PrState (open|merged|closed)"
        exit 1
    }
    $n = if ($PrNumber -gt 0) { $PrNumber } else { 'null' }
    $u = if ($PrUrl) { ConvertTo-YamlScalar $PrUrl } else { 'null' }
    $mergedLine = if ($MergedBy) { ConvertTo-YamlScalar $MergedBy } else { 'null' }
    $evidenceBlock = @"
  pr:
    number: $n
    url: $u
    state: $(ConvertTo-YamlScalar $PrState)
    merged_by: $mergedLine
"@
}

if ($Outcome -eq 'completed' -and $evidenceBlock.Trim() -eq '{}') {
    Write-Error 'outcome completed exige ao menos um campo em evidence (ex.: -PrUrl ou ARAH_EVIDENCE_PR_URL)'
    exit 1
}

$notesLine = if ($Notes) { "notes: $(ConvertTo-YamlScalar $Notes)" } else { $null }

$yaml = @"
schema_version: $(ConvertTo-YamlScalar '1.0.0')
evidence_id: $(ConvertTo-YamlScalar $EvidenceId)
request_id: $(ConvertTo-YamlScalar $RequestId)
correlation_id: $(ConvertTo-YamlScalar $CorrelationId)
project_id: $(ConvertTo-YamlScalar $ProjectId)
demand_id: $demandLine
spec_id: $specLine
task_id: $(ConvertTo-YamlScalar $TaskId)
run_id: $(ConvertTo-YamlScalar $RunId)
outcome: $(ConvertTo-YamlScalar $Outcome)
produced_by: $(ConvertTo-YamlScalar $ProducedBy)
produced_at: $(ConvertTo-YamlScalar $producedAt)
evidence:
$evidenceBlock
$notesLine
"@.TrimEnd() + "`n"

if ($DryRun) {
    Write-Output $yaml
    Write-Host "dry-run ok evidence_id=$EvidenceId task_id=$TaskId run_id=$RunId"
    exit 0
}

$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$destDir = Join-Path $Root $OutDir
New-Item -ItemType Directory -Force -Path $destDir | Out-Null
$dest = Join-Path $destDir "$EvidenceId.yaml"
Set-Content -Path $dest -Value $yaml -Encoding utf8
Write-Host "wrote $dest"
exit 0

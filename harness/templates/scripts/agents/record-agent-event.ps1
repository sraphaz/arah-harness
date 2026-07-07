#Requires -Version 5.1
<#
.SYNOPSIS
  Registra evento de agente na trilha de auditoria ARAH (append-only JSONL).
#>
param(
    [Parameter(Mandatory = $true)][string]$AgentId,
    [Parameter(Mandatory = $true)][string]$Action,
    [ValidateSet('ok', 'blocked', 'denied', 'error', 'pending')][string]$Outcome = 'ok',
    [string]$AutonomyLevel = 'activate',
    [string]$Details = '',
    [string]$CorrelationId = '',
    [string]$HumanGate = '',
    [string]$Project = '',
    [switch]$Blocked
)

$ErrorActionPreference = 'Stop'
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot '..\..')
$AuditDir = Join-Path $RepoRoot '.arah\audit'
$AuditFile = Join-Path $AuditDir 'events.jsonl'
$ObsDir = Join-Path $RepoRoot '.arah\observability'
$SummaryFile = Join-Path $ObsDir 'summary.yaml'

if (-not $Project) { $Project = Split-Path $RepoRoot -Leaf }
New-Item -ItemType Directory -Path $AuditDir -Force | Out-Null
New-Item -ItemType Directory -Path $ObsDir -Force | Out-Null
if (-not $CorrelationId) { $CorrelationId = [guid]::NewGuid().ToString('N').Substring(0, 12) }

$event = [ordered]@{
    ts = (Get-Date).ToUniversalTime().ToString('o')
    correlation_id = $CorrelationId
    project = $Project
    agent_id = $AgentId
    action = $Action
    autonomy_level = $AutonomyLevel
    outcome = if ($Blocked) { 'blocked' } else { $Outcome }
    human_gate = if ($HumanGate) { $HumanGate } else { $null }
    details = $Details
}
Add-Content -Path $AuditFile -Value ($event | ConvertTo-Json -Compress -Depth 5) -Encoding UTF8

$total = 0
if (Test-Path $SummaryFile) {
    $existing = Get-Content $SummaryFile -Raw
    if ($existing -match 'total_events:\s*(\d+)') { $total = [int]$Matches[1] }
}
$total++
Set-Content -Path $SummaryFile -Value @"
updated_at: $($event.ts)
project: $Project
total_events: $total
last_agent: $AgentId
last_action: $Action
last_outcome: $($event.outcome)
"@ -Encoding UTF8

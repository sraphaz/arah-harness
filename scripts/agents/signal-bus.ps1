#Requires -Version 5.1
<#
.SYNOPSIS
  Barramento de sinais tipados entre células do organismo ARAH.
.DESCRIPTION
  Comunicação orgânica e auditável: append-only em .arah/bus/signals.jsonl.
  Tipos: attract | consult | propose | acknowledge | coalesce | evolve | status.
  Default econômico: sinais são artefatos (não chat multi-turno).
.EXAMPLE
  ./signal-bus.ps1 -From orchestrator -SignalTo backend -SignalType attract -Topic delivery
  ./signal-bus.ps1 -List -Last 10
  ./signal-bus.ps1 -From qa -SignalTo '*' -SignalType propose -Topic craft -Payload '{"change":"add craft-review to frontend paths"}'
#>
param(
    [string]$From = '',
    # Destination cell/tissue ('*' = broadcast). Named SignalTo to avoid -To/-Topic prefix clash.
    [Alias('To')]
    [string]$SignalTo = '*',
    [ValidateSet('attract', 'consult', 'propose', 'acknowledge', 'coalesce', 'evolve', 'status', '')]
    [Alias('Type')]
    [string]$SignalType = '',
    [string]$Topic = 'general',
    [string]$Payload = '',
    [string]$CorrelationId = '',
    [ValidateSet('observe', 'consult', 'route', 'activate', 'invoke_skill', 'side_effect', 'public')]
    [string]$AutonomyLevel = 'consult',
    [switch]$List,
    [int]$Last = 20
)

$ErrorActionPreference = 'Stop'
$Root = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '../..')).Path
$BusDir = Join-Path (Join-Path $Root '.arah') 'bus'
$BusFile = Join-Path $BusDir 'signals.jsonl'

if ($List) {
    if (-not (Test-Path $BusFile)) {
        Write-Host "signal-bus: empty (no signals yet)"
        exit 0
    }
    Get-Content $BusFile | Select-Object -Last $Last
    exit 0
}

if (-not $From -or -not $SignalType) {
    Write-Error "Use: signal-bus.ps1 -From <cell> -SignalType <attract|consult|propose|...> [-SignalTo cell|tissue|*] [-Topic t] [-Payload json]"
    exit 1
}

# Soft validate against organism manifest if present
$manifest = Join-Path (Join-Path (Join-Path $Root 'docs') '_meta') 'organism.manifest.yaml'
if (Test-Path $manifest) {
    $mraw = Get-Content $manifest -Raw
    if ($mraw -notmatch "(?m)^\s+- id:\s*$([regex]::Escape($From))\s*$") {
        Write-Warning "signal-bus: emitter '$From' not in organism.manifest.yaml (allowed for bootstrap)"
    }
}

New-Item -ItemType Directory -Path $BusDir -Force | Out-Null

if (-not $CorrelationId) {
    $CorrelationId = [guid]::NewGuid().ToString('N').Substring(0, 12)
}

$signalId = [guid]::NewGuid().ToString('N').Substring(0, 10)
$payloadObj = $null
if ($Payload) {
    try { $payloadObj = $Payload | ConvertFrom-Json } catch { $payloadObj = @{ text = $Payload } }
}

$event = [ordered]@{
    ts = (Get-Date).ToUniversalTime().ToString('o')
    signal_id = $signalId
    type = $SignalType
    from = $From
    to = $SignalTo
    topic = $Topic
    payload = $payloadObj
    correlation_id = $CorrelationId
    autonomy_level = $AutonomyLevel
}

$line = ($event | ConvertTo-Json -Compress -Depth 6)
Add-Content -Path $BusFile -Value $line -Encoding UTF8
Write-Host "signal-bus: $SignalType from=$From to=$SignalTo topic=$Topic id=$signalId"

# Mirror propose/evolve into live diagnostics when available
$diagDir = Join-Path (Join-Path $Root '.cursor') 'arah-live'
if ($SignalType -in @('propose', 'evolve', 'coalesce')) {
    if (-not (Test-Path $diagDir)) { New-Item -ItemType Directory -Path $diagDir -Force | Out-Null }
    $diag = Join-Path $diagDir 'diagnostics.jsonl'
    Add-Content -Path $diag -Value $line -Encoding UTF8
}

$record = Join-Path $PSScriptRoot 'record-agent-event.ps1'
if (Test-Path $record) {
    & $record -AgentId $From -Action "signal.$SignalType" -Outcome ok -AutonomyLevel $AutonomyLevel -CorrelationId $CorrelationId -Details "to=$SignalTo;topic=$Topic" 2>$null
}

exit 0

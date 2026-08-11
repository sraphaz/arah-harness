<#
.SYNOPSIS
  Valida Agent Graph + harness-model (domain agents, governance, audit, observability).
.EXAMPLE
  ./validate-agent-graph.ps1 -Target ../meu-repo
#>
param(
  [string] $Target = (Get-Location).Path,
  [switch]$Strict
)

$ErrorActionPreference = 'Stop'
$LibPath = Join-Path $PSScriptRoot 'harness-model-lib.ps1'
if (Test-Path $LibPath) { . $LibPath }

$GraphFile = Join-Path $Target 'agent-graph.yaml'
$AgentsDir = Join-Path $Target '.agents'
$failures = @()

if (-not (Test-Path $GraphFile) -and -not (Test-Path $AgentsDir)) {
  if (Test-Path $LibPath) {
    $modelResult = Test-HarnessModelCompleteness -Target $Target -Strict:$Strict
    if ($modelResult.failures.Count -gt 0) {
      $modelResult.failures | ForEach-Object { Write-Host $_ }
      exit 1
    }
  }
  Write-Host "OK (aviso): repo sem agentes declarados — model check only."
  exit 0
}

$agentIds = @()
if (Test-Path $AgentsDir) {
  $yamlFiles = Get-ChildItem $AgentsDir -Filter '*.yaml' -Recurse |
    Where-Object { $_.Name -match '\.agent\.yaml$' -or ($_.DirectoryName -eq $AgentsDir -and $_.Extension -eq '.yaml' -and $_.Name -ne 'choreography.yaml' -and $_.Name -notlike 'choreography.*') }
  foreach ($f in $yamlFiles) {
    $raw = Get-Content $f.FullName -Raw
    if ($raw -notmatch '(?m)^id:\s*(\S+)') { continue }
    $id = $Matches[1]
    if ($agentIds -contains $id) { continue }
    $agentIds += $id
    if ($raw -match '(?m)^type:\s*domain') {
      if ($raw -match '(?m)max_autonomy:\s*(\S+)') {
        $ma = $Matches[1]
        if ((Get-AutonomyRank $ma) -gt $Script:ConsultMaxRank) {
          $failures += "FAIL domain agent ${id}: max_autonomy '$ma' exceeds consult"
        }
      }
    }
  }
}

$choreo = Join-Path $AgentsDir 'choreography.yaml'
if (Test-Path $choreo) {
  $craw = Get-Content $choreo -Raw
  foreach ($id in $Script:DomainAgentIds) {
    if ($craw -match "(?m)id:\s*$id\b") {
      if (-not (Test-AgentManifestExists -Target $Target -AgentId $id)) {
        $failures += "FAIL choreography references '$id' but manifest missing in .agents/domain/"
      }
    }
  }
}

if (Test-Path $GraphFile) {
  $raw = Get-Content $GraphFile -Raw
  $edges = [regex]::Matches($raw, '-\s*\{\s*from:\s*([\w:.-]+),\s*to:\s*([\w:.-]+)(?:.*?approval:\s*(true|false))?') |
           ForEach-Object { @{ from = $_.Groups[1].Value; to = $_.Groups[2].Value; approval = ($_.Groups[3].Value -eq 'true') } }
  foreach ($e in $edges) {
    foreach ($n in @($e.from, $e.to)) {
      if ($n -notlike 'human:*' -and $agentIds -notcontains $n) {
        $failures += "FAIL grafo: node '$n' não existe em .agents/"
      }
    }
  }
}

if (Test-Path $LibPath) {
  $modelResult = Test-HarnessModelCompleteness -Target $Target -Strict:$Strict
  $failures += @($modelResult.failures)
  $modelResult.warnings | ForEach-Object { Write-Warning $_ }
}

if ($failures.Count -gt 0) { $failures | ForEach-Object { Write-Host $_ }; exit 1 }
Write-Host "OK agent graph + harness-model válidos ($($agentIds.Count) agentes)"
exit 0

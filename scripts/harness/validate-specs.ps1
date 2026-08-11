<#
.SYNOPSIS
  Valida specs + completude do harness-model (domain agents, governance, audit, observability).
.EXAMPLE
  ./validate-specs.ps1 -Target ../meu-repo
#>
param(
  [Parameter(Mandatory = $true)] [string] $Target,
  [string[]] $ChangedFiles = @(),
  [string[]] $Labels = @(),
  [switch]$Strict
)

$ErrorActionPreference = 'Stop'
$LibPath = Join-Path $PSScriptRoot 'harness-model-lib.ps1'
if (Test-Path $LibPath) { . $LibPath }

$SpecsDir = Join-Path $Target 'docs/specs'
$failures = @()

if (-not (Test-Path $SpecsDir)) {
  Write-Host "FAIL docs/specs/ não existe. Rode install-harness.ps1 primeiro."
  exit 1
}

$required = @('id','title','status','owner','covers','updated_at')
$specs = Get-ChildItem $SpecsDir -Filter '*.md' -Recurse | Where-Object {
  $_.Name -ne 'README.md' -and $_.Name -ne 'REGISTRY.md'
}

foreach ($spec in $specs) {
  $raw = Get-Content $spec.FullName -Raw
  if ($raw -notmatch '(?s)^---\s*\n(.*?)\n---') {
    $failures += "FAIL $($spec.Name): sem frontmatter YAML"
    continue
  }
  $fm = $Matches[1]
  foreach ($field in $required) {
    if ($fm -notmatch "(?m)^${field}:") { $failures += "FAIL $($spec.Name): campo '$field' ausente" }
  }
  if ($fm -match '(?m)^status:\s*active' -and $fm -match '(?m)^covers:\s*\[\s*\]') {
    $failures += "FAIL $($spec.Name): active com covers vazio"
  }
}

if ($ChangedFiles.Count -gt 0 -and $Labels -notcontains 'spec-reviewed') {
  $activeCovers = @{}
  foreach ($spec in $specs) {
    $raw = Get-Content $spec.FullName -Raw
    if ($raw -match '(?m)^status:\s*active') {
      $globs = [regex]::Matches($raw, '"([^"]+\*[^"]*)"') | ForEach-Object { $_.Groups[1].Value }
      $activeCovers[$spec.Name] = $globs
    }
  }
  $specChanged = $ChangedFiles | Where-Object { $_ -like 'docs/specs/*' }
  foreach ($file in ($ChangedFiles | Where-Object { $_ -notlike 'docs/specs/*' })) {
    foreach ($entry in $activeCovers.GetEnumerator()) {
      foreach ($glob in $entry.Value) {
        if ($file -like $glob -and $specChanged.Count -eq 0) {
          $failures += "FAIL spec-before-work: '$file' coberto por $($entry.Key), mas nenhuma spec mudou (ou label spec-reviewed)"
        }
      }
    }
  }
}

# --- harness-model completeness (first-class) ---
if (Test-Path $LibPath) {
  $modelResult = Test-HarnessModelCompleteness -Target $Target -Strict:$Strict
  $failures += @($modelResult.failures)
  $modelResult.warnings | ForEach-Object { Write-Warning $_ }
  if ($modelResult.tier) { Write-Host "  model tier: $($modelResult.tier)" }
}

if ($failures.Count -gt 0) { $failures | ForEach-Object { Write-Host $_ }; exit 1 }
Write-Host "OK $($specs.Count) specs + harness-model válidos em $Target"
exit 0

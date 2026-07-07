#Requires -Version 5.1
<#
.SYNOPSIS
  Wrapper — validação completa do agent graph (kernel) + harness-model.
  Sem -Target: valida o repo atual (arah-harness self-hosted).
  Com -Target: valida repo-alvo (paridade com harness/scripts/).
#>
param(
    [string]$Target = '',
    [switch]$Json,
    [switch]$Strict
)

$ErrorActionPreference = 'Stop'

if ($Target) {
    $repoRoot = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
    $harnessScript = Join-Path $repoRoot 'harness\scripts\validate-agent-graph.ps1'
    if (-not (Test-Path $harnessScript)) {
        $harnessScript = Join-Path $PSScriptRoot 'validate-agent-graph-target.ps1'
    }
    & $harnessScript -Target $Target -Strict:$Strict
    exit $LASTEXITCODE
}

$repoRoot = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$kernelScript = Join-Path $repoRoot 'kernel\scripts\harness\validate-agent-graph.ps1'
if (-not (Test-Path $kernelScript)) {
    Write-Error "Kernel validate script not found: $kernelScript"
    exit 1
}
& $kernelScript -Json:$Json -Strict:$Strict
exit $LASTEXITCODE

#Requires -Version 5.1
<#
.SYNOPSIS
  Smoke tests for experimental assess-repo / repo-perspective-assess.
#>
$ErrorActionPreference = 'Stop'
$Root = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$Pwsh = if (Get-Command pwsh -ErrorAction SilentlyContinue) { 'pwsh' } else { 'powershell' }
$Script = Join-Path $Root 'scripts/agents/assess-repo.ps1'
$Fail = 0

function Assert-True {
    param([bool]$Cond, [string]$Msg)
    if ($Cond) { Write-Host "  OK  $Msg" }
    else { Write-Host "  FAIL $Msg"; $script:Fail++ }
}

Write-Host "=== assess-repo tests → $Root ==="
Assert-True (Test-Path -LiteralPath $Script) 'assess-repo.ps1 exists'

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("arah-assess-" + [guid]::NewGuid().ToString('n'))
New-Item -ItemType Directory -Path $tmp -Force | Out-Null
try {
    # Empty-ish repo
    Set-Content -LiteralPath (Join-Path $tmp 'README.md') -Value '# empty seed' -Encoding UTF8
    $agentsDir = Join-Path $tmp '.agents'
    New-Item -ItemType Directory -Path $agentsDir -Force | Out-Null
    @'
id: qa
name: QA
description: test lens
execution_role:
  can_route: false
  can_execute: true
  can_consult: false
  can_review: true
scope:
  paths:
    - "**"
'@ | Set-Content (Join-Path $agentsDir 'qa.agent.yaml') -Encoding UTF8
    @'
id: backend
name: Backend
description: backend lens
execution_role:
  can_route: false
  can_execute: true
  can_consult: true
  can_review: false
scope:
  paths:
    - backend/**
'@ | Set-Content (Join-Path $agentsDir 'backend.agent.yaml') -Encoding UTF8

    $out = Join-Path $tmp 'visions-out'
    & $Pwsh -NoProfile -ExecutionPolicy Bypass -File $Script -RepoRoot $tmp -OutDir $out -Force
    Assert-True ($LASTEXITCODE -eq 0) 'exit 0 on emptyish repo'
    Assert-True (Test-Path (Join-Path $out 'qa.md')) 'qa.md written'
    Assert-True (Test-Path (Join-Path $out 'backend.md')) 'backend.md written'
    Assert-True (Test-Path (Join-Path $out 'README.md')) 'index README written'
    Assert-True (Test-Path (Join-Path $out 'summary.yaml')) 'summary.yaml written'
    $qa = Get-Content (Join-Path $out 'qa.md') -Raw
    Assert-True ($qa -match 'As-Is') 'qa vision has As-Is'
    Assert-True ($qa -match 'Gaps') 'qa vision has Gaps'
    Assert-True ($qa -match 'To-Be') 'qa vision has To-Be'
    Assert-True ($qa -match 'bootstrap') 'emptyish mentions bootstrap'

    # Filter agents
    $out2 = Join-Path $tmp 'visions-filter'
    & $Pwsh -NoProfile -ExecutionPolicy Bypass -File $Script -RepoRoot $tmp -OutDir $out2 -Agents 'qa' -Force
    Assert-True ($LASTEXITCODE -eq 0) 'filter exit 0'
    Assert-True (Test-Path (Join-Path $out2 'qa.md')) 'filter wrote qa'
    Assert-True (-not (Test-Path (Join-Path $out2 'backend.md'))) 'filter skipped backend'

    # Dry-run does not create when using fresh dir — script still mkdir; check dry-run flag path
    $out3 = Join-Path $tmp 'visions-dry'
    & $Pwsh -NoProfile -ExecutionPolicy Bypass -File $Script -RepoRoot $tmp -OutDir $out3 -DryRun
    Assert-True ($LASTEXITCODE -eq 0) 'dry-run exit 0'
} finally {
    Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

# CLI wiring
$cli = Join-Path $Root 'cli/arah.ps1'
$cliRaw = Get-Content $cli -Raw
Assert-True ($cliRaw -match "assess-repo") 'CLI mentions assess-repo'
Assert-True ($cliRaw -match "bootstrap-vision") 'CLI alias bootstrap-vision'
Assert-True (Test-Path (Join-Path $Root 'kernel/.skills/repo-perspective-assess.skill.yaml')) 'skill in kernel'
Assert-True (Test-Path (Join-Path $Root 'docs/REPO_VISIONS.md')) 'docs present'

if ($Fail -gt 0) {
    Write-Host "`nFAILED: $Fail assertion(s)"
    exit 1
}
Write-Host "`nAll assess-repo tests passed."
exit 0

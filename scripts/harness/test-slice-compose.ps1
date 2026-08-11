#Requires -Version 5.1
<#
.SYNOPSIS
  Smoke tests for experimental slice-compose / arah slice plan.
#>
$ErrorActionPreference = 'Stop'
$Root = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$Pwsh = if (Get-Command pwsh -ErrorAction SilentlyContinue) { 'pwsh' } else { 'powershell' }
$Script = Join-Path $Root 'scripts/agents/slice-compose.ps1'
$Fail = 0

function Assert-True {
    param([bool]$Cond, [string]$Msg)
    if ($Cond) { Write-Host "  OK  $Msg" }
    else { Write-Host "  FAIL $Msg"; $script:Fail++ }
}

Write-Host "=== slice-compose tests → $Root ==="
Assert-True (Test-Path -LiteralPath $Script) 'slice-compose.ps1 exists'

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("arah-slice-" + [guid]::NewGuid().ToString('n'))
New-Item -ItemType Directory -Path $tmp -Force | Out-Null
try {
    $docs = Join-Path $tmp 'docs'
    $visions = Join-Path $docs '_arah\visions'
    $plans = Join-Path $docs '_arah\slice-plans'
    New-Item -ItemType Directory -Path $visions -Force | Out-Null

    @'
# Backlog

## E1 — Boundaries

### E1-S4 — Adapter IndexedDB implementando as ports
- **P0** · Status: `todo` · Depende: E1-S3
- **Em uma frase:** Persistência local continua via IndexedDB, agora atrás das ports.
- **Pronto quando:** Regressões de persistência/features verdes.

### E1-S5 — OpenAPI
- **P0** · Status: `todo`
'@ | Set-Content (Join-Path $docs 'BACKLOG_HUB_PRODUCTION.md') -Encoding UTF8

    @'
# Agent backlog
version: 2
agent: qa
updated_at: "2026-07-28T00:00:00-03:00"
items:
  - id: qa-BL-005
    title: "Coverage mínimo unitário no domínio (MVP)"
    status: todo
    priority: P1
    goal: "coverage ≥ 50% em packages/domain"
    acceptance: "test:coverage + baseline"
    links: []
  - id: qa-BL-003
    title: "Checklist QA no PR"
    status: todo
    priority: P2
    links: []
'@ | Set-Content (Join-Path $visions 'qa.backlog.yaml') -Encoding UTF8

    @'
version: 2
agent: test-architect
updated_at: "2026-07-28T00:00:00-03:00"
items:
  - id: test-architect-BL-004
    title: "Plano de coverage: tooling + meta MVP domain"
    status: todo
    priority: P1
    goal: "coverage ≥ 50% em packages/domain primeiro"
    acceptance: "c8 instalado; CI warn"
    links: []
'@ | Set-Content (Join-Path $visions 'test-architect.backlog.yaml') -Encoding UTF8

    & $Pwsh -NoProfile -ExecutionPolicy Bypass -File $Script `
        -RepoRoot $tmp -SliceId E1-S4 `
        -Suggestions 'quero coverage no domain' `
        -VisionDir 'docs/_arah/visions' `
        -OutDir 'docs/_arah/slice-plans' `
        -Force
    Assert-True ($LASTEXITCODE -eq 0) 'exit 0'
    $out = Join-Path $plans 'E1-S4.md'
    Assert-True (Test-Path -LiteralPath $out) 'wrote E1-S4.md'
    $md = Get-Content -LiteralPath $out -Raw
    Assert-True ($md -match '## Product') 'has Product section'
    Assert-True ($md -match '## Agent priorities pulled') 'has Agent priorities'
    Assert-True ($md -match '## User suggestions') 'has User suggestions'
    Assert-True ($md -match '## Deferred') 'has Deferred'
    Assert-True ($md -match '## Executor') 'has Executor'
    Assert-True ($md -match 'IndexedDB') 'parsed product title'
    Assert-True ($md -match 'qa-BL-005') 'pulled qa coverage item'
    Assert-True ($md -match 'coverage') 'mentions coverage'
    Assert-True ($md -match 'quero coverage') 'echoes user suggestion'

    & $Pwsh -NoProfile -ExecutionPolicy Bypass -File $Script `
        -RepoRoot $tmp -SliceId E1-S4 -DryRun
    Assert-True ($LASTEXITCODE -eq 0) 'dry-run exit 0'
}
finally {
    Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

if ($Fail -gt 0) {
    Write-Host "FAILED: $Fail assertion(s)"
    exit 1
}
Write-Host 'ALL OK'
exit 0

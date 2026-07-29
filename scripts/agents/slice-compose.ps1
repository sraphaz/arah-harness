#Requires -Version 5.1
<#
.SYNOPSIS
  Compose a product slice plan from product backlog ∩ agent vision backlogs ∩ user suggestions.

.DESCRIPTION
  Experimental — fatia composta (não é entrega). Lê:
    - story de produto (ex. E1-S4 em docs/BACKLOG*.md)
    - sidecars *.backlog.yaml em VisionDir
    - sugestões opcionais do usuário
  e grava docs/_arah/slice-plans/<SliceId>.md com seções:
    Product | Agent priorities pulled | User suggestions | Deferred | Executor/consultants

  Prioriza itens agent com status todo e priority P0/P1 (lançamento),
  e itens com goal/acceptance mensuráveis quando cabem na fatia.

.EXAMPLE
  ./slice-compose.ps1 -SliceId E1-S4 -RepoRoot .
  ./slice-compose.ps1 -SliceId E1-S4 -Suggestions "quero coverage no domain" -Executor backend
  ./slice-compose.ps1 -SliceId E1-S4 -DryRun
#>
param(
    [Parameter(Mandatory = $true)]
    [string]$SliceId,

    [string]$RepoRoot = '.',
    [string]$VisionDir = 'docs/_arah/visions',
    [string]$OutDir = 'docs/_arah/slice-plans',
    [string]$ProductBacklog = '',
    [string]$Suggestions = '',
    [string]$SuggestionsFile = '',
    [string]$Agents = '',
    [string]$Executor = 'backend',
    [string]$Consultants = 'solutions-architect,qa',
    [string]$Title = '',
    [switch]$DryRun,
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
$schemaVersion = 1
$stamp = Get-Date -Format 'yyyy-MM-ddTHH:mm:ssK'

$RepoRoot = (Resolve-Path -LiteralPath $RepoRoot).Path
$visionPath = if ([System.IO.Path]::IsPathRooted($VisionDir)) { $VisionDir } else { Join-Path $RepoRoot $VisionDir }
$outPath = if ([System.IO.Path]::IsPathRooted($OutDir)) { $OutDir } else { Join-Path $RepoRoot $OutDir }

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
function Get-SuggestionList {
    $list = @()
    if ($Suggestions) {
        $list += @($Suggestions -split '[\r\n;|]+' | ForEach-Object { $_.Trim() } | Where-Object { $_ })
    }
    if ($SuggestionsFile) {
        $sf = if ([System.IO.Path]::IsPathRooted($SuggestionsFile)) { $SuggestionsFile } else { Join-Path $RepoRoot $SuggestionsFile }
        if (Test-Path -LiteralPath $sf) {
            $list += @(Get-Content -LiteralPath $sf | ForEach-Object { $_.Trim() } | Where-Object { $_ -and $_ -notmatch '^\s*#' })
        }
    }
    return @($list | Select-Object -Unique)
}

function Find-ProductBacklogFile {
    if ($ProductBacklog) {
        $p = if ([System.IO.Path]::IsPathRooted($ProductBacklog)) { $ProductBacklog } else { Join-Path $RepoRoot $ProductBacklog }
        if (Test-Path -LiteralPath $p) { return $p }
        Write-Error "ProductBacklog not found: $p"
    }
    $candidates = @(
        'docs/BACKLOG_HUB_PRODUCTION.md'
        'docs/BACKLOG.md'
        'docs/ROADMAP.md'
        'BACKLOG.md'
        'ROADMAP.md'
    )
    foreach ($c in $candidates) {
        $full = Join-Path $RepoRoot $c
        if (Test-Path -LiteralPath $full) { return $full }
    }
    return $null
}

function Parse-ProductStory {
    param([string]$Path, [string]$Id)
    if (-not $Path) {
        return @{
            id          = $Id
            title       = $(if ($Title) { $Title } else { $Id })
            status      = 'unknown'
            priority    = ''
            depends     = ''
            summary     = ''
            ready_when  = ''
            raw_excerpt = ''
            found       = $false
        }
    }
    $raw = Get-Content -LiteralPath $Path -Raw
    # Match ### E1-S4 — Title ... until next ### or ##
    $pattern = "(?ms)^###\s+$([regex]::Escape($Id))\s*[—\-–:]\s*(?<title>[^\r\n]+)\s*(?<body>.*?)(?=^###\s+|^##\s+|\z)"
    $m = [regex]::Match($raw, $pattern)
    if (-not $m.Success) {
        return @{
            id          = $Id
            title       = $(if ($Title) { $Title } else { $Id })
            status      = 'unknown'
            priority    = ''
            depends     = ''
            summary     = ''
            ready_when  = ''
            raw_excerpt = ''
            found       = $false
            source      = $Path
        }
    }
    $body = $m.Groups['body'].Value
    $storyTitle = $m.Groups['title'].Value.Trim()
    $priority = ''
    $status = ''
    $depends = ''
    $summary = ''
    $ready = ''
    if ($body -match '(?m)\*\*P(?<p>\d)\*\*') { $priority = "P$($Matches['p'])" }
    if ($body -match '(?m)Status:\s*`?(?<s>\w+)`?') { $status = $Matches['s'] }
    if ($body -match '(?m)Depende:\s*(?<d>[^\r\n*]+)') { $depends = $Matches['d'].Trim() }
    if ($body -match '(?ms)\*\*Em uma frase:\*\*\s*(?<s>[^\r\n]+)') { $summary = $Matches['s'].Trim() }
    if ($body -match '(?ms)\*\*Pronto quando:\*\*\s*(?<s>.+?)(?=\r?\n\r?\n|\r?\n###|\z)') {
        $ready = ($Matches['s'] -replace '\s+', ' ').Trim()
    }
    return @{
        id          = $Id
        title       = $(if ($Title) { $Title } else { $storyTitle })
        status      = $status
        priority    = $priority
        depends     = $depends
        summary     = $summary
        ready_when  = $ready
        raw_excerpt = ($body.Trim() -split "`n" | Select-Object -First 12) -join "`n"
        found       = $true
        source      = $Path
    }
}

function Read-AgentBacklog {
    param([string]$Path)
    if (-not (Test-Path -LiteralPath $Path)) { return @() }
    $raw = Get-Content -LiteralPath $Path -Raw
    $agent = ''
    if ($raw -match '(?m)^agent:\s*(?<a>\S+)') { $agent = $Matches['a'] }
    $items = @()
    $blocks = [regex]::Split($raw, '(?m)^\s*-\s+id:\s*')
    foreach ($b in $blocks) {
        if ($b -notmatch '(?m)^(?<id>[A-Za-z0-9_\-]+)') { continue }
        $id = $Matches['id'].Trim()
        if ($id -notmatch '-BL-\d+') { continue }
        $title = ''; $status = 'todo'; $priority = 'P2'; $goal = ''; $acceptance = ''; $links = ''
        if ($b -match '(?m)^\s*title:\s*"?(?<t>.+?)"?\s*$') { $title = $Matches['t'].Trim().Trim('"') }
        if ($b -match '(?m)^\s*status:\s*(?<s>\w+)') { $status = $Matches['s'].Trim() }
        if ($b -match '(?m)^\s*priority:\s*(?<p>\S+)') { $priority = $Matches['p'].Trim() }
        if ($b -match '(?m)^\s*goal:\s*"?(?<g>.+?)"?\s*$') { $goal = $Matches['g'].Trim().Trim('"') }
        if ($b -match '(?m)^\s*acceptance:\s*"?(?<a>.+?)"?\s*$') { $acceptance = $Matches['a'].Trim().Trim('"') }
        if ($b -match '(?m)^\s*links:\s*\[(?<l>[^\]]*)\]') { $links = $Matches['l'].Trim() }
        $items += [pscustomobject]@{
            agent      = $agent
            id         = $id
            title      = $title
            status     = $status
            priority   = $priority
            goal       = $goal
            acceptance = $acceptance
            links      = $links
            file       = $Path
        }
    }
    return $items
}

function Test-LaunchViable {
    param($Item)
    if ($Item.status -notin @('todo', 'doing')) { return $false }
    if ($Item.priority -in @('P0', 'P1')) { return $true }
    if ($Item.goal) { return $true }
    return $false
}

function Test-FitsSliceHeuristic {
    param($Item, [string]$Slice, [string]$ProductTitle, [string[]]$UserHints)
    $hay = (@($Item.title, $Item.goal, $Item.acceptance, $ProductTitle, $Slice) -join ' ').ToLowerInvariant()
    $hints = @()
    if ($UserHints) {
        $hints = @($UserHints | Where-Object { $_ } | ForEach-Object { "$_".ToLowerInvariant() })
    }
    # Strong pull: user mentioned keywords present in item
    foreach ($h in $hints) {
        $tokens = @($h -split '\s+' | Where-Object { $_.Length -ge 4 })
        foreach ($t in $tokens) {
            if ($hay.Contains($t)) { return 'pull' }
        }
    }
    # Domain/coverage/ports/adapter heuristics aligned to strangler slices
    $sliceLower = $Slice.ToLowerInvariant()
    $productLower = "$ProductTitle $Slice".ToLowerInvariant()
    if ($Item.goal -match '(?i)coverage|cobertura|≥\s*\d+%|>=\s*\d+%') {
        if ($productLower -match 'domain|ports|indexeddb|adapter|e1') { return 'pull-if-fits' }
    }
    if ($Item.title -match '(?i)coverage|c8|nyc|test:coverage|pirâmide|TEA|matriz') {
        if ($productLower -match 'domain|ports|indexeddb|adapter|e1|test') { return 'pull-if-fits' }
    }
    if ($Item.priority -eq 'P0') { return 'pull' }
    if ($Item.priority -eq 'P1' -and $sliceLower -match 'e1') { return 'pull-if-fits' }
    return 'defer'
}

# ---------------------------------------------------------------------------
# Load inputs
# ---------------------------------------------------------------------------
$productFile = Find-ProductBacklogFile
$story = Parse-ProductStory -Path $productFile -Id $SliceId
$userHints = Get-SuggestionList

$filterAgents = @()
if ($Agents) {
    $filterAgents = @($Agents -split '[,;\s]+' | Where-Object { $_ })
}

$allItems = @()
if (Test-Path -LiteralPath $visionPath) {
    $files = Get-ChildItem -LiteralPath $visionPath -Filter '*.backlog.yaml' -File -ErrorAction SilentlyContinue
    foreach ($f in $files) {
        $agentId = $f.BaseName -replace '\.backlog$', ''
        if ($filterAgents.Count -gt 0 -and $agentId -notin $filterAgents) { continue }
        $allItems += @(Read-AgentBacklog -Path $f.FullName)
    }
} else {
    Write-Warning "VisionDir not found: $visionPath — agent priorities will be empty"
}

$pulled = @()
$deferred = @()
foreach ($it in $allItems) {
    if (-not (Test-LaunchViable -Item $it)) {
        if ($it.status -in @('todo', 'doing')) {
            $deferred += [pscustomobject]@{ item = $it; reason = "priority $($it.priority) / status $($it.status) — pós-lançamento ou não acionável agora" }
        }
        continue
    }
    $fit = Test-FitsSliceHeuristic -Item $it -Slice $SliceId -ProductTitle $story.title -UserHints $userHints
    if ($fit -eq 'pull') {
        $pulled += [pscustomobject]@{ item = $it; fit = 'pull'; note = 'Prioridade de lançamento / match com sugestão do usuário' }
    } elseif ($fit -eq 'pull-if-fits') {
        $pulled += [pscustomobject]@{ item = $it; fit = 'pull-if-fits'; note = 'Candidato se couber na fatia (senão fatia irmã)' }
    } else {
        $deferred += [pscustomobject]@{ item = $it; reason = 'Não alinhado a esta fatia — adiar' }
    }
}

# Sort pulled: pull first, then P0/P1, then with goal
$pulled = @($pulled | Sort-Object {
        if ($_.fit -eq 'pull') { 0 } else { 1 }
    }, {
        switch ($_.item.priority) { 'P0' { 0 } 'P1' { 1 } default { 2 } }
    }, { $_.item.id })

$consultantList = @($Consultants -split '[,;\s]+' | Where-Object { $_ })
$safeId = ($SliceId -replace '[^\w\-]+', '-')
$outFile = Join-Path $outPath "$safeId.md"

# ---------------------------------------------------------------------------
# Render
# ---------------------------------------------------------------------------
function Format-AgentRow {
    param($Entry)
    $it = $Entry.item
    $goal = if ($it.goal) { $it.goal } else { '—' }
    $acc = if ($it.acceptance) { $it.acceptance } else { '—' }
    $action = if ($Entry.fit -eq 'pull') { '**Puxar**' } else { '**Se couber** / irmã' }
    return "| ``$($it.agent)`` | ``$($it.id)`` | $($it.priority) | $($it.title) | $goal | $acc | $action |"
}

$productSource = if ($story.source) { $story.source } else { '(não encontrado)' }
$foundLabel = if ($story.found) { 'sim' } else { 'não — preencha manualmente' }
$userBlock = if ($userHints.Count -eq 0) {
    '_Nenhuma sugestão do usuário nesta composição._'
} else {
    ($userHints | ForEach-Object { "- $_" }) -join "`n"
}

$pulledBlock = if ($pulled.Count -eq 0) {
    '_Nenhum item agent-BL puxado automaticamente — revise backlogs em VisionDir._'
} else {
    $hdr = @(
        '| Agent | ID | Pri | Title | Goal | Acceptance | Nesta fatia |'
        '|-------|----|-----|-------|------|------------|-------------|'
    )
    $rows = @($pulled | ForEach-Object { Format-AgentRow $_ })
    ($hdr + $rows) -join "`n"
}

$deferredBlock = if ($deferred.Count -eq 0) {
    '_Nada adiado a partir dos backlogs filtrados._'
} else {
    $lines = @($deferred | Select-Object -First 25 | ForEach-Object {
            "- ``$($_.item.id)`` ($($_.item.agent) · $($_.item.priority)): $($_.item.title) — _$($_.reason)_"
        })
    if ($deferred.Count -gt 25) { $lines += "- _… +$($deferred.Count - 25) itens_" }
    $lines -join "`n"
}

$consultantRows = @($consultantList | ForEach-Object { "| **Consulta** | ``$_`` | Parecer limitado ao mandato; não executa |" })
$md = @"
# Slice plan — $SliceId$(if ($story.title -and $story.title -ne $SliceId) { " · $($story.title)" } else { '' })

> **Experimental** · ARAH Slice Compose v$schemaVersion  
> Gerado em $stamp · ``arah slice plan`` / skill ``slice-compose``  
> Fatia **composta**: backlog produto ∩ prioridades agents ∩ sugestões do usuário.

| Campo | Valor |
|-------|-------|
| Slice | ``$SliceId`` |
| Story encontrada | $foundLabel |
| Fonte produto | ``$productSource`` |
| VisionDir | ``$visionPath`` |
| Executor sugerido | ``$Executor`` |

---

## Product

| | |
|--|--|
| **ID** | ``$SliceId`` |
| **Título** | $($story.title) |
| **Priority** | $(if ($story.priority) { $story.priority } else { '—' }) |
| **Status** | $(if ($story.status) { $story.status } else { '—' }) |
| **Depende** | $(if ($story.depends) { $story.depends } else { '—' }) |
| **Em uma frase** | $(if ($story.summary) { $story.summary } else { '—' }) |
| **Pronto quando** | $(if ($story.ready_when) { $story.ready_when } else { '—' }) |

DoD desta fatia (produto): entregar o story acima **sem** expandir para itens listados em Deferred.

---

## Agent priorities pulled

Itens ``todo``/``doing`` com prioridade de **lançamento** (P0/P1) ou ``goal``/``acceptance`` mensurável, alinhados à fatia ou às sugestões do usuário.

$pulledBlock

Convenção: **Puxar** = incluir no escopo da fatia se o executor concordar; **Se couber** = mesma fatia *ou* fatia irmã (não bloquear o DoD de produto).

---

## User suggestions

$userBlock

---

## Deferred

$deferredBlock

---

## Executor / consultants

| Papel | Agent | Faz |
|-------|-------|-----|
| **Executor** | ``$Executor`` | Entrega o DoD de produto + itens **Puxar** acordados; evidência ECP |
$($consultantRows -join "`n")
| **Review** | ``qa`` / ``pr-steward`` | Evidência; não redefine executor |

Orchestrator: stood down após routing (Execution Control).

---

## Como usar na próxima interação (“outra slice”)

1. Agents leem o próprio backlog (``*.backlog.yaml``) e este plano.
2. Usuário pode acrescentar sugestões (``-Suggestions`` / arquivo).
3. Rodar de novo: ``arah slice plan -SliceId <próxima>`` — compõe Product ∩ Agent-BL ∩ User.
4. Abrir tarefa ECP com **um** executor: ``arah task create -Objective "…"``.

---
_Artefato gerado por ``arah slice plan`` · não é entrega de produto_
"@

if ($DryRun) {
    Write-Host "dry-run → $outFile"
    Write-Host "  product found: $($story.found) · pulled: $($pulled.Count) · deferred: $($deferred.Count) · suggestions: $($userHints.Count)"
    Write-Host $md
    exit 0
}

if (-not (Test-Path -LiteralPath $outPath)) {
    New-Item -ItemType Directory -Path $outPath -Force | Out-Null
}
if ((Test-Path -LiteralPath $outFile) -and -not $Force) {
    Write-Warning "Já existe $outFile — use -Force para sobrescrever"
}
Set-Content -LiteralPath $outFile -Value $md -Encoding UTF8
Write-Host "slice-compose → $outFile (pulled=$($pulled.Count) deferred=$($deferred.Count) suggestions=$($userHints.Count))"
exit 0

#Requires -Version 5.1
<#
.SYNOPSIS
  Assessment experimental: cada agent observa o repo na sua perspectiva
  e grava opinião + As-Is + Gaps + Action Plan + Backlog + Memory em .arah/visions/.
.DESCRIPTION
  Fase de bootstrap controlada (exceção documentada ao modelo 1-executor):
  pareceres em série, um por agent, sem handoff consultor→consultor.
  Não altera código de produto. Não faz merge. Opt-in pós-install.

  Schema v2 por agent:
    - opinion (perspectiva no tipo de aplicação + lente Clean Architecture)
    - as_is / gaps / action_plan
    - backlog (YAML sidecar + tabela MD; IDs estáveis {agent}-BL-NNN)
    - memory/events (append-only)

  -Refresh / vision update / backlog sync: regenera narrativa e mescla backlog
  (preserva status done/cancelled) + registra evento.
.EXAMPLE
  ./assess-repo.ps1
  ./assess-repo.ps1 -OutDir docs/_arah/visions -Agents qa,backend,security
  ./assess-repo.ps1 -Refresh
  ./assess-repo.ps1 -DryRun
#>
param(
    [string]$RepoRoot = '',
    [string]$OutDir = '.arah/visions',
    [string]$Agents = '',
    [switch]$IncludeDomain,
    [switch]$IncludeSpecialists,
    [switch]$Force,
    [switch]$Refresh,
    [switch]$BacklogOnly,
    [switch]$DryRun,
    [switch]$SkipIndex
)

$ErrorActionPreference = 'Stop'

if (-not $RepoRoot) {
    $RepoRoot = (Get-Location).Path
}
$RepoRoot = (Resolve-Path -LiteralPath $RepoRoot).Path

. (Join-Path $PSScriptRoot 'yaml-lite.ps1')

# Refresh implica atualização (merge); BacklogOnly implica Refresh
if ($BacklogOnly) { $Refresh = $true }
if ($Refresh) { $Force = $true }

$stamp = Get-Date -Format 'yyyy-MM-ddTHH:mm:ssK'
$schemaVersion = 2
$harnessVersion = 'unknown'
$verFile = Join-Path $RepoRoot '.arah-version'
if (Test-Path -LiteralPath $verFile) {
    $vr = Get-Content -LiteralPath $verFile -Raw
    if ($vr -match '(?m)^version:\s*(.+)$') { $harnessVersion = $Matches[1].Trim() }
}
$hvFile = Join-Path (Split-Path $PSScriptRoot -Parent | Split-Path -Parent) 'VERSION'
if ($harnessVersion -eq 'unknown' -and (Test-Path -LiteralPath $hvFile)) {
    $harnessVersion = (Get-Content -LiteralPath $hvFile -Raw).Trim()
}

# ---------------------------------------------------------------------------
# Repo snapshot (shared evidence for all lenses)
# ---------------------------------------------------------------------------
function Test-RepoPath {
    param([string]$Rel)
    return Test-Path -LiteralPath (Join-Path $RepoRoot $Rel)
}

function Get-TopDirs {
    Get-ChildItem $RepoRoot -Directory -Force -ErrorAction SilentlyContinue | Where-Object {
        $_.Name -notmatch '^\.(git|cursor|arah)$' -and
        $_.Name -notin @('node_modules', 'dist', 'build', 'coverage', '.venv', 'vendor', '.next', 'out')
    } | ForEach-Object { $_.Name }
}

function Count-Files {
    param([string[]]$Patterns)
    $n = 0
    foreach ($pat in $Patterns) {
        $n += @(Get-ChildItem -Path $RepoRoot -Recurse -File -Filter $pat -ErrorAction SilentlyContinue |
            Where-Object {
                $_.FullName -notmatch '[\\/](node_modules|\.git|dist|build|coverage|\.venv|vendor)[\\/]'
            }).Count
    }
    return $n
}

$topDirs = @(Get-TopDirs)
$languages = New-Object System.Collections.Generic.List[string]
$frameworks = New-Object System.Collections.Generic.List[string]
$evidence = New-Object System.Collections.Generic.List[string]

$stackChecks = @(
    @{ file = 'package.json'; lang = 'javascript' },
    @{ file = 'pnpm-lock.yaml'; lang = 'javascript' },
    @{ file = 'pyproject.toml'; lang = 'python' },
    @{ file = 'requirements.txt'; lang = 'python' },
    @{ file = 'go.mod'; lang = 'go' },
    @{ file = 'Cargo.toml'; lang = 'rust' },
    @{ file = 'pom.xml'; lang = 'java' },
    @{ file = 'build.gradle'; lang = 'java' },
    @{ file = 'composer.json'; lang = 'php' }
)
foreach ($c in $stackChecks) {
    if (Test-RepoPath $c.file) {
        if ($c.lang -notin $languages) { [void]$languages.Add($c.lang) }
        [void]$evidence.Add($c.file)
    }
}

if (Test-RepoPath 'package.json') {
    $pkg = Get-Content (Join-Path $RepoRoot 'package.json') -Raw
    $fwMap = @{
        '"next"' = 'nextjs'; '"react"' = 'react'; '"vue"' = 'vue'
        '"@nestjs/core"' = 'nestjs'; '"express"' = 'express'
        '"vite"' = 'vite'; '"playwright"' = 'playwright'
    }
    foreach ($k in $fwMap.Keys) {
        if ($pkg -match [regex]::Escape($k) -and $fwMap[$k] -notin $frameworks) {
            [void]$frameworks.Add($fwMap[$k])
        }
    }
}

$hasTests = (Test-RepoPath 'tests') -or (Test-RepoPath 'test') -or (Test-RepoPath '__tests__') -or (Test-RepoPath 'e2e') -or (Test-RepoPath 'spec')
$hasCi = (Test-RepoPath '.github/workflows') -or (Test-RepoPath '.gitlab-ci.yml') -or (Test-RepoPath 'azure-pipelines.yml')
$hasDocs = Test-RepoPath 'docs'
$hasAgents = Test-RepoPath '.agents'
$hasArahConfig = Test-RepoPath 'arah.config.yaml'
$hasReadme = (Test-RepoPath 'README.md') -or (Test-RepoPath 'README')
$hasAdr = (Test-RepoPath 'docs/adr') -or (Test-RepoPath 'docs/architecture')
$hasSpecs = Test-RepoPath 'docs/specs'
$hasSecurityFiles = (Test-RepoPath 'SECURITY.md') -or (Test-RepoPath '.github/dependabot.yml') -or (Test-RepoPath '.snyk')
$hasEnvExample = (Test-RepoPath '.env.example') -or (Test-RepoPath '.env.sample')
$hasPackages = (Test-RepoPath 'packages') -or ($topDirs -contains 'packages')
$hasAndroid = ($topDirs -contains 'android') -or (Test-RepoPath 'android')
$hasDomainPkg = (Test-RepoPath 'packages/domain') -or (Test-RepoPath 'domain.js') -or ($topDirs -contains 'domain')
$srcFileCount = Count-Files @('*.js', '*.ts', '*.tsx', '*.py', '*.go', '*.rs', '*.java', '*.cs', '*.cjs', '*.mjs')
$testFileCount = Count-Files @('*.test.js', '*.test.ts', '*.test.cjs', '*.spec.js', '*.spec.ts', '*_test.go', 'test_*.py')
$docFileCount = @(Get-ChildItem (Join-Path $RepoRoot 'docs') -Recurse -File -ErrorAction SilentlyContinue |
    Where-Object { $_.Extension -match '\.(md|yaml|yml|mdx)$' }).Count

$trackedish = @(Get-ChildItem $RepoRoot -Force -ErrorAction SilentlyContinue | Where-Object {
    $_.Name -notin @('.git', 'node_modules', '.venv', 'dist', 'build', 'coverage', 'vendor')
}).Count
$isEmptyish = ($srcFileCount -eq 0) -and ($trackedish -le 8) -and (-not (Test-RepoPath 'package.json')) -and (-not (Test-RepoPath 'go.mod'))

# App-type heuristic (feeds opinion per role)
$appType = 'unknown'
if ($isEmptyish) { $appType = 'bootstrap-empty' }
elseif ($hasAndroid -and $hasPackages) { $appType = 'multi-surface-product' }
elseif ($frameworks -contains 'nextjs' -or $frameworks -contains 'react') { $appType = 'web-frontend-heavy' }
elseif ($frameworks -contains 'nestjs' -or ($topDirs -contains 'api') -or ($topDirs -contains 'backend')) { $appType = 'api-backend' }
elseif ($hasDomainPkg -and $hasDocs) { $appType = 'domain-centric-monorepo' }
elseif ($hasDocs -and $srcFileCount -lt 5) { $appType = 'docs-first' }
else { $appType = 'general-codebase' }

$snapshot = [ordered]@{
    emptyish          = $isEmptyish
    app_type          = $appType
    languages         = @($languages)
    frameworks        = @($frameworks)
    top_dirs          = $topDirs
    evidence          = @($evidence)
    has_tests         = $hasTests
    has_ci            = $hasCi
    has_docs          = $hasDocs
    has_agents        = $hasAgents
    has_arah_config   = $hasArahConfig
    has_readme        = $hasReadme
    has_adr           = $hasAdr
    has_specs         = $hasSpecs
    has_security      = $hasSecurityFiles
    has_env_example   = $hasEnvExample
    has_domain_pkg    = $hasDomainPkg
    src_files         = $srcFileCount
    test_files        = $testFileCount
    doc_files         = $docFileCount
}

function Get-AppTypeBlurb {
    switch ($appType) {
        'bootstrap-empty' { 'repositório vazio/quase vazio — visão de bootstrap, não auditoria de legado' }
        'multi-surface-product' { 'produto multi-superfície (ex.: domínio compartilhado + Android/PWA/hub) — fronteiras e contratos importam mais que pastas' }
        'web-frontend-heavy' { 'aplicação web (React/Next) — UI como adapter; domínio não deve vazar para componentes' }
        'api-backend' { 'API/backend — ports & adapters, casos de uso testáveis sem HTTP' }
        'domain-centric-monorepo' { 'monorepo centrado em domínio — Clean Architecture / strangler progressivo' }
        'docs-first' { 'docs-first — intenção documentada à frente do código' }
        default { 'codebase geral — aplicar lentes de papel com proporcionalidade' }
    }
}

# ---------------------------------------------------------------------------
# Perspective lenses (role → opinion CA + signals + action_plan + backlog seeds)
# ---------------------------------------------------------------------------
$lenses = @{
    'qa' = @{
        headline = 'o que é este repositório na cabeça do teste / QA?'
        ca_lens  = 'Pirâmide TEA: unit no domínio (núcleo), integração nos ports, e2e só nos fluxos críticos dos adapters.'
        opinion_seed = 'Como QA, meu escopo não é “achar pastas de teste” — é garantir que o risco do produto vira evidência executável antes do merge. Em apps multi-superfície, cobro contrato domínio↔UI↔device; em bootstrap, cobro DoD mínimo e smoke.'
        questions = @(
            'Há pirâmide de testes observável (unit / integração / e2e)?'
            'CI cobre o caminho crítico antes do merge?'
            'Specs e Definition of Done estão amarrados a evidência de teste?'
            'Regressões conhecidas têm dono e gate?'
        )
        signals = @(
            @{ id = 'tests-dir'; ok = $hasTests; asis = 'Diretório/padrão de testes presente'; gap = 'Sem pasta ou convenção de testes detectável' }
            @{ id = 'test-files'; ok = ($testFileCount -gt 0); asis = "$testFileCount arquivo(s) de teste detectados"; gap = 'Nenhum arquivo *test* / *spec* encontrado' }
            @{ id = 'ci'; ok = $hasCi; asis = 'Workflows CI presentes'; gap = 'Sem CI detectável (.github/workflows etc.)' }
            @{ id = 'specs'; ok = $hasSpecs; asis = 'docs/specs presente'; gap = 'Sem docs/specs — SDD frágil para QA' }
        )
        action_plan = @(
            'Matriz de risco → casos críticos com gate no PR'
            'Comando `tests.all` estável no arah.config.yaml'
            'Checklist QA amarrado a evidência (não só review textual)'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Matriz de risco → casos críticos com gate no PR'; priority = 'P1' }
            @{ n = 2; title = 'Estabilizar comando tests.all no arah.config.yaml'; priority = 'P1' }
            @{ n = 3; title = 'Checklist QA com evidência obrigatória no PR template'; priority = 'P2' }
            @{ n = 4; title = 'Mapear pirâmide TEA (unit domínio / integração ports / e2e adapters)'; priority = 'P2' }
        )
    }
    'test-architect' = @{
        headline = 'o que é este repositório na cabeça da arquitetura de testes?'
        ca_lens  = 'TEA risk-based: NFRs e classes de risco definem onde investir na pirâmide; adapters externos ficam na borda.'
        opinion_seed = 'Como Test Architect, desenho a estratégia — não executo checklist genérico. Em produto multi-surface, priorizo contratos e regressão de domínio; e2e caro só onde o risco de integração justifica.'
        questions = @(
            'Existe estratégia risk-based documentada?'
            'NFRs (a11y, performance, segurança) têm critérios mensuráveis?'
            'A pirâmide está alinhada ao risco do domínio?'
        )
        signals = @(
            @{ id = 'testing-docs'; ok = (Test-RepoPath 'docs/testing') -or (Test-RepoPath 'docs/TEST_MATRIX.md'); asis = 'Documentação de testes/estratégia presente'; gap = 'Sem docs/testing ou TEST_MATRIX' }
            @{ id = 'tests'; ok = $hasTests; asis = 'Harness de testes detectado'; gap = 'Sem base de testes para derivar pirâmide' }
            @{ id = 'specs'; ok = $hasSpecs; asis = 'Specs disponíveis como input TEA'; gap = 'Sem specs — TEA opera no vazio' }
        )
        action_plan = @(
            'docs/testing/test-strategy.md com riscos e pirâmide'
            'Gates CI por classe de risco'
            'Critérios a11y/perf explícitos onde o produto exigir'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Escrever test-strategy.md (riscos + pirâmide TEA)'; priority = 'P1' }
            @{ n = 2; title = 'Definir gates CI por classe de risco'; priority = 'P1' }
            @{ n = 3; title = 'Critérios NFR a11y/perf mensuráveis'; priority = 'P2' }
        )
    }
    'solutions-architect' = @{
        headline = 'o que é este repositório na cabeça da arquitetura?'
        ca_lens  = 'Boundaries Clean Architecture: entidades/casos de uso no centro; UI, DB, IoT e billing como adapters; ADRs congelam decisões.'
        opinion_seed = 'Como SA, meu escopo é fronteiras e evolução controlada — não inventário de pastas. Em multi-surface, cobro mapa de bounded contexts e contratos públicos; sem ADR, toda mudança estrutural é dívida silenciosa.'
        questions = @(
            'Há fronteiras claras (módulos, apps, contratos)?'
            'Decisões estruturais estão em ADRs?'
            'O desenho alvo (To-Be) está documentado vs o As-Is?'
            'Dependências apontam para dentro (DIP) entre packages?'
        )
        signals = @(
            @{ id = 'adr'; ok = $hasAdr; asis = 'docs/adr ou docs/architecture presente'; gap = 'Sem ADRs / docs de arquitetura' }
            @{ id = 'structure'; ok = ($topDirs.Count -ge 2); asis = ("Top-level: " + ($topDirs -join ', ')); gap = 'Estrutura flat ou quase vazia — fronteiras ainda implícitas' }
            @{ id = 'domain-pkg'; ok = $hasDomainPkg; asis = 'Pacote/módulo de domínio detectado'; gap = 'Domínio não isolado como package/módulo' }
            @{ id = 'specs'; ok = $hasSpecs; asis = 'Specs SDD presentes'; gap = 'Sem specs — arquitetura sem âncora de requisito' }
            @{ id = 'readme'; ok = $hasReadme; asis = 'README presente'; gap = 'Sem README — onboarding arquitetural frágil' }
        )
        action_plan = @(
            'Mapa de contextos/limites versionado (C4 ou equivalente)'
            'ADRs para decisões estruturais (strangler, packages, hub)'
            'Contratos públicos (OpenAPI/events) quando houver integração'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Mapa de bounded contexts / limites de package'; priority = 'P1' }
            @{ n = 2; title = 'ADR inicial: strangler e fronteira domínio×adapters'; priority = 'P1' }
            @{ n = 3; title = 'Inventário de contratos públicos (API/events/schemas)'; priority = 'P2' }
            @{ n = 4; title = 'Diagrama C4 L1/L2 versionado'; priority = 'P2' }
        )
    }
    'backend' = @{
        headline = 'o que é este repositório na cabeça do developer/backend?'
        ca_lens  = 'Ports & adapters: casos de uso no domínio; HTTP/DB/filas só na borda; testes de domínio sem subir servidor.'
        opinion_seed = 'Como backend, meu escopo é lógica de negócio e ports — não a árvore inteira. Em produto com PWA/APK/hub, cobro contratos estáveis e testes de domínio no CI; pasta “backend/” é opcional se packages/domain existir.'
        questions = @(
            'Onde mora a lógica de domínio e a API?'
            'Há testes de domínio/API executáveis?'
            'Há caminho claro de delivery (scripts, CI)?'
            'Ports estão explícitos (interfaces) ou implícitos (imports diretos)?'
        )
        signals = @(
            @{ id = 'src'; ok = ($srcFileCount -gt 0); asis = "$srcFileCount arquivo(s) de código detectados"; gap = 'Repo sem código — bootstrap de backend necessário' }
            @{ id = 'api-shape'; ok = (($topDirs -contains 'backend') -or ($topDirs -contains 'api') -or ($topDirs -contains 'services') -or ($topDirs -contains 'apps') -or (Test-RepoPath 'domain.js') -or $hasPackages); asis = 'Há indício de domínio/API/pacotes'; gap = 'Sem backend/api/services/packages observáveis' }
            @{ id = 'domain-pkg'; ok = $hasDomainPkg; asis = 'Domínio empacotado/isolado'; gap = 'Domínio ainda acoplado a app/UI' }
            @{ id = 'tests'; ok = ($testFileCount -gt 0); asis = 'Testes presentes'; gap = 'Sem testes — regressão de backend sem rede de segurança' }
            @{ id = 'stack'; ok = ($languages.Count -gt 0); asis = ("Stack: " + ($languages -join ', ') + $(if ($frameworks.Count) { ' / ' + ($frameworks -join ', ') } else { '' })); gap = 'Stack não detectada' }
        )
        action_plan = @(
            'Fronteira domínio × app × infra explícita (ports)'
            'Testes de domínio no caminho crítico do CI'
            'Contratos versionados se houver clientes (PWA/APK/hub)'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Expor ports de domínio (interfaces) sem dependência de UI'; priority = 'P1' }
            @{ n = 2; title = 'Testes de domínio no CI do path crítico'; priority = 'P1' }
            @{ n = 3; title = 'Versionar contratos para clientes PWA/APK/hub'; priority = 'P2' }
        )
    }
    'frontend' = @{
        headline = 'o que é este repositório na cabeça do frontend?'
        ca_lens  = 'UI = adapter delivery: consome casos de uso/DTOs; sem regras de negócio em componentes.'
        opinion_seed = 'Como frontend, meu escopo é experiência e contrato com o domínio — não reinventar regras. Em PWA+Android, cobro smoke dos fluxos críticos e a11y mínima; estado de negócio fica fora da árvore de views.'
        questions = @(
            'UI é PWA, app nativo, ou ambos?'
            'Há smoke/e2e para fluxos críticos?'
            'Assets e acessibilidade têm dono?'
        )
        signals = @(
            @{ id = 'ui'; ok = (Test-RepoPath 'index.html') -or (Test-RepoPath 'app.js') -or ($topDirs -contains 'frontend') -or ($topDirs -contains 'web') -or ($topDirs -contains 'apps') -or ($topDirs -contains 'android') -or ($frameworks -contains 'react') -or ($frameworks -contains 'nextjs'); asis = 'Sinais de UI/PWA/frontend/Android detectados'; gap = 'Sem frontend observável' }
            @{ id = 'tests'; ok = ($testFileCount -gt 0) -or ($frameworks -contains 'playwright'); asis = 'Há base de teste (incl. possível e2e)'; gap = 'Sem testes de UI/e2e' }
            @{ id = 'a11y'; ok = $false; asis = ''; gap = 'Sem evidência automática de a11y (heurística conservadora)' }
        )
        action_plan = @(
            'Smoke dos fluxos críticos no CI'
            'Contrato UI ↔ domínio estável'
            'Checklist a11y mínimo no QA'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Smoke e2e dos fluxos críticos no CI'; priority = 'P1' }
            @{ n = 2; title = 'Documentar contrato UI ↔ domínio (DTOs/events)'; priority = 'P2' }
            @{ n = 3; title = 'Checklist a11y mínimo no gate QA'; priority = 'P2' }
        )
    }
    'security' = @{
        headline = 'o que é este repositório na cabeça da segurança?'
        ca_lens  = 'ASVS lite + threat model nas bordas: auth, secrets, dependências e adapters (billing/IoT) — domínio sem secrets.'
        opinion_seed = 'Como security, meu escopo é superfície de ataque e higiene — não “tem pasta secure/”. Em produtos com hub/IoT/billing, cobro threat model leve, Dependabot e gate de secrets; ASVS lite guia o mínimo verificável.'
        questions = @(
            'Secrets e .env estão fora do git?'
            'Dependências têm audit/Dependabot?'
            'Superfície de ataque (auth, pagamento, IoT) está mapeada?'
            'Adapters externos validam input na borda?'
        )
        signals = @(
            @{ id = 'security-md'; ok = $hasSecurityFiles; asis = 'SECURITY.md / Dependabot / scanner presente'; gap = 'Sem política de segurança versionada ou Dependabot' }
            @{ id = 'env-example'; ok = $hasEnvExample; asis = '.env.example presente (bom sinal de higiene)'; gap = 'Sem .env.example — risco de secrets mal documentados' }
            @{ id = 'gitignore'; ok = (Test-RepoPath '.gitignore'); asis = '.gitignore presente'; gap = 'Sem .gitignore' }
            @{ id = 'ci'; ok = $hasCi; asis = 'CI pode hospedar gates de security'; gap = 'Sem CI para gate de security' }
        )
        action_plan = @(
            'Gate de secrets no pre-commit/CI'
            'Dependabot ou equivalente'
            'Threat model leve para auth/billing/IoT se aplicável'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Gate de secrets (pre-commit/CI) + .env.example'; priority = 'P1' }
            @{ n = 2; title = 'Ativar Dependabot / audit de dependências'; priority = 'P1' }
            @{ n = 3; title = 'Threat model leve ASVS: auth, billing, IoT/hub'; priority = 'P1' }
            @{ n = 4; title = 'SECURITY.md com canal de reporte'; priority = 'P2' }
        )
    }
    'docs-steward' = @{
        headline = 'o que é este repositório na cabeça da documentação?'
        ca_lens  = 'Docs como mapa do sistema: PRD/ADR/spec espelham camadas (intenção → decisão → contrato → código).'
        opinion_seed = 'Como docs-steward, meu escopo é navegabilidade e verdade compartilhada — índice canônico e sync doc↔código. Em monorepos grandes, cobro taxonomia estável e donos por artefato.'
        questions = @(
            'Taxonomia de docs é navegável?'
            'README e índices refletem o produto real?'
            'Doc e código divergem?'
        )
        signals = @(
            @{ id = 'docs'; ok = $hasDocs; asis = "docs/ presente ($docFileCount arquivos)"; gap = 'Sem docs/' }
            @{ id = 'readme'; ok = $hasReadme; asis = 'README presente'; gap = 'Sem README' }
            @{ id = 'agents-md'; ok = (Test-RepoPath 'AGENTS.md'); asis = 'AGENTS.md presente'; gap = 'Sem AGENTS.md' }
        )
        action_plan = @(
            'Índice canônico (docs/README ou equivalente)'
            'Sync-docs no PR quando paths de doc mudam'
            'PRD/ADR/spec com donos claros'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Índice canônico docs/README'; priority = 'P2' }
            @{ n = 2; title = 'Donos explícitos PRD/ADR/spec'; priority = 'P2' }
        )
    }
    'planner' = @{
        headline = 'o que é este repositório na cabeça do planejamento?'
        ca_lens  = 'Backlog por valor/risco alinhado a camadas: épicos de domínio vs adapters vs plataforma.'
        opinion_seed = 'Como planner, meu escopo é fatiar trabalho acionável — não listar pastas. Alinho épicos ao ECP e às visões dos papéis; backlog de visão alimenta (não substitui) o roadmap de produto.'
        questions = @(
            'Há backlog / roadmap acionável?'
            'Classes de trabalho (trivial→release) estão claras?'
            'Dependências entre épicos estão explícitas?'
        )
        signals = @(
            @{ id = 'backlog'; ok = (Test-RepoPath 'docs/BACKLOG_HUB_PRODUCTION.md') -or (Test-RepoPath 'docs/ROADMAP.md') -or (Test-RepoPath 'ROADMAP.md'); asis = 'Backlog/roadmap detectado'; gap = 'Sem backlog/roadmap versionado' }
            @{ id = 'prd'; ok = (Test-RepoPath 'docs/prd'); asis = 'docs/prd presente'; gap = 'Sem PRD estruturado' }
        )
        action_plan = @(
            'Backlog com Definition of Ready'
            'Fatiamento por valor + risco'
            'Alinhamento planner ↔ orchestrator (ECP) ↔ visões por agent'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Definition of Ready no backlog de produto'; priority = 'P1' }
            @{ n = 2; title = 'Mapear épicos × visões de agent (feedback loop)'; priority = 'P2' }
        )
    }
    'orchestrator' = @{
        headline = 'o que é este repositório na cabeça do orquestrador?'
        ca_lens  = 'ECP: um executor por mudança; visões são memória de papéis, não handoff livre entre consultores.'
        opinion_seed = 'Como orchestrator, meu escopo é coreografia e terminalidade — visões alimentam briefing, não substituem contrato ECP. Cobro ECP enabled, path-based choreography e refresh de visões após mudanças estruturais.'
        questions = @(
            'Execution Control está ligado?'
            'Coreografia cobre os paths reais?'
            'Há um executor canônico por área?'
        )
        signals = @(
            @{ id = 'arah'; ok = $hasArahConfig; asis = 'arah.config.yaml presente'; gap = 'Sem arah.config.yaml' }
            @{ id = 'agents'; ok = $hasAgents; asis = '.agents/ presente'; gap = 'Sem .agents/' }
            @{ id = 'choreography'; ok = (Test-RepoPath '.agents/choreography.yaml'); asis = 'choreography.yaml presente'; gap = 'Sem choreography.yaml' }
            @{ id = 'ecp'; ok = $hasArahConfig; asis = 'Config ARAH (verificar execution_control.enabled)'; gap = 'ECP não configurável sem arah.config.yaml' }
        )
        action_plan = @(
            'ECP enabled com limites explícitos'
            'Coreografia path-based alinhada à árvore real'
            'Assessment de visões revisada após mudanças estruturais'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Confirmar execution_control.enabled + limites'; priority = 'P1' }
            @{ n = 2; title = 'Alinhar choreography paths à árvore real'; priority = 'P1' }
            @{ n = 3; title = 'Ritual: assess-repo -Refresh após mudanças estruturais'; priority = 'P2' }
        )
    }
    'spec-steward' = @{
        headline = 'o que é este repositório na cabeça do steward de specs?'
        ca_lens  = 'Specs como contratos da camada de aplicação — input de TEA e de implementação; proporção por classe ECP.'
        opinion_seed = 'Como spec-steward, meu escopo é especificação verificável — template, validação e rastreio Spec-Id↔PR. Sem specs, QA e SA operam no vazio.'
        questions = @(
            'Specs existem e validam?'
            'Spec-before-code é proporcional à classe?'
            'Há template e CI de specs?'
        )
        signals = @(
            @{ id = 'specs'; ok = $hasSpecs; asis = 'docs/specs presente'; gap = 'Sem docs/specs' }
            @{ id = 'template'; ok = (Test-RepoPath 'docs/specs/_template.spec.yaml'); asis = 'Template de spec presente'; gap = 'Sem template de spec' }
        )
        action_plan = @(
            'Specs para mudanças standard+'
            'Gate validate-specs no CI'
            'Rastreio Spec-Id ↔ PR'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Gate validate-specs no CI'; priority = 'P1' }
            @{ n = 2; title = 'Rastreio Spec-Id ↔ PR no template'; priority = 'P2' }
        )
    }
    'pr-steward' = @{
        headline = 'o que é este repositório na cabeça do steward de PRs?'
        ca_lens  = 'PR como fronteira de integração: DoD, checks e evidência por camada tocada.'
        opinion_seed = 'Como pr-steward, meu escopo é qualidade do merge — template, protection e resposta a bots. Não redesenho arquitetura; garanto que o gate reflete o DoD dos papéis.'
        questions = @(
            'Template de PR e branch protection existem?'
            'Ready-for-merge é verificável?'
            'Bots de review têm dono de resposta?'
        )
        signals = @(
            @{ id = 'pr-template'; ok = (Test-RepoPath '.github/PULL_REQUEST_TEMPLATE.md'); asis = 'PR template presente'; gap = 'Sem PR template' }
            @{ id = 'ci'; ok = $hasCi; asis = 'CI presente para checks de PR'; gap = 'Sem CI' }
        )
        action_plan = @(
            'Checklist de PR alinhado ao DoD'
            'Branch protection documentada'
            'Fluxo address-bot-review quando houver bots'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Checklist PR alinhado ao DoD + evidência QA'; priority = 'P2' }
            @{ n = 2; title = 'Documentar branch protection'; priority = 'P2' }
        )
    }
    'release' = @{
        headline = 'o que é este repositório na cabeça do release?'
        ca_lens  = 'Release corta artefatos reproduzíveis das bordas; versão e CHANGELOG amarram o núcleo entregue.'
        opinion_seed = 'Como release, meu escopo é corte e reprodutibilidade — não feature work. Cobro VERSION/CHANGELOG, human gate em classe release e checksum quando houver binários.'
        questions = @(
            'Versionamento e CHANGELOG estão vivos?'
            'Artefatos de release são reproduzíveis?'
            'Há human gate para release?'
        )
        signals = @(
            @{ id = 'changelog'; ok = (Test-RepoPath 'CHANGELOG.md'); asis = 'CHANGELOG presente'; gap = 'Sem CHANGELOG' }
            @{ id = 'version'; ok = (Test-RepoPath 'package.json') -or (Test-RepoPath 'VERSION'); asis = 'Fonte de versão detectada'; gap = 'Sem VERSION/package.json' }
            @{ id = 'ci'; ok = $hasCi; asis = 'CI pode publicar artefatos'; gap = 'Sem CI de release' }
        )
        action_plan = @(
            'Release cut documentado (skill release-cut)'
            'Artefatos com checksum quando aplicável'
            'Classe release no ECP com human gate'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Documentar release-cut + human gate ECP'; priority = 'P2' }
            @{ n = 2; title = 'Checksum de artefatos Android/hub se aplicável'; priority = 'P3' }
        )
    }
    'clean-craft-advisor' = @{
        headline = 'o que é este repositório na cabeça do clean craft / Uncle Bob?'
        ca_lens  = 'Clean Architecture clássica: regras de negócio testáveis sem UI; DIP; nomes que revelam intenção.'
        opinion_seed = 'Como clean-craft, meu escopo é craft do núcleo — regras sem frameworks, testes que protegem invariantes, PRs de domínio com craft-review. Em strangler, cobro que o pacote de domínio não importe adapters.'
        questions = @(
            'Há fronteiras de domínio limpas?'
            'Testes protegem regras de negócio?'
            'Acoplamento UI↔domínio está sob controle?'
        )
        signals = @(
            @{ id = 'domain'; ok = $hasDomainPkg; asis = 'Sinais de pacote/módulo de domínio'; gap = 'Domínio não isolado / inexistente' }
            @{ id = 'tests'; ok = ($testFileCount -gt 0); asis = 'Há testes (verificar se cobrem regras)'; gap = 'Sem testes de domínio' }
        )
        action_plan = @(
            'Regras de negócio testáveis sem UI'
            'Dependências apontando para dentro (DIP)'
            'Craft-review nos PRs de domínio'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'Garantir domínio testável sem UI/HTTP'; priority = 'P1' }
            @{ n = 2; title = 'Lint/arch-rule: domínio não importa adapters'; priority = 'P1' }
            @{ n = 3; title = 'Craft-review checklist em PRs de packages/domain'; priority = 'P2' }
        )
    }
    'architecture-documenter' = @{
        headline = 'o que é este repositório na cabeça do documentador de arquitetura?'
        ca_lens  = 'C4 + glossário de contexts: documentar o mapa Clean Architecture que o SA decide.'
        opinion_seed = 'Como architecture-documenter, meu escopo é tornar o desenho legível — C4, glossário, índice de ADRs. Não decido fronteiras; as registro e mantenho vivas.'
        questions = @(
            'Diagramas e glossário existem?'
            'C4 / fluxos estão atualizados?'
            'Decisões estão linkadas a ADRs?'
        )
        signals = @(
            @{ id = 'arch-docs'; ok = $hasAdr; asis = 'Pasta de arquitetura/ADR presente'; gap = 'Sem docs de arquitetura' }
            @{ id = 'docs'; ok = $hasDocs; asis = 'docs/ presente'; gap = 'Sem docs/' }
        )
        action_plan = @(
            'Diagrama C4 (ou equivalente) versionado'
            'Glossário de bounded contexts'
            'ADRs indexados'
        )
        backlog_seeds = @(
            @{ n = 1; title = 'C4 L1/L2 + glossário de contexts'; priority = 'P1' }
            @{ n = 2; title = 'Índice de ADRs'; priority = 'P2' }
        )
    }
}

$defaultLens = @{
    headline = 'o que é este repositório na perspectiva deste agent?'
    ca_lens  = 'Clean Architecture: atuar na camada do mandato; não atravessar boundaries sem contrato.'
    opinion_seed = 'Meu escopo é o mandato do manifesto — observar artefatos do papel, propor gaps e manter backlog próprio sem assumir execução fora do ECP.'
    questions = @(
        'Qual o mandato deste papel neste repo?'
        'Quais artefatos As-Is sustentam o trabalho?'
        'O que falta para o action plan mínimo viável?'
    )
    signals = @(
        @{ id = 'agents'; ok = $hasAgents; asis = 'Harness de agents presente'; gap = 'Sem .agents — papel ainda não operacional' }
        @{ id = 'docs'; ok = $hasDocs; asis = 'Documentação presente'; gap = 'Sem docs' }
        @{ id = 'src'; ok = ($srcFileCount -gt 0); asis = 'Código presente'; gap = 'Repo vazio / sem código — visão de bootstrap' }
    )
    action_plan = @(
        'Definir mandato e paths no manifest'
        'Alinhar coreografia e skills'
        'Revisar esta visão após a primeira entrega'
    )
    backlog_seeds = @(
        @{ n = 1; title = 'Clarificar mandato e paths no manifest'; priority = 'P2' }
        @{ n = 2; title = 'Revisar visão após primeira entrega do papel'; priority = 'P3' }
    )
}

function Get-DomainLens {
    param([string]$Id, [string]$Name, [string]$Description)
    return @{
        headline = "o que é este repositório na cabeça do domínio «$Name»?"
        ca_lens  = "Domínio «$Name» = entidades + invariantes no núcleo; adapters (UI/IoT/billing) não ditam regras."
        opinion_seed = "Como agent de domínio «$Name», meu escopo é proteger invariantes e fatos de negócio — não a stack. Avalio o que já está no código/docs vs intenção (PRD) e mantenho backlog de regras ainda não materializadas."
        questions = @(
            "Quais fatos de negócio «$Name» já estão no código/docs?"
            'Quais invariantes o domínio deve proteger?'
            'O que ainda é só intenção (PRD) sem implementação?'
        )
        signals = @(
            @{ id = 'domain-docs'; ok = $hasDocs; asis = 'Docs disponíveis para enriquecer o domínio'; gap = 'Sem docs — domínio só no manifesto' }
            @{ id = 'domain-code'; ok = ($srcFileCount -gt 0); asis = 'Há código que pode carregar regras de domínio'; gap = 'Sem código — domínio é visão de bootstrap' }
            @{ id = 'mandate'; ok = [bool]$Description; asis = "Mandato: $Description"; gap = 'Sem description no domain agent' }
        )
        action_plan = @(
            "Materializar invariantes de «$Name» em testes/specs"
            'Paths do domain sync batendo com árvore real'
            'Pareceres de domínio só via consulta (ECP)'
        )
        backlog_seeds = @(
            @{ n = 1; title = "Inventariar invariantes de «$Name» (docs↔código)"; priority = 'P1' }
            @{ n = 2; title = "Specs/testes para invariantes críticos de «$Name»"; priority = 'P1' }
            @{ n = 3; title = 'Alinhar paths do domain sync à árvore real'; priority = 'P2' }
        )
    }
}

function Build-Opinion {
    param([hashtable]$Lens, [string]$AgentId, [string]$Kind)
    $blurb = Get-AppTypeBlurb
    $parts = @(
        $Lens.opinion_seed
        ""
        "**Tipo de aplicação (heurística):** ``$appType`` — $blurb."
        "**Lente Clean Architecture / especialidade:** $($Lens.ca_lens)"
        "**Papel ($Kind / ``$AgentId``):** opinar e planejar no próprio escopo; execução de produto só via ECP."
    )
    return ($parts -join "`n")
}

# ---------------------------------------------------------------------------
# Backlog + events (YAML sidecars, merge-safe)
# ---------------------------------------------------------------------------
function Format-BacklogId {
    param([string]$AgentId, [int]$N)
    return ('{0}-BL-{1:D3}' -f $AgentId, $N)
}

function Parse-BacklogYaml {
    param([string]$Path)
    $items = @()
    if (-not (Test-Path -LiteralPath $Path)) { return $items }
    $raw = Get-Content -LiteralPath $Path -Raw
    $blocks = [regex]::Split($raw, '(?m)^\s*-\s+id:\s*')
    foreach ($b in $blocks) {
        if ($b -notmatch '^(?<id>[A-Za-z][A-Za-z0-9_-]*-BL-\d{3})\s*') { continue }
        $id = $Matches['id'].Trim()
        $title = ''; $status = 'todo'; $priority = 'P2'; $links = ''
        if ($b -match '(?m)^\s*title:\s*"?(?<t>.+?)"?\s*$') { $title = $Matches['t'].Trim().Trim('"') }
        if ($b -match '(?m)^\s*status:\s*(?<s>\w+)') { $status = $Matches['s'].Trim() }
        if ($b -match '(?m)^\s*priority:\s*(?<p>\S+)') { $priority = $Matches['p'].Trim() }
        if ($b -match '(?m)^\s*links:\s*\[(?<l>[^\]]*)\]') { $links = $Matches['l'].Trim() }
        $items += [ordered]@{
            id       = $id
            title    = $title
            status   = $status
            priority = $priority
            links    = $links
        }
    }
    return $items
}

function Merge-Backlog {
    param(
        [string]$AgentId,
        [array]$Seeds,
        [array]$Existing
    )
    $byId = @{}
    foreach ($e in $Existing) {
        if ($e.id) { $byId[$e.id] = [ordered]@{
                id = $e.id; title = $e.title; status = $e.status
                priority = $e.priority; links = $e.links
            }
        }
    }
    $preservedDone = 0
    $added = 0
    $updated = 0
    foreach ($s in $Seeds) {
        $id = Format-BacklogId -AgentId $AgentId -N ([int]$s.n)
        if ($byId.ContainsKey($id)) {
            $cur = $byId[$id]
            if ($cur.status -in @('done', 'cancelled')) {
                $preservedDone++
                continue
            }
            if ($s.title -and $cur.title -ne $s.title) {
                $cur.title = $s.title
                $updated++
            }
            if ($s.priority) { $cur.priority = $s.priority }
            $byId[$id] = $cur
        } else {
            $byId[$id] = [ordered]@{
                id       = $id
                title    = $s.title
                status   = 'todo'
                priority = $(if ($s.priority) { $s.priority } else { 'P2' })
                links    = ''
            }
            $added++
        }
    }
    $merged = @($byId.Values | Sort-Object { $_.id })
    return @{
        items          = $merged
        added          = $added
        updated        = $updated
        preserved_done = $preservedDone
    }
}

function Write-BacklogYaml {
    param([string]$AgentId, [array]$Items, [string]$Path)
    $lines = @(
        '# Agent backlog — merge-safe; status done/cancelled preservados no Refresh'
        "version: $schemaVersion"
        "agent: $AgentId"
        "updated_at: `"$stamp`""
        'items:'
    )
    foreach ($it in $Items) {
        $titleEsc = ($it.title -replace '"', '''')
        $linksVal = if ($it.links) { "[$($it.links)]" } else { '[]' }
        $lines += "  - id: $($it.id)"
        $lines += "    title: `"$titleEsc`""
        $lines += "    status: $($it.status)"
        $lines += "    priority: $($it.priority)"
        $lines += "    links: $linksVal"
    }
    Set-Content -LiteralPath $Path -Value ($lines -join "`n") -Encoding UTF8
}

function Append-VisionEvent {
    param(
        [string]$Path,
        [string]$AgentId,
        [string]$Type,
        [string]$Summary,
        [hashtable]$Meta
    )
    $prev = ''
    if (Test-Path -LiteralPath $Path) {
        $prev = Get-Content -LiteralPath $Path -Raw
    }
    if (-not $prev -or $prev -notmatch '(?m)^events:\s*$') {
        $prev = @"
# Append-only vision events (feedback loop)
version: $schemaVersion
agent: $AgentId
events:
"@
    }
    $metaLines = @()
    foreach ($k in $Meta.Keys) {
        $metaLines += "    $($k): $($Meta[$k])"
    }
    $entry = @"

  - at: "$stamp"
    type: $Type
    summary: "$Summary"
$($metaLines -join "`n")
"@
    Set-Content -LiteralPath $Path -Value ($prev.TrimEnd() + $entry) -Encoding UTF8
}

function Format-BacklogTable {
    param([array]$Items)
    if (-not $Items -or $Items.Count -eq 0) {
        return '_Nenhum item de backlog._'
    }
    $rows = @('| ID | Priority | Status | Title |', '|----|----------|--------|-------|')
    foreach ($it in $Items) {
        $rows += "| ``$($it.id)`` | $($it.priority) | $($it.status) | $($it.title) |"
    }
    return ($rows -join "`n")
}

function Format-EventsPreview {
    param([string]$EventsPath)
    if (-not (Test-Path -LiteralPath $EventsPath)) {
        return '- (sem eventos ainda)'
    }
    $raw = Get-Content -LiteralPath $EventsPath -Raw
    $matches = [regex]::Matches($raw, '(?ms)^\s*-\s+at:\s*"(?<at>[^"]+)"\s*\r?\n\s*type:\s*(?<type>\S+)\s*\r?\n\s*summary:\s*"(?<sum>[^"]*)"')
    if ($matches.Count -eq 0) { return '- (sem eventos parseáveis)' }
    $take = $matches | Select-Object -Last 5
    $lines = @()
    foreach ($m in $take) {
        $lines += "- ``$($m.Groups['at'].Value)`` · **$($m.Groups['type'].Value)** — $($m.Groups['sum'].Value)"
    }
    return ($lines -join "`n")
}

# ---------------------------------------------------------------------------
# Discover agents
# ---------------------------------------------------------------------------
function Get-AgentManifests {
    $list = @()
    $roots = @(
        (Join-Path $RepoRoot '.agents')
    )
    foreach ($base in $roots) {
        if (-not (Test-Path $base)) { continue }
        Get-ChildItem $base -Filter '*.agent.yaml' -File -ErrorAction SilentlyContinue | ForEach-Object {
            $list += $_
        }
        $domainDir = Join-Path $base 'domain'
        if ($IncludeDomain -or $true) {
            if (Test-Path $domainDir) {
                Get-ChildItem $domainDir -Filter '*.agent.yaml' -File -ErrorAction SilentlyContinue | ForEach-Object {
                    $list += $_
                }
            }
        }
        $specDir = Join-Path $base 'specialists'
        if ($IncludeSpecialists) {
            if (Test-Path $specDir) {
                Get-ChildItem $specDir -Filter '*.agent.yaml' -File -ErrorAction SilentlyContinue | ForEach-Object {
                    $list += $_
                }
            }
        }
    }
    return $list
}

$filterIds = @()
if ($Agents) {
    $filterIds = @($Agents -split '[,;\s]+' | Where-Object { $_ })
}

$manifests = @(Get-AgentManifests)
if ($filterIds.Count -gt 0) {
    $manifests = @($manifests | Where-Object {
        $raw = Get-Content $_.FullName -Raw
        $id = Get-ScalarField -Raw $raw -Field 'id'
        $id -in $filterIds
    })
}

if ($manifests.Count -eq 0) {
    Write-Warning 'Nenhum agent manifest encontrado em .agents/ — gerando visões seed para papéis canônicos.'
}

# ---------------------------------------------------------------------------
# Render vision markdown
# ---------------------------------------------------------------------------
function Build-VisionMarkdown {
    param(
        [string]$AgentId,
        [string]$AgentName,
        [string]$Description,
        [string]$Kind,
        [hashtable]$Lens,
        [string[]]$ScopePaths,
        [array]$BacklogItems,
        [string]$EventsPreview,
        [string]$EventType
    )

    $asis = New-Object System.Collections.Generic.List[string]
    $gaps = New-Object System.Collections.Generic.List[string]
    foreach ($sig in $Lens.signals) {
        if ($sig.ok) {
            if ($sig.asis) { [void]$asis.Add("- $($sig.asis) _(signal: $($sig.id))_") }
        } else {
            if ($sig.gap) { [void]$gaps.Add("- $($sig.gap) _(signal: $($sig.id))_") }
        }
    }

    if ($isEmptyish) {
        [void]$gaps.Add('- Repositório vazio / quase vazio — esta visão é **bootstrap** (action plan inicial), não auditoria de legado.')
        if ($asis.Count -eq 0) {
            [void]$asis.Add('- Sem produto instalado ainda; harness/agents podem existir como intenção.')
        }
    }

    if ($ScopePaths -and $ScopePaths.Count -gt 0) {
        $missingPaths = @()
        foreach ($p in $ScopePaths) {
            $first = ($p -split '[/\\]')[0] -replace '\*\*', '' -replace '\*', ''
            if (-not $first -or $first -eq '**') { continue }
            if ($first -eq '.' ) { continue }
            if (-not (Test-Path -LiteralPath (Join-Path $RepoRoot $first)) -and $first -notin $topDirs) {
                $missingPaths += $p
            }
        }
        if ($missingPaths.Count -gt 0) {
            [void]$gaps.Add('- Paths do escopo ainda inexistentes: `' + (($missingPaths | Select-Object -First 8) -join '`, `') + '`')
        } else {
            [void]$asis.Add('- Paths de escopo do manifest observáveis na árvore (heurística de 1º segmento).')
        }
    }

    $opinion = Build-Opinion -Lens $Lens -AgentId $AgentId -Kind $Kind
    $qLines = ($Lens.questions | ForEach-Object { "- $_" }) -join "`n"
    $asisBlock = if ($asis.Count) { ($asis -join "`n") } else { '- (sem sinais positivos — ver gaps)' }
    $gapsBlock = if ($gaps.Count) { ($gaps -join "`n") } else { '- Nenhum gap heurístico óbvio (revisão humana ainda recomendada).' }
    $planSource = if ($Lens.action_plan) { $Lens.action_plan } elseif ($Lens.tobe) { $Lens.tobe } else { @('Definir próximos passos do papel') }
    $planBlock = ($planSource | ForEach-Object { "1. $_" }) -join "`n"
    $backlogBlock = Format-BacklogTable -Items $BacklogItems
    $mode = if ($isEmptyish) { 'bootstrap-empty' } else { 'observe-existing' }
    $bt = [char]96
    $codeExec = $bt + 'execution_role.can_execute' + $bt
    $codeTask = $bt + 'arah task create' + $bt
    $codeAssess = $bt + 'arah assess-repo' + $bt
    $codeRefresh = $bt + 'arah assess-repo -Refresh' + $bt
    $codeVision = $bt + 'arah vision update' + $bt
    $codeBacklog = $bt + 'arah backlog sync' + $bt
    $codeSkill = $bt + 'repo-perspective-assess' + $bt
    $codeId = $bt + $AgentId + $bt

    @"
# Visão — $AgentName ($codeId)

> **Experimental** · ARAH Repo Perspective Assessment v$schemaVersion  
> Gerado/atualizado em $stamp · harness $harnessVersion · modo $mode · event ``$EventType``  
> Fase de bootstrap (pareceres em série; não substitui Execution Control em tarefas de entrega).

## Opinião (escopo neste tipo de aplicação)

$opinion

$Description

**Tipo:** $Kind

$($Lens.headline)

## Perguntas-guia

$qLines

## As-Is (status atual · escopo do papel)

$asisBlock

### Snapshot compartilhado (evidência)

| Sinal | Valor |
|-------|-------|
| App type | $appType |
| Linguagens | $(if ($languages.Count) { $languages -join ', ' } else { '—' }) |
| Frameworks | $(if ($frameworks.Count) { $frameworks -join ', ' } else { '—' }) |
| Top dirs | $(if ($topDirs.Count) { ($topDirs | Select-Object -First 12) -join ', ' } else { '—' }) |
| Arquivos src / test / docs | $srcFileCount / $testFileCount / $docFileCount |
| Testes / CI / Specs / ADR | $hasTests / $hasCi / $hasSpecs / $hasAdr |

## Gaps (deste ponto de vista)

$gapsBlock

## Action plan (plano de ação deste papel)

$planBlock

## Backlog (stories do papel)

IDs estáveis ``{agent}-BL-NNN``. Status ``done``/``cancelled`` são preservados em $codeRefresh / $codeVision / $codeBacklog.

Sidecar: ``$AgentId.backlog.yaml``

$backlogBlock

## Memory / events (append-only)

Sidecar: ``$AgentId.events.yaml``

$EventsPreview

## Autonomia sugerida (gates / padrões)

- Pode **propor** padrões, checklists, backlog e gates alinhados a este papel.
- Só **executa** mudanças de produto se $codeExec e houver contrato ECP com este agent como primary_executor.
- Em dúvida de segurança/compliance: parecer → humano ou security review — nunca bypass de gate.
- Em interações futuras: atualizar este backlog/memória ($codeRefresh) — feedback loop do papel.

## Próximo passo concreto

Priorize 1–3 itens ``todo`` do backlog, abra tarefa ECP ($codeTask) com **um** executor. Não trate esta assessment como entrega.

---
_Artefato gerado por $codeAssess · skill $codeSkill_
"@
}

# ---------------------------------------------------------------------------
# Execute serial assessments
# ---------------------------------------------------------------------------
$outPath = if ([System.IO.Path]::IsPathRooted($OutDir)) { $OutDir } else { Join-Path $RepoRoot $OutDir }
if (-not $DryRun) {
    New-Item -ItemType Directory -Path $outPath -Force | Out-Null
}

$results = @()
$order = 0

function Emit-Vision {
    param(
        [string]$Id,
        [string]$Name,
        [string]$Description,
        [string]$Kind,
        [hashtable]$Lens,
        [string[]]$ScopePaths
    )
    $script:order++
    Write-Host ("[{0}] assess {1} ({2})" -f $script:order, $Id, $Kind)

    $mdFile = Join-Path $outPath "$Id.md"
    $blFile = Join-Path $outPath "$Id.backlog.yaml"
    $evFile = Join-Path $outPath "$Id.events.yaml"

    if ((Test-Path $mdFile) -and -not $Force -and -not $Refresh) {
        Write-Host "  skip (exists, use -Force or -Refresh): $mdFile"
        $script:results += [ordered]@{
            id          = $Id
            name        = $Name
            kind        = $Kind
            file        = Join-Path $OutDir "$Id.md"
            gap_signals = 0
            backlog     = 0
            order       = $script:order
            skipped     = $true
        }
        return
    }

    $seeds = @()
    if ($Lens.backlog_seeds) { $seeds = @($Lens.backlog_seeds) }
    $existing = @(Parse-BacklogYaml -Path $blFile)
    # Also recover done items from MD table if sidecar missing
    if ($existing.Count -eq 0 -and (Test-Path $mdFile)) {
        $mdRaw = Get-Content $mdFile -Raw
        foreach ($m in [regex]::Matches($mdRaw, '\|\s*`(?<id>[a-z0-9-]+-BL-\d+)`\s*\|\s*(?<pri>P\d)\s*\|\s*(?<st>\w+)\s*\|\s*(?<title>[^|]+)\|')) {
            $existing += [ordered]@{
                id       = $m.Groups['id'].Value
                title    = $m.Groups['title'].Value.Trim()
                status   = $m.Groups['st'].Value.Trim()
                priority = $m.Groups['pri'].Value.Trim()
                links    = ''
            }
        }
    }

    $merge = Merge-Backlog -AgentId $Id -Seeds $seeds -Existing $existing
    $eventType = if ($Refresh) { if ($BacklogOnly) { 'backlog-sync' } else { 'refresh' } } `
        elseif (Test-Path $mdFile) { 'force-rewrite' } else { 'assess' }

    if ($DryRun) {
        Write-Host "  dry-run → $mdFile + backlog ($($merge.items.Count) items, +$($merge.added)/~$($merge.updated), done-kept $($merge.preserved_done))"
    } else {
        if (-not $BacklogOnly -or -not (Test-Path $mdFile)) {
            $eventsPreview = '- (evento atual será anexado após gravação)'
            $md = Build-VisionMarkdown -AgentId $Id -AgentName $Name -Description $Description -Kind $Kind `
                -Lens $Lens -ScopePaths $ScopePaths -BacklogItems $merge.items `
                -EventsPreview $eventsPreview -EventType $eventType
            # write backlog+events first so preview can include history on next runs
            Write-BacklogYaml -AgentId $Id -Items $merge.items -Path $blFile
            Append-VisionEvent -Path $evFile -AgentId $Id -Type $eventType `
                -Summary "Vision $eventType · gaps/signals refresh · backlog merge" `
                -Meta @{
                    gap_signals        = @($Lens.signals | Where-Object { -not $_.ok }).Count
                    backlog_total      = $merge.items.Count
                    backlog_added      = $merge.added
                    backlog_updated    = $merge.updated
                    backlog_kept_done  = $merge.preserved_done
                    app_type           = $appType
                }
            $eventsPreview = Format-EventsPreview -EventsPath $evFile
            $md = Build-VisionMarkdown -AgentId $Id -AgentName $Name -Description $Description -Kind $Kind `
                -Lens $Lens -ScopePaths $ScopePaths -BacklogItems $merge.items `
                -EventsPreview $eventsPreview -EventType $eventType
            Set-Content -LiteralPath $mdFile -Value $md -Encoding UTF8
            Write-Host "  wrote $mdFile"
            Write-Host "  wrote $blFile ($($merge.items.Count) items)"
            Write-Host "  appended $evFile"
        } else {
            Write-BacklogYaml -AgentId $Id -Items $merge.items -Path $blFile
            Append-VisionEvent -Path $evFile -AgentId $Id -Type $eventType `
                -Summary "Backlog sync only · merge seeds" `
                -Meta @{
                    backlog_total     = $merge.items.Count
                    backlog_added     = $merge.added
                    backlog_updated   = $merge.updated
                    backlog_kept_done = $merge.preserved_done
                }
            $preview = Format-EventsPreview -EventsPath $evFile
            $md = Build-VisionMarkdown -AgentId $Id -AgentName $Name -Description $Description -Kind $Kind `
                -Lens $Lens -ScopePaths $ScopePaths -BacklogItems $merge.items `
                -EventsPreview $preview -EventType $eventType
            Set-Content -LiteralPath $mdFile -Value $md -Encoding UTF8
            Write-Host "  backlog-only sync → $blFile + $mdFile"
        }
    }

    $gapCount = @($Lens.signals | Where-Object { -not $_.ok }).Count
    if ($isEmptyish) { $gapCount++ }
    $script:results += [ordered]@{
        id              = $Id
        name            = $Name
        kind            = $Kind
        file            = Join-Path $OutDir "$Id.md"
        backlog_file    = Join-Path $OutDir "$Id.backlog.yaml"
        gap_signals     = $gapCount
        backlog         = $merge.items.Count
        backlog_added   = $merge.added
        order           = $script:order
        skipped         = $false
    }
}

# Seed agents if none
$seedOps = @(
    @{ id = 'qa'; name = 'QA'; desc = 'Qualidade e testes.' },
    @{ id = 'solutions-architect'; name = 'Solutions Architect'; desc = 'Arquitetura e ADRs.' },
    @{ id = 'backend'; name = 'Backend'; desc = 'API e domínio server-side.' },
    @{ id = 'security'; name = 'Security'; desc = 'Segurança e compliance.' },
    @{ id = 'docs-steward'; name = 'Docs Steward'; desc = 'Documentação.' },
    @{ id = 'test-architect'; name = 'Test Architect'; desc = 'Estratégia de testes.' }
)

if ($manifests.Count -eq 0) {
    foreach ($s in $seedOps) {
        if ($filterIds.Count -gt 0 -and $s.id -notin $filterIds) { continue }
        $lens = if ($lenses.ContainsKey($s.id)) { $lenses[$s.id] } else { $defaultLens }
        Emit-Vision -Id $s.id -Name $s.name -Description $s.desc -Kind 'operational' -Lens $lens -ScopePaths @()
    }
} else {
    foreach ($m in ($manifests | Sort-Object Name)) {
        $raw = Get-Content $m.FullName -Raw
        $id = Get-ScalarField -Raw $raw -Field 'id'
        if (-not $id) { $id = ($m.BaseName -replace '\.agent$', '') }
        $name = Get-ScalarField -Raw $raw -Field 'name'
        if (-not $name) { $name = $id }
        $desc = Get-ScalarField -Raw $raw -Field 'description'
        if (-not $desc) { $desc = '' }
        if ($raw -match '(?ms)^description:\s*\|\s*\r?\n((?:\s+[^\r\n]+\r?\n)+)') {
            $desc = (($Matches[1] -split "`n" | ForEach-Object { $_.Trim() } | Where-Object { $_ }) -join ' ')
        }
        if (-not $desc -and $raw -match '(?ms)^enrich:\s*\|\s*\r?\n((?:\s+[^\r\n]+\r?\n)+)') {
            $desc = (($Matches[1] -split "`n" | ForEach-Object { $_.Trim() } | Where-Object { $_ }) -join ' ')
        }
        if (-not $desc) {
            $desc = Get-ScalarField -Raw $raw -Field 'enrich'
            if (-not $desc) { $desc = '' }
        }
        $kind = 'operational'
        if ($m.Directory.Name -eq 'domain' -or ($raw -match '(?m)^type:\s*domain')) { $kind = 'domain' }
        if ($m.Directory.Name -eq 'specialists') { $kind = 'specialist' }

        $scopePaths = @()
        $scopeBlock = Get-TopLevelBlock -Raw $raw -Key 'scope'
        if ($scopeBlock -match '(?ms)paths:\s*\r?\n((?:\s+-\s+.+\r?\n?)+)') {
            $scopePaths = @([regex]::Matches($Matches[1], '-\s+(.+)$', 'Multiline') | ForEach-Object {
                $_.Groups[1].Value.Trim().Trim('"').Trim("'")
            })
        }

        if ($kind -eq 'domain') {
            $lens = Get-DomainLens -Id $id -Name $name -Description $desc
            if ($lenses.ContainsKey($id)) {
                $lens = $lenses[$id]
            }
        } elseif ($lenses.ContainsKey($id)) {
            $lens = $lenses[$id]
        } else {
            $lens = $defaultLens
        }

        Emit-Vision -Id $id -Name $name -Description $desc -Kind $kind -Lens $lens -ScopePaths $scopePaths
    }
}

# Index + summary
$bt = [char]96
$codeAssessIdx = $bt + 'arah assess-repo' + $bt
$codeRefreshIdx = $bt + 'arah assess-repo -Refresh' + $bt
$codeVisionIdx = $bt + 'arah vision update' + $bt
$codeBacklogIdx = $bt + 'arah backlog sync' + $bt
$codeSkillIdx = $bt + 'repo-perspective-assess' + $bt
$indexRows = ($results | ForEach-Object {
    $idCode = $bt + $_.id + $bt
    $bl = if ($null -ne $_.backlog) { $_.backlog } else { '—' }
    "| $($_.order) | $idCode | $($_.kind) | $($_.gap_signals) | $bl | [$($_.id).md]($($_.id).md) |"
}) -join "`n"

$indexMd = @"
# ARAH Repo Visions (experimental v$schemaVersion)

Gerado/atualizado em $stamp · harness $harnessVersion · app_type **$appType**  
Comando: $codeAssessIdx · skill: $codeSkillIdx

## Modelo de execução

Esta assessment é uma **fase de bootstrap**: pareceres **em série** (um agent por vez).
Não viola o protocolo de entrega (1 primary_executor + consultas limitadas) —
são artefatos de observação + **backlog/memória dinâmicos** por papel, não uma tarefa de produto.

## Schema por visão

| Seção | Função |
|-------|--------|
| Opinião | Perspectiva do papel neste tipo de aplicação + lente Clean Architecture |
| As-Is / Gaps | Avaliação scoped ao mandato |
| Action plan | Plano de ação do papel |
| Backlog | Stories com IDs ``{agent}-BL-NNN`` (sidecar YAML) |
| Memory/events | Log append-only do feedback loop |

## Snapshot

- Empty/bootstrap: **$isEmptyish**
- App type: **$appType**
- Linguagens: $(if ($languages.Count) { $languages -join ', ' } else { '—' })
- Frameworks: $(if ($frameworks.Count) { $frameworks -join ', ' } else { '—' })
- Visões: **$($results.Count)**

## Índice

| # | Agent | Tipo | Gaps | Backlog | Arquivo |
|---|-------|------|------|---------|---------|
$indexRows

## Como usar

1. Leia a **opinião** + backlog do papel na primeira interação.
2. Priorize itens ``todo`` → $bt$('arah task create')$bt com um executor.
3. Em interações futuras: $codeRefreshIdx (ou $codeVisionIdx / $codeBacklogIdx) — **mescla** backlog (não apaga ``done``) e registra evento.
4. Re-rode após mudanças estruturais.

Ver docs/REPO_VISIONS.md no harness (ou cópia no consumidor).
"@

$summaryYaml = @"
# Generated by arah assess-repo — do not hand-edit as source of truth
version: $schemaVersion
experimental: true
generated_at: "$stamp"
harness_version: "$harnessVersion"
repo_root: "$RepoRoot"
mode: $(if ($isEmptyish) { 'bootstrap-empty' } else { 'observe-existing' })
refresh: $($Refresh.ToString().ToLower())
out_dir: "$OutDir"
snapshot:
  emptyish: $($snapshot.emptyish.ToString().ToLower())
  app_type: '$appType'
  languages: [$( ($snapshot.languages | ForEach-Object { "'$_'" }) -join ', ' )]
  frameworks: [$( ($snapshot.frameworks | ForEach-Object { "'$_'" }) -join ', ' )]
  src_files: $($snapshot.src_files)
  test_files: $($snapshot.test_files)
  doc_files: $($snapshot.doc_files)
  has_tests: $($snapshot.has_tests.ToString().ToLower())
  has_ci: $($snapshot.has_ci.ToString().ToLower())
  has_docs: $($snapshot.has_docs.ToString().ToLower())
  has_specs: $($snapshot.has_specs.ToString().ToLower())
visions:
$(($results | ForEach-Object { @"
  - id: $($_.id)
    kind: $($_.kind)
    file: $($_.file)
    backlog_file: $($_.backlog_file)
    gap_signals: $($_.gap_signals)
    backlog_items: $($_.backlog)
"@ }) -join "`n")
"@

if (-not $SkipIndex) {
    $idxFile = Join-Path $outPath 'README.md'
    $sumFile = Join-Path $outPath 'summary.yaml'
    if ($DryRun) {
        Write-Host "dry-run index → $idxFile"
        Write-Host "dry-run summary → $sumFile"
    } else {
        Set-Content -LiteralPath $idxFile -Value $indexMd -Encoding UTF8
        Set-Content -LiteralPath $sumFile -Value $summaryYaml -Encoding UTF8
        Write-Host "wrote $idxFile"
        Write-Host "wrote $sumFile"
    }
}

Write-Host ""
Write-Host "assess-repo complete: $($results.Count) vision(s) → $outPath (schema v$schemaVersion)"
if ($isEmptyish) {
    Write-Host "mode: bootstrap-empty · app_type: $appType"
} else {
    Write-Host "mode: observe-existing · app_type: $appType"
}
if ($Refresh) {
    Write-Host "refresh: backlog merge enabled (done/cancelled preserved)"
}
exit 0

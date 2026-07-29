#Requires -Version 5.1
<#
.SYNOPSIS
  Assessment experimental: cada agent observa o repo na sua perspectiva
  e grava As-Is + Gaps + To-Be em .arah/visions/ (ou docs/_arah/visions/).
.DESCRIPTION
  Fase de bootstrap controlada (exceção documentada ao modelo 1-executor):
  pareceres em série, um por agent, sem handoff consultor→consultor.
  Não altera código de produto. Não faz merge. Opt-in pós-install.
.EXAMPLE
  ./assess-repo.ps1
  ./assess-repo.ps1 -OutDir docs/_arah/visions -Agents qa,backend,security
  ./assess-repo.ps1 -DryRun
#>
param(
    [string]$RepoRoot = '',
    [string]$OutDir = '.arah/visions',
    [string]$Agents = '',
    [switch]$IncludeDomain,
    [switch]$IncludeSpecialists,
    [switch]$Force,
    [switch]$DryRun,
    [switch]$SkipIndex
)

$ErrorActionPreference = 'Stop'

if (-not $RepoRoot) {
    $RepoRoot = (Get-Location).Path
}
$RepoRoot = (Resolve-Path -LiteralPath $RepoRoot).Path

. (Join-Path $PSScriptRoot 'yaml-lite.ps1')

$stamp = Get-Date -Format 'yyyy-MM-ddTHH:mm:ssK'
$harnessVersion = 'unknown'
$verFile = Join-Path $RepoRoot '.arah-version'
if (Test-Path -LiteralPath $verFile) {
    $vr = Get-Content -LiteralPath $verFile -Raw
    if ($vr -match '(?m)^version:\s*(.+)$') { $harnessVersion = $Matches[1].Trim() }
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
$srcFileCount = Count-Files @('*.js', '*.ts', '*.tsx', '*.py', '*.go', '*.rs', '*.java', '*.cs', '*.cjs', '*.mjs')
$testFileCount = Count-Files @('*.test.js', '*.test.ts', '*.test.cjs', '*.spec.js', '*.spec.ts', '*_test.go', 'test_*.py')
$docFileCount = @(Get-ChildItem (Join-Path $RepoRoot 'docs') -Recurse -File -ErrorAction SilentlyContinue |
    Where-Object { $_.Extension -match '\.(md|yaml|yml|mdx)$' }).Count

$trackedish = @(Get-ChildItem $RepoRoot -Force -ErrorAction SilentlyContinue | Where-Object {
    $_.Name -notin @('.git', 'node_modules', '.venv', 'dist', 'build', 'coverage', 'vendor')
}).Count
$isEmptyish = ($srcFileCount -eq 0) -and ($trackedish -le 8) -and (-not (Test-RepoPath 'package.json')) -and (-not (Test-RepoPath 'go.mod'))

$snapshot = [ordered]@{
    emptyish          = $isEmptyish
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
    src_files         = $srcFileCount
    test_files        = $testFileCount
    doc_files         = $docFileCount
}

# ---------------------------------------------------------------------------
# Perspective lenses (role → guiding questions + heuristic checks)
# ---------------------------------------------------------------------------
$lenses = @{
    'qa' = @{
        headline = 'o que é este repositório na cabeça do teste / QA?'
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
        tobe = @(
            'Matriz de risco → casos críticos com gate no PR'
            'Comando `tests.all` estável no arah.config.yaml'
            'Checklist QA amarrado a evidência (não só review textual)'
        )
    }
    'test-architect' = @{
        headline = 'o que é este repositório na cabeça da arquitetura de testes?'
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
        tobe = @(
            'docs/testing/test-strategy.md com riscos e pirâmide'
            'Gates CI por classe de risco'
            'Critérios a11y/perf explícitos onde o produto exigir'
        )
    }
    'solutions-architect' = @{
        headline = 'o que é este repositório na cabeça da arquitetura?'
        questions = @(
            'Há fronteiras claras (módulos, apps, contratos)?'
            'Decisões estruturais estão em ADRs?'
            'O desenho alvo (To-Be) está documentado vs o As-Is?'
        )
        signals = @(
            @{ id = 'adr'; ok = $hasAdr; asis = 'docs/adr ou docs/architecture presente'; gap = 'Sem ADRs / docs de arquitetura' }
            @{ id = 'structure'; ok = ($topDirs.Count -ge 2); asis = ("Top-level: " + ($topDirs -join ', ')); gap = 'Estrutura flat ou quase vazia — fronteiras ainda implícitas' }
            @{ id = 'specs'; ok = $hasSpecs; asis = 'Specs SDD presentes'; gap = 'Sem specs — arquitetura sem âncora de requisito' }
            @{ id = 'readme'; ok = $hasReadme; asis = 'README presente'; gap = 'Sem README — onboarding arquitetural frágil' }
        )
        tobe = @(
            'Mapa de contextos/limites versionado'
            'ADRs para decisões estruturais'
            'Contratos públicos (OpenAPI/events) quando houver integração'
        )
    }
    'backend' = @{
        headline = 'o que é este repositório na cabeça do developer/backend?'
        questions = @(
            'Onde mora a lógica de domínio e a API?'
            'Há testes de domínio/API executáveis?'
            'Há caminho claro de delivery (scripts, CI)?'
        )
        signals = @(
            @{ id = 'src'; ok = ($srcFileCount -gt 0); asis = "$srcFileCount arquivo(s) de código detectados"; gap = 'Repo sem código — bootstrap de backend necessário' }
            @{ id = 'api-shape'; ok = (($topDirs -contains 'backend') -or ($topDirs -contains 'api') -or ($topDirs -contains 'services') -or ($topDirs -contains 'apps') -or (Test-RepoPath 'domain.js') -or (Test-RepoPath 'packages')); asis = 'Há indício de domínio/API/pacotes'; gap = 'Sem backend/api/services/packages observáveis' }
            @{ id = 'tests'; ok = ($testFileCount -gt 0); asis = 'Testes presentes'; gap = 'Sem testes — regressão de backend sem rede de segurança' }
            @{ id = 'stack'; ok = ($languages.Count -gt 0); asis = ("Stack: " + ($languages -join ', ') + $(if ($frameworks.Count) { ' / ' + ($frameworks -join ', ') } else { '' })); gap = 'Stack não detectada' }
        )
        tobe = @(
            'Fronteira domínio × app × infra explícita'
            'Testes de domínio no caminho crítico do CI'
            'Contratos versionados se houver clientes (PWA/APK/hub)'
        )
    }
    'frontend' = @{
        headline = 'o que é este repositório na cabeça do frontend?'
        questions = @(
            'UI é PWA, app nativo, ou ambos?'
            'Há smoke/e2e para fluxos críticos?'
            'Assets e acessibilidade têm dono?'
        )
        signals = @(
            @{ id = 'ui'; ok = (Test-RepoPath 'index.html') -or (Test-RepoPath 'app.js') -or ($topDirs -contains 'frontend') -or ($topDirs -contains 'web') -or ($topDirs -contains 'apps') -or ($frameworks -contains 'react') -or ($frameworks -contains 'nextjs'); asis = 'Sinais de UI/PWA/frontend detectados'; gap = 'Sem frontend observável' }
            @{ id = 'tests'; ok = ($testFileCount -gt 0) -or ($frameworks -contains 'playwright'); asis = 'Há base de teste (incl. possível e2e)'; gap = 'Sem testes de UI/e2e' }
            @{ id = 'a11y'; ok = $false; asis = ''; gap = 'Sem evidência automática de a11y (heurística conservadora)' }
        )
        tobe = @(
            'Smoke dos fluxos críticos no CI'
            'Contrato UI ↔ domínio estável'
            'Checklist a11y mínimo no QA'
        )
    }
    'security' = @{
        headline = 'o que é este repositório na cabeça da segurança?'
        questions = @(
            'Secrets e .env estão fora do git?'
            'Dependências têm audit/Dependabot?'
            'Superfície de ataque (auth, pagamento, IoT) está mapeada?'
        )
        signals = @(
            @{ id = 'security-md'; ok = $hasSecurityFiles; asis = 'SECURITY.md / Dependabot / scanner presente'; gap = 'Sem política de segurança versionada ou Dependabot' }
            @{ id = 'env-example'; ok = $hasEnvExample; asis = '.env.example presente (bom sinal de higiene)'; gap = 'Sem .env.example — risco de secrets mal documentados' }
            @{ id = 'gitignore'; ok = (Test-RepoPath '.gitignore'); asis = '.gitignore presente'; gap = 'Sem .gitignore' }
            @{ id = 'ci'; ok = $hasCi; asis = 'CI pode hospedar gates de security'; gap = 'Sem CI para gate de security' }
        )
        tobe = @(
            'Gate de secrets no pre-commit/CI'
            'Dependabot ou equivalente'
            'Threat model leve para auth/billing/IoT se aplicável'
        )
    }
    'docs-steward' = @{
        headline = 'o que é este repositório na cabeça da documentação?'
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
        tobe = @(
            'Índice canônico (docs/README ou equivalente)'
            'Sync-docs no PR quando paths de doc mudam'
            'PRD/ADR/spec com donos claros'
        )
    }
    'planner' = @{
        headline = 'o que é este repositório na cabeça do planejamento?'
        questions = @(
            'Há backlog / roadmap acionável?'
            'Classes de trabalho (trivial→release) estão claras?'
            'Dependências entre épicos estão explícitas?'
        )
        signals = @(
            @{ id = 'backlog'; ok = (Test-RepoPath 'docs/BACKLOG_HUB_PRODUCTION.md') -or (Test-RepoPath 'docs/ROADMAP.md') -or (Test-RepoPath 'ROADMAP.md'); asis = 'Backlog/roadmap detectado'; gap = 'Sem backlog/roadmap versionado' }
            @{ id = 'prd'; ok = (Test-RepoPath 'docs/prd'); asis = 'docs/prd presente'; gap = 'Sem PRD estruturado' }
        )
        tobe = @(
            'Backlog com Definition of Ready'
            'Fatiamento por valor + risco'
            'Alinhamento planner ↔ orchestrator (ECP)'
        )
    }
    'orchestrator' = @{
        headline = 'o que é este repositório na cabeça do orquestrador?'
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
        tobe = @(
            'ECP enabled com limites explícitos'
            'Coreografia path-based alinhada à árvore real'
            'Assessment de visões revisada após mudanças estruturais'
        )
    }
    'spec-steward' = @{
        headline = 'o que é este repositório na cabeça do steward de specs?'
        questions = @(
            'Specs existem e validam?'
            'Spec-before-code é proporcional à classe?'
            'Há template e CI de specs?'
        )
        signals = @(
            @{ id = 'specs'; ok = $hasSpecs; asis = 'docs/specs presente'; gap = 'Sem docs/specs' }
            @{ id = 'template'; ok = (Test-RepoPath 'docs/specs/_template.spec.yaml'); asis = 'Template de spec presente'; gap = 'Sem template de spec' }
        )
        tobe = @(
            'Specs para mudanças standard+'
            'Gate validate-specs no CI'
            'Rastreio Spec-Id ↔ PR'
        )
    }
    'pr-steward' = @{
        headline = 'o que é este repositório na cabeça do steward de PRs?'
        questions = @(
            'Template de PR e branch protection existem?'
            'Ready-for-merge é verificável?'
            'Bots de review têm dono de resposta?'
        )
        signals = @(
            @{ id = 'pr-template'; ok = (Test-RepoPath '.github/PULL_REQUEST_TEMPLATE.md'); asis = 'PR template presente'; gap = 'Sem PR template' }
            @{ id = 'ci'; ok = $hasCi; asis = 'CI presente para checks de PR'; gap = 'Sem CI' }
        )
        tobe = @(
            'Checklist de PR alinhado ao DoD'
            'Branch protection documentada'
            'Fluxo address-bot-review quando houver bots'
        )
    }
    'release' = @{
        headline = 'o que é este repositório na cabeça do release?'
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
        tobe = @(
            'Release cut documentado (skill release-cut)'
            'Artefatos com checksum quando aplicável'
            'Classe release no ECP com human gate'
        )
    }
    'clean-craft-advisor' = @{
        headline = 'o que é este repositório na cabeça do clean craft / Uncle Bob?'
        questions = @(
            'Há fronteiras de domínio limpas?'
            'Testes protegem regras de negócio?'
            'Acoplamento UI↔domínio está sob controle?'
        )
        signals = @(
            @{ id = 'domain'; ok = (Test-RepoPath 'domain.js') -or (Test-RepoPath 'packages/domain') -or ($topDirs -contains 'domain'); asis = 'Sinais de pacote/módulo de domínio'; gap = 'Domínio não isolado / inexistente' }
            @{ id = 'tests'; ok = ($testFileCount -gt 0); asis = 'Há testes (verificar se cobrem regras)'; gap = 'Sem testes de domínio' }
        )
        tobe = @(
            'Regras de negócio testáveis sem UI'
            'Dependências apontando para dentro (DIP)'
            'Craft-review nos PRs de domínio'
        )
    }
    'architecture-documenter' = @{
        headline = 'o que é este repositório na cabeça do documentador de arquitetura?'
        questions = @(
            'Diagramas e glossário existem?'
            'C4 / fluxos estão atualizados?'
            'Decisões estão linkadas a ADRs?'
        )
        signals = @(
            @{ id = 'arch-docs'; ok = $hasAdr; asis = 'Pasta de arquitetura/ADR presente'; gap = 'Sem docs de arquitetura' }
            @{ id = 'docs'; ok = $hasDocs; asis = 'docs/ presente'; gap = 'Sem docs/' }
        )
        tobe = @(
            'Diagrama C4 (ou equivalente) versionado'
            'Glossário de bounded contexts'
            'ADRs indexados'
        )
    }
}

$defaultLens = @{
    headline = 'o que é este repositório na perspectiva deste agent?'
    questions = @(
        'Qual o mandato deste papel neste repo?'
        'Quais artefatos As-Is sustentam o trabalho?'
        'O que falta para o To-Be mínimo viável?'
    )
    signals = @(
        @{ id = 'agents'; ok = $hasAgents; asis = 'Harness de agents presente'; gap = 'Sem .agents — papel ainda não operacional' }
        @{ id = 'docs'; ok = $hasDocs; asis = 'Documentação presente'; gap = 'Sem docs' }
        @{ id = 'src'; ok = ($srcFileCount -gt 0); asis = 'Código presente'; gap = 'Repo vazio / sem código — visão de bootstrap' }
    )
    tobe = @(
        'Definir mandato e paths no manifest'
        'Alinhar coreografia e skills'
        'Revisar esta visão após a primeira entrega'
    )
}

# Domain-flavored lens factory
function Get-DomainLens {
    param([string]$Id, [string]$Name, [string]$Description)
    $pathHits = @()
    $missing = @()
    # best-effort: read paths from manifest later; here use id heuristics
    return @{
        headline = "o que é este repositório na cabeça do domínio «$Name»?"
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
        tobe = @(
            "Materializar invariantes de «$Name» em testes/specs"
            'Paths do domain sync batendo com árvore real'
            'Pareceres de domínio só via consulta (ECP)'
        )
        pathHits = $pathHits
        missing = $missing
    }
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
        [string]$Kind,  # operational | domain | specialist
        [hashtable]$Lens,
        [string[]]$ScopePaths
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
        [void]$gaps.Add('- Repositório vazio / quase vazio — esta visão é **bootstrap** (To-Be inicial), não auditoria de legado.')
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

    $qLines = ($Lens.questions | ForEach-Object { "- $_" }) -join "`n"
    $asisBlock = if ($asis.Count) { ($asis -join "`n") } else { '- (sem sinais positivos — ver gaps)' }
    $gapsBlock = if ($gaps.Count) { ($gaps -join "`n") } else { '- Nenhum gap heurístico óbvio (revisão humana ainda recomendada).' }
    $tobeBlock = ($Lens.tobe | ForEach-Object { "1. $_" }) -join "`n"
    $mode = if ($isEmptyish) { 'bootstrap-empty' } else { 'observe-existing' }

    @"
# Visão — $AgentName (`$AgentId`)

> **Experimental** · ARAH Repo Perspective Assessment  
> Gerado em `$stamp` · harness `$harnessVersion` · modo `$mode`  
> Fase de bootstrap (pareceres em série; não substitui Execution Control em tarefas de entrega).

## Na cabeça deste papel

$($Lens.headline)

$Description

**Tipo:** $Kind

## Perguntas-guia

$qLines

## As-Is (status atual)

$asisBlock

### Snapshot compartilhado (evidência)

| Sinal | Valor |
|-------|-------|
| Linguagens | $(if ($languages.Count) { $languages -join ', ' } else { '—' }) |
| Frameworks | $(if ($frameworks.Count) { $frameworks -join ', ' } else { '—' }) |
| Top dirs | $(if ($topDirs.Count) { ($topDirs | Select-Object -First 12) -join ', ' } else { '—' }) |
| Arquivos src / test / docs | $srcFileCount / $testFileCount / $docFileCount |
| Testes / CI / Specs / ADR | $hasTests / $hasCi / $hasSpecs / $hasAdr |

## Gaps (deste ponto de vista)

$gapsBlock

## To-Be (como deveria ser)

$tobeBlock

## Autonomia sugerida (gates / padrões)

- Pode **propor** padrões, checklists e gates alinhados a este papel.
- Só **executa** mudanças de produto se `execution_role.can_execute` e houver contrato ECP com este agent como primary_executor.
- Em dúvida de segurança/compliance: parecer → humano ou security review — nunca bypass de gate.

## Próximo passo concreto

Revise este arquivo, priorize 1–3 gaps, abra tarefa ECP (`arah task create`) com **um** executor. Não trate esta assessment como entrega.

---
_Artefato gerado por ``arah assess-repo`` · skill ``repo-perspective-assess``_
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
    $md = Build-VisionMarkdown -AgentId $Id -AgentName $Name -Description $Description -Kind $Kind -Lens $Lens -ScopePaths $ScopePaths
    $file = Join-Path $outPath "$Id.md"
    if ($DryRun) {
        Write-Host "  dry-run → $file ($($md.Length) chars)"
    } else {
        if ((Test-Path $file) -and -not $Force) {
            Write-Host "  skip (exists, use -Force): $file"
        } else {
            Set-Content -LiteralPath $file -Value $md -Encoding UTF8
            Write-Host "  wrote $file"
        }
    }
    $gapCount = @($Lens.signals | Where-Object { -not $_.ok }).Count
    if ($isEmptyish) { $gapCount++ }
    $script:results += [ordered]@{
        id         = $Id
        name       = $Name
        kind       = $Kind
        file       = Join-Path $OutDir "$Id.md"
        gap_signals = $gapCount
        order      = $script:order
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
        # multiline description: take first non-empty line after description:
        if ($raw -match '(?ms)^description:\s*\|\s*\r?\n((?:\s+[^\r\n]+\r?\n)+)') {
            $desc = (($Matches[1] -split "`n" | ForEach-Object { $_.Trim() } | Where-Object { $_ }) -join ' ')
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
                # merge known specialized domain lenses (test-architect etc.)
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
$indexMd = @"
# ARAH Repo Visions (experimental)

Gerado em `$stamp` · harness `$harnessVersion`  
Comando: ``arah assess-repo`` · skill: ``repo-perspective-assess``

## Modelo de execução

Esta assessment é uma **fase de bootstrap**: pareceres **em série** (um agent por vez).
Não viola o protocolo de entrega (1 primary_executor + consultas limitadas) —
são artefatos de observação, não uma tarefa de produto.

## Snapshot

- Empty/bootstrap: **$isEmptyish**
- Linguagens: $(if ($languages.Count) { $languages -join ', ' } else { '—' })
- Frameworks: $(if ($frameworks.Count) { $frameworks -join ', ' } else { '—' })
- Visões: **$($results.Count)**

## Índice

| # | Agent | Tipo | Gaps (sinais) | Arquivo |
|---|-------|------|---------------|---------|
$(($results | ForEach-Object { "| $($_.order) | ``$($_.id)`` | $($_.kind) | $($_.gap_signals) | [$($_.id).md]($($_.id).md) |" }) -join "`n")

## Como usar

1. Leia a visão do papel relevante na primeira interação com o repo.
2. Priorize gaps → ``arah task create`` com um executor.
3. Re-rode ``arah assess-repo -Force`` após mudanças estruturais.

Ver: [docs/REPO_VISIONS.md](../../docs/REPO_VISIONS.md) (no harness) ou a cópia instalada no consumidor.
"@

$summaryYaml = @"
# Generated by arah assess-repo — do not hand-edit as source of truth
version: 1
experimental: true
generated_at: "$stamp"
harness_version: "$harnessVersion"
repo_root: "$RepoRoot"
mode: $(if ($isEmptyish) { 'bootstrap-empty' } else { 'observe-existing' })
out_dir: "$OutDir"
snapshot:
  emptyish: $($snapshot.emptyish.ToString().ToLower())
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
    gap_signals: $($_.gap_signals)
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
Write-Host "assess-repo complete: $($results.Count) vision(s) → $outPath"
if ($isEmptyish) {
    Write-Host "mode: bootstrap-empty (cada agent criou visão inicial)"
} else {
    Write-Host "mode: observe-existing"
}
exit 0

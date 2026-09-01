# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

## [Unreleased]

### Added

- **Supply chain do release** — `release.yml` ganha job `artifacts`: binários do CLI
  `arah` para linux/darwin (amd64+arm64) e windows (amd64) com `CGO_ENABLED=0`,
  `SHA256SUMS.txt`, assinatura keyless Cosign/Sigstore (OIDC, sem chave privada),
  SBOM CycloneDX (Trivy) e attestation de provenance SLSA
  (`actions/attest-build-provenance`) — tudo anexado ao GitHub Release; dry-run via
  `workflow_dispatch` (builda + assina sem publicar); guia de verificação em
  `docs/INSTALL.md`

## [0.5.0] - 2026-08-11

Release "Runtime Cohesion": runtime Go `arah-core` (PRs [#16](https://github.com/sraphaz/arah-harness/pull/16)–[#23](https://github.com/sraphaz/arah-harness/pull/23)).
**Status honesto:** MCP server, Evidence Graph e conformance suite são **experimentais** —
a estabilização (paridade CLI↔MCP completa, schema de edges, fixtures amplas) segue nas
PRs [#24](https://github.com/sraphaz/arah-harness/pull/24)–[#26](https://github.com/sraphaz/arah-harness/pull/26).

### Added

- **arah-core (fundação 0.5)** — runtime Go hexagonal inspirado em
  [rafaelnicolett/kern](https://github.com/rafaelnicolett/kern) (engenharia, não RAG) — PRs #16/#17
  - `internal/core` — Execution Control tipado (transitions, evidence, TaskService)
  - `internal/adapters/sqlitestore` — StateStore SQLite WAL (`.arah/local/runtime.db`) + EventStore + mirror YAML best-effort
  - `internal/adapters/fsstore` — mirror/compatibilidade PS
  - `internal/envelope` — contrato JSON `ok/code/trace_id/remediation`
  - CLI: `arah task create|status|complete|block`, `arah task timeline`
- **H-16 StateStore schema migrations** — Up/Down versionados (`schema_meta`); v2 `idx_events_kind`; testes upgrade v1→v2 e rollback v2→v1 preservando dados — PRs #17/#23
- **H-14 dry-run + mutation diff + idempotência** — `--dry-run`/`dry_run` em create/complete/block (CLI + MCP) sem persistir/emitir; `MutationResult.diff` textual e `idempotent` ao repetir evidência/razão em estado terminal — PRs #18/#21
- **H-15 kernel gerado** — `kernel/` deixa de ser segunda fonte editável — PRs #19/#22
  - Fase 1: `internal/kernel` + CLI `arah kernel sync|verify`; `kernel/manifest.json` (SHA-256 por arquivo); CI `kernel-integrity`; CONTRIBUTING proíbe edição manual de `kernel/`
  - Fase 2: `go:embed` de `internal/kernel/payload/kernel.zip` + CLI `arah kernel install` (proteção zip-slip/symlink, rollback com restauro de overwrites, payload determinístico)
- **MCP server (experimental)** — `arah mcp serve`, 7 tools (capabilities, get/create/complete/block task, timeline, evidence graph); contrato em [`docs/architecture/mcp-tool-contract.md`](docs/architecture/mcp-tool-contract.md) — PR #16+; paridade complete/block em estabilização (PR #24)
- **Evidence Graph determinístico (experimental)** — `arah evidence graph` (+ MCP) — PR #17; schema de edges em estabilização (PR #25)
- **Conformance suite (experimental, parcial)** — `internal/conformance` (dry-run, error codes, paridade create CLI≡MCP; binário `.exe` no Windows — PR #20); fixtures amplas TBD (PR #26)
- **Direção 0.5 — Runtime Cohesion** (docs)
  - ADR-002: [`docs/adr/002-runtime-cohesion-0.5.md`](docs/adr/002-runtime-cohesion-0.5.md)
  - Arquitetura: [`docs/architecture/RUNTIME_COHESION.md`](docs/architecture/RUNTIME_COHESION.md)
  - Spec draft: `arah-runtime-cohesion`

### Changed

- **Consolidação de versões internas** — `VERSION` é a fonte única, embutida no binário Go via `go:embed` (fim do hardcode em `cmd/arah`); `harness/VERSION`, `.arah-version` e `arah.config.yaml` self-hosted alinhados; `templates/arah.config.yaml` usa placeholder `{{HARNESS_VERSION}}` resolvido pelo `init`; `docs/INSTALL.md` (citava v0.3.0), `SECURITY.md` (declarava 0.2.x) e `capabilities.yaml` (kernel install constava como planned) atualizados

## [0.4.4] - 2026-07-28

### Added

- **Slice Compose (experimental)** — fatia composta: product ∩ agent-BL ∩ sugestões do usuário
  - CLI: `arah slice plan -SliceId E1-Sx`
  - Script `scripts/agents/slice-compose.ps1` · skill `slice-compose`
  - Output: `docs/_arah/slice-plans/<slice>.md` (Product | Agent priorities | User | Deferred | Executor)
  - Docs [SLICE_COMPOSE.md](docs/SLICE_COMPOSE.md); testes `test-slice-compose.ps1`
- **Agent backlog `goal` / `acceptance`** — metas mensuráveis (ex. coverage ≥ 50% em `packages/domain`)
  - Seeds QA / test-architect para coverage MVP + sinal `coverage-tool`
- **Knowledge Graph (Graphify)** — integração opcional complementar ao Agent Graph
  - Avaliação/ADR: [`docs/architecture/GRAPHIFY.md`](docs/architecture/GRAPHIFY.md)
  - Spec-Id: `arah-graphify-knowledge-graph`
  - CLI `arah knowledge-graph [status|code-only|full]` + skill `graphify-knowledge-graph`
  - Manifesto `docs/_meta/knowledge-graph.manifest.yaml`; default `--code-only` (sem tokens)
  - `regenerate -IncludeKnowledgeGraph` ou `knowledge_graph.enabled` na config
  - Capability experimental `knowledge-graph-graphify`

## [0.4.3] - 2026-07-28

### Added

- **Repo Visions schema v2** — opinião por tipo de app, action plan, backlog próprio e memory/events
  - Sidecars `{agent}.backlog.yaml` (merge-safe) + `{agent}.events.yaml` (append-only)
  - IDs estáveis `{agent}-BL-NNN`; Refresh **não apaga** `done`/`cancelled`
  - CLI: `arah assess-repo -Refresh`, `arah vision update`, `arah backlog sync`
  - Lentes Clean Architecture / especialidade (TEA, ports, ASVS lite, boundaries, …)
  - Heurística `app_type` alimenta a seção Opinião
  - Docs [REPO_VISIONS.md](docs/REPO_VISIONS.md); testes de merge no `test-assess-repo.ps1`
- **Experimental — Repo Perspective Assessment** (`arah assess-repo` / alias `bootstrap-vision`)
  - Skill `repo-perspective-assess`
  - Script `scripts/agents/assess-repo.ps1` — As-Is + Gaps + To-Be por agent sob `.arah/visions/`
  - Lentes por papel (qa, solutions-architect, backend, security, test-architect, domains, …)
  - Modo `bootstrap-empty` para repo vazio
  - Fase de bootstrap serial documentada como exceção controlada ao ECP ([REPO_VISIONS.md](docs/REPO_VISIONS.md), [EXECUTION_CONTROL.md](docs/EXECUTION_CONTROL.md))
  - Opt-in nos next steps do `install` (não dispara automaticamente)
  - Testes `scripts/harness/test-assess-repo.ps1`

## [0.4.2] - 2026-07-24

### Added

- **Publicação autônoma de releases** — após merge em `main`, o workflow `release.yml` garante tag `vX.Y.Z` + GitHub Release a partir de `VERSION` (idempotente)
  - `scripts/agents/cut-release.ps1` + CLI `arah release cut`
  - Skill `release-cut` aponta para o script real
  - Humano só mergeia o bump; a Action publica (sem `git tag` manual)

### Changed

- [UPDATE_NOTIFICATIONS.md](docs/UPDATE_NOTIFICATIONS.md) — fluxo “merge → release autônomo → notify consumidores”

## [0.4.1] - 2026-07-19

### Added

- **Notificações de atualização do harness** (padrão Releases + cron no consumidor)
  - `arah update-check` / `scripts/agents/check-harness-update.ps1` — pin × GitHub Release
  - Workflow consumidor `harness-update-check.yml` (issue `arah-harness-update`)
  - Workflow upstream `release.yml` — publica Release a partir de tags `v*.*.*`
  - Template Renovate `templates/renovate-arah.json` (opcional)
  - Docs: [UPDATE_NOTIFICATIONS.md](docs/UPDATE_NOTIFICATIONS.md)
- `init` lê `VERSION` (não hardcode); migra `update_check` na config

## [0.4.0] - 2026-07-19

### Added

- **Execution Control Protocol** — terminalidade determinística (Spec-Id: `arah-execution-control`)
  - Schemas `execution-contract` + `consultation-result`
  - Ledger `.arah/local/execution/{active,completed,blocked}/`
  - Runtime `execute-task.ps1` / `task-control.ps1` + validator
  - Autonomia `execute_change` (rank 4); ranks `invoke_skill`+ deslocados
  - `execution_role` nos manifests; `execution.primary_executor` na coreografia
  - CLI `arah task create|status|validate|complete|block`
  - Cursor rule `.cursor/rules/arah-execution-control.mdc`
  - Docs [EXECUTION_CONTROL.md](docs/EXECUTION_CONTROL.md); testes `test-execution-control.ps1`
- Migração: `init`/`regenerate` adicionam `execution_control` sem sobrescrever overlays do consumidor
- `init` passa a distribuir `schemas/arah-harness/`

### Changed

- Fluxo documentado: intenção → um executor → consultas limitadas → done|blocked (não rede livre de handoffs)
- `spec_before_work` pula classe `trivial`; aplica-se a `execute_change`

### Migration notes

Consumidores existentes: rode `arah regenerate -UpdateKernel` (ou `update`) para receber scripts, rule Cursor e bloco `execution_control` na config. Customizações em `arah.config.yaml` são preservadas; apenas a seção ausente é acrescentada. Para desligar temporariamente: `execution_control.enabled: false`.

## [0.3.1] - 2026-07-19

### Added

- **ARAH Live Service** (`live/`) — Go read-only: REST + WebSocket, FS watch, SQLite descartável, adapter GitHub (C-01→C-06)
- **Schemas console** — `schemas/console/` (event, summary, gate-run, domain-health)
- **Live Console wired** — `website` tenta `127.0.0.1:8787` (ou `?api=` / `NEXT_PUBLIC_LIVE_API`); fallback mock
- **CLI Go fase 1** (`cmd/arah`) — `doctor`, `sync-check`, `version` (H-07); CI `.github/workflows/go.yml`
- **Economy Intelligence** — scorecard de eficiência do harness (Spec-Id: `arah-economy-metrics`)
- `arah metrics rollup|report` — agrega audit/live/signals → `.arah/observability/summary.yaml`
- Schema `metrics-summary`; campos opcionais de tokens/custo em `audit-event`
- Skill `metrics-rollup`; digest opcional `docs/_meta/metrics.digest.md` (`-Digest`)
- Evolve consome scorecard e pode propor `kind: economy`
- Docs: [ECONOMY.md](docs/ECONOMY.md)
- **Website** (`website/`) — Next.js bilíngue EN/PT: Home, Architecture, How It Works, TechOrganism, Use Cases, portal de docs + CLI explorer, Live Console (mock)
- Conteúdo extraído dos protótipos Control Plane em `website/content/`; CI `.github/workflows/website.yml`
- **Deploy GitHub Pages** — export estático (`output: 'export'`) + Actions (`deploy-pages`); URL `https://sraphaz.github.io/arah-harness/` (sem domínio próprio)
- **Estado quente × evidência fria** — `.arah/local/` (gitignored) + `docs/_meta/runs/*/summary.json`
- **Arquivo-por-evento** — bus/audit em `<ULID>.json`; `arah compact` e `arah migrate-state`
- **Scrubbing de secrets** antes de persistir payloads (`arah-event-io.ps1`)
- **`arah hooks install`** — pre-commit + [BRANCH_PROTECTION.md](docs/BRANCH_PROTECTION.md)
- **`install -Minimal`** — manifests + gates; upgrade path documentado
- **`capabilities.yaml`** — fonte única available/experimental/planned
- **ADR-001** — CLI portátil em Go
- Schemas signal/audit **0.2.0** com campo `v`; [SIGNAL_COMPATIBILITY.md](docs/SIGNAL_COMPATIBILITY.md), [STATE_MODEL.md](docs/STATE_MODEL.md), [CLI_SURFACE.md](docs/CLI_SURFACE.md)
- Backlog Control Plane: [docs/backlog/](docs/backlog/)
- Design handoff em [docs/design/control-plane/](docs/design/control-plane/)
- Spec `arah-state-model`

### Changed

- `record-agent-event` preserva scorecard rico (não sobrescreve metrics-summary)
- `regenerate` inclui passo metrics rollup
- Evolution schema: `kind: economy` + `based_on.metrics_semaphore`
- `signal-bus` / `record-agent-event` / `evolve-harness` / `metrics-rollup` leem pending+archive+legado
- Gate security também varre evidência fria e `.arah` (exceto `local/`)
- Organism `bus_path` default → `.arah/local/bus/`

## [0.3.0] - 2026-07-17

### Added

- **TechOrganism** — harness como organismo instalado no repositório
- `arah discover` — observa stack/domínio → `docs/_meta/discovery.proposed.yaml`
- `arah organism bootstrap|status|signal` — ritual ontogênico + barramento tipado
- `arah evolve` — self-learning a partir de audit/sinais/telemetria → `evolution.proposed.yaml`
- `arah regenerate` — homeostase: update + domain sync + discover + organism + evolve + graph + doctor
- Schemas: `discovery`, `organism`, `signal`, `evolution` em `schemas/arah-harness/`
- Skills: `discover-repo`, `evolve-harness`, `regenerate-harness`
- Docs: [TECHORGANISM.md](docs/TECHORGANISM.md); spec `arah-biocomponent`

### Changed

- Princípio de mercado: “agentes que criam agentes” → **proposta + Apply + PR** (sem spawn silencioso)
- CLI e METHOD documentam ciclo TechOrganism
- Consumidores atualizam com `regenerate -UpdateKernel` para receber a dimensão
- Documentação premium: README, TECHORGANISM, METHOD, INSTALL, BOOTSTRAP, AGENTS.md
- Brand: **ARAH Harness · TechOrganism**

## [0.2.3] - 2026-07-06

### Added

- **harness-model.schema.yaml** — domain agents, governance, observability e audit como first-class
- Domain agents: `clean-craft-advisor`, `test-architect`, `architecture-documenter`
- `.agents/autonomy.yaml` — níveis 0–6 e gates humanos
- Scripts de auditoria: `record-agent-event.ps1`, `check-autonomy.ps1`
- `harness-model-lib.ps1` — validação compartilhada em validate-specs e validate-agent-graph
- Docs: GOVERNANCE.md, OBSERVABILITY.md, AUDIT.md, HARNESS_PROFILES.md, MODEL.md
- `schemas/arah-harness/` — schemas canônicos incluindo audit-event e harness-model

### Changed

- Profiles declaram `model:` com domain agents e autonomia por tier
- validate-specs.ps1 e validate-agent-graph.ps1 rejeitam repos incompletos para o tier
- install-harness.ps1 escreve harness-profile.yaml com bloco model completo

## [0.2.2] - 2026-07-05

### Added

- **ARAH Live Session** — telemetria em `.cursor/arah-live/` (`state.json`, `events.jsonl`)
- Hooks Cursor expandidos: `sessionStart`, `sessionEnd`, `subagentStart`, `subagentStop`, `postToolUse` (Task), `afterFileEdit`, `turn-stop`
- `session-telemetry.ps1` — resolve coreografia ao vivo via `choreograph-agents.ps1`
- Extensão `extension/arah-live/` — painel lateral Cursor/VS Code com agentes pulsando em tempo real
- Extensão v0.1.1: grafo SVG com arestas regra→agente, lane specialists, status bar
- VSIX empacotável: `npm run package` em `extension/arah-live/`
- [docs/LIVE_SESSION.md](docs/LIVE_SESSION.md)

## [0.2.1] - 2026-07-05

### Added

- `validate-solution-choreography.ps1` — valida agentes runtime da solução (`runtime.path` em `arah.config.yaml`)

## [0.2.0] - 2026-07-04

### Added

- Kernel instalável (`arah init` / `update` / `doctor`)
- Domínios, specialists, coreografia por paths
- Gates e CI de manifests

## [0.1.0] - 2026-07-01

### Added

- Scaffold inicial do ARAH Harness

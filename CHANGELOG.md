# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

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

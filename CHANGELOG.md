# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

## [0.3.1] - 2026-07-19

### Added

- **Website** (`website/`) — Next.js bilíngue EN/PT: Home, Architecture, How It Works, TechOrganism, Use Cases, portal de docs + CLI explorer, Live Console (mock)
- Conteúdo extraído dos protótipos Control Plane em `website/content/`; CI `.github/workflows/website.yml`
- **Estado quente × evidência fria** — `.arah/local/` (gitignored) + `docs/_meta/runs/*/summary.json`
- **Arquivo-por-evento** — bus/audit em `<ULID>.json`; `arah compact` e `arah migrate-state`
- **Scrubbing de secrets** antes de persistir payloads (`arah-event-io.ps1`)
- **`arah hooks install`** — pre-commit + [BRANCH_PROTECTION.md](docs/BRANCH_PROTECTION.md)
- **`install -Minimal`** — manifests + gates; upgrade path documentado
- **`capabilities.yaml`** — fonte única available/experimental/planned
- **ADR-001** — CLI portátil em Go
- Schemas signal/audit **0.2.0** com campo `v`; [SIGNAL_COMPATIBILITY.md](docs/SIGNAL_COMPATIBILITY.md), [STATE_MODEL.md](docs/STATE_MODEL.md), [CLI_SURFACE.md](docs/CLI_SURFACE.md)
- Backlog Control Plane: [docs/backlog/](docs/backlog/) (W/C UI em espera; H parcialmente entregue)
- Design handoff em [docs/design/control-plane/](docs/design/control-plane/)
- Spec `arah-state-model`

### Changed

- `signal-bus` / `record-agent-event` / `evolve-harness` leem pending+archive+legado
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
- CLI `arah validate-runtime` — checagem de coreografia de entrega (manifests, co_activation, harness antes de draft)
- Parser `co_activation` em `choreography-parser.ps1` + expansão em `choreograph-agents.ps1`
- `domain sync` gera rules para **specialists** em `choreography.domains.yaml` (elimina órfãos nextjs/prisma)
- `doctor` executa `validate-runtime` quando configurado

### Fixed

- `choreograph-agents.ps1` trata `type: specialist` como consulta de domínio (path-based)
- `config-parser.ps1` lê bloco `runtime:` aninhado em `arah.config.yaml`
- **`arah update` preserva `arah.config.yaml` e `AGENTS.md`** (modo `-KernelOnly`)
- `Get-ArahObjectList` parseia `paths` de specialists via `Get-ArahListBlock` (fix nextjs/prisma órfãos)
- `Parse-ChoreographyRules` limita parsing à seção `rules:` (ignora `triggers` em runtime yaml)

## [0.2.0] - 2026-07-04

### Added

- Kernel ARAH: 11 agentes operacionais, 18 skills, coreografia, checklists, hooks Cursor
- CLI: `init`, `install`, `update`, `doctor`, `sync-check`, `domain sync`, `export-graph`
- `domain sync` — gera `.agents/domain/` e `choreography.domains.yaml` a partir de `arah.config.yaml`
- Agent Graph: `export-agent-graph.ps1`, `validate-agent-graph.ps1`, schema YAML
- Overlay `choreography*.yaml` — regras locais sem perder `arah update`
- Template CI `agents-validate.yml` instalado no `init`
- Documentação: METHOD, MARKET_REFERENCE, BOOTSTRAP, MIGRATION_FROM_ARAH, INSTALL
- Self-test script e workflow CI do próprio harness

### Fixed

- Parser `arah.config.yaml`: separação domains/specialists, paths vs references, blocos enrich/validate

### Proven in

- [IAutos](https://github.com/sraphaz/iautos) — legaltech monorepo (6 domínios consultivos)

## [0.1.0] - 2026-07-04

### Added

- Estrutura inicial do repositório e proof-of-concept `init` + `doctor`

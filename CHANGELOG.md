# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

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

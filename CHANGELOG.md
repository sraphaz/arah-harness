# Changelog

Formato baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/).

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

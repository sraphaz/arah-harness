# Roadmap — arah-harness

## Shipped · v0.4.4

- Execution Control, TechOrganism, Economy, Graphify (fase 0), Slice Compose
- CLI Go fase 1 (H-07): `doctor` / `sync-check` / `version`
- Estado quente × frio; arquivo-por-evento; capabilities.yaml
- Site / Live Console em andamento (`docs/backlog/`)

## Now · v0.5 — Runtime Cohesion

**Norte da minor:** coesão do runtime, não mais features agentic.  
ADR: [`docs/adr/002-runtime-cohesion-0.5.md`](adr/002-runtime-cohesion-0.5.md) ·  
Arquitetura: [`docs/architecture/RUNTIME_COHESION.md`](architecture/RUNTIME_COHESION.md) ·  
Spec: `arah-runtime-cohesion`

### P0

- **H-13** `arah-core` (Go) — modelo tipado Task/Run/Policy/Evidence/ExecutionContract
- **H-14** Pipeline único `plan → validate → apply` + dry-run + diff + idempotência
- **H-15** Kernel gerado — sync/verify + go:embed zip + `arah kernel install`
- **H-16** StateStore (SQLite WAL) + migração — **done**

### P1

- **H-17** MCP serve sobre os mesmos casos de uso da CLI
- **H-18** Evidence Graph determinístico
- **H-19** Timeline unificada por task/run + correlacionadores
- **H-20** Conformance suite + repositórios-fixture

Compat / estrangulamento: PowerShell como wrapper; priorizar no Go
`install` · `update` · `task *` · `policy check` · `graph export` · `evidence explain` · `mcp serve`.

## Later · v0.5.x → v1.0

- **H-08** Daemon `arahd` (P2)
- Knowledge providers plugáveis; full-text local; semântica opcional
- Produto Arah consome harness como dependência versionada
- **W** Site + portal docs · **C** Live Console MVP
- Control plane multi-repo (org) — só após contratos 0.5 estáveis
- Profiles enterprise / retention contractual

## Explicitamente adiado (protege coesão)

Novos tipos de agentes · RAG próprio · ontologia LLM · swarm irrestrito ·
expandir Graphify antes da porta `KnowledgeProvider` · dashboards antes de
contratos estáveis.

## Norte

**ARAH Harness · TechOrganism** como control plane de repositórios sérios:
runtime tipado e explicável, autonomia crescente, ledger intacto, humano no merge.

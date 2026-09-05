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
Spec: `arah-runtime-cohesion` (**AC-10 covered**)

### P0 — done

- **H-13** `arah-core` (Go) — Task/Policy/Evidence/ExecutionContract + briefing + context budget
- **H-14** Pipeline `plan → validate → apply` + dry-run + diff + idempotência
- **H-15** Kernel gerado — sync/verify + go:embed zip + `arah kernel install`
- **H-16** StateStore (SQLite WAL) + migração schema v3

### P1 — done (base 0.5)

- **H-17** MCP serve + progressive disclosure (`arah_get_task_context`)
- **H-18** Evidence Graph + `evidence explain`
- **H-19** Timeline com correlação (`run_id` / `correlation_id` / `agent_id`)
- **H-20** Conformance suite + fixtures

### Pós-close 0.5 (polish)

- Context budget medido (`arah economy context-compare`)
- OTel adapter opcional; `arahd` (H-08)
- PS strangulation restante (`install` / `update` / organism writes)

## Later · v0.5.x → v1.0

- **H-08** Daemon `arahd` (P2)
- Knowledge providers plugáveis; full-text local; semântica opcional
- Produto Arah consome harness como dependência versionada
- **W** Site + portal docs · **C** Live Console MVP (single-repo)
- Profiles enterprise / retention contractual

> **Fora de escopo permanente ([ADR-003](adr/003-kernel-workspace-boundary.md)):**
> control plane multi-repo / agregação org-level é responsabilidade do Surya Labs
> Workspace. O kernel expõe dados de um repo via contratos; nunca agrega portfólio.

## Explicitamente adiado (protege coesão)

Novos tipos de agentes · RAG próprio · ontologia LLM · swarm irrestrito ·
expandir Graphify antes da porta `KnowledgeProvider` · dashboards antes de
contratos estáveis.

## Norte

**ARAH Harness · TechOrganism** como engineering kernel de repositórios sérios:
runtime tipado e explicável, autonomia crescente, ledger intacto, humano no merge.
Escopo: **um repositório por instalação** — agregação de portfólio é do Workspace (ADR-003).

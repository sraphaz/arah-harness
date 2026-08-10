# ADR-002 — Runtime Cohesion como norte da v0.5

- **Status:** Proposed
- **Date:** 2026-08-10
- **Deciders:** maintainers (proposta agente; merge humano)
- **Supersedes / relates:** ADR-001 (CLI Go); estende sem invalidar

## Context

O Arah Harness 0.4.x já concentra visão de produto forte: Execution Control,
TechOrganism, Agent Graph, Knowledge Graph (Graphify), Economy, Live e gates.
O risco dominante **não** é falta de capacidade — é a superfície de
funcionalidades crescer mais rápido que o **núcleo de runtime**.

Comportamento relevante ainda está distribuído entre PowerShell, YAML,
manifests, templates e regras. O CLI Go (H-07) cobre só leitura
(`doctor` / `sync-check` / `version`). Há fontes duplicadas
(`.agents/` × `kernel/.agents/`, etc.) e estado quente baseado em
arquivo-por-evento, adequado hoje, mas frágil para CLI + MCP + extensão +
futuro daemon concorrentes.

Avaliou-se absorver o produto
[rafaelnicolett/kern](https://github.com/rafaelnicolett/kern) (RAG local +
ontologia incremental via MCP). **Não** copiamos RAG nem ontologia aprendida.
Absorvemos decisões de engenharia documentadas nos ADRs do kern:

| Kern | Arah Harness |
|------|--------------|
| ADR-0001 Hexagonal | `internal/core` + adapters (`fsstore`, `choreography`, `mcp`) |
| ADR-0007 MCP primary contract | `arah mcp serve` + `docs/architecture/mcp-tool-contract.md` |
| ADR-0004 SQLite local / per-project | StateStore port; filesystem agora, SQLite (H-16) em seguida |
| Envelope de erro tipado | Implementado de forma **mais estrita** que o kern (JSON `code`/`trace_id`) |
| Binário local coeso | CLI Go canônica para Execution Control |

## Options

| Opção | Coesão | Risco | Valor |
|-------|--------|-------|-------|
| **A — Runtime Cohesion (0.5)** — `arah-core` tipado em Go, portas/adapters, MCP, StateStore, contratos JSON, Evidence Graph | Alta | Médio (migração por estrangulamento) | Alto — previsibilidade |
| B — Mais features agentic (novos agentes, RAG, control plane multi-repo) | Baixa | Alto (mais superfície) | Médio curto prazo / baixo longo prazo |
| C — Reescrita abrupta PS → Go | Alta potencial | Muito alto (quebra consumidores) | Médio |
| D — Incorporar kernel externo (RAG/ontologia) | Baixa no Arah | Alto (dependência e drift de método) | Baixo para o harness |

## Decision

Adotar a **opção A**: a próxima versão minor do Arah Harness chama-se
**0.5 — Runtime Cohesion**.

Princípios:

1. **Não** é release de mais funcionalidades agentic; é release de coesão.
2. Consolidar domínio e invariantes em **`arah-core` (Go)** — ADR-001 permanece.
3. PowerShell migra por **estrangulamento** (wrappers), não big-bang.
4. **Kernel distribuído é artefato gerado**, nunca editado como segunda fonte.
5. Interfaces estáveis: **CLI e MCP** chamam os mesmos casos de uso.
6. Estado quente via **StateStore** (SQLite WAL alvo); Git continua canônico para
   config, manifests, specs e approvals.
7. **Evidence Graph** determinístico (schemas → arestas), sem LLM como fonte de verdade.
8. Deliberadamente **fora** de 0.5: RAG próprio, ontologia aprendida, swarm
   irrestrito, control plane multi-repo, novos tipos de agentes, expansão
   Graphify antes de `KnowledgeProvider`.

Detalhamento normativo: [`docs/architecture/RUNTIME_COHESION.md`](../architecture/RUNTIME_COHESION.md) ·
spec [`arah-runtime-cohesion`](../specs/arah-runtime-cohesion.spec.yaml).

## Consequences

### Positivas

- Uma implementação das invariantes do Execution Control / policy / evidência.
- Agentes operam por contrato MCP estável, sem conhecer dezenas de scripts.
- Drift kernel↔fonte eliminável por geração + hashes.
- Observabilidade unificada por `task_id` / `run_id` / `trace_id`.
- Conformance suite prova comportamento do harness, não só scripts isolados.

### Negativas / trade-offs

- Esforço de migração PS → Go ao longo de 0.5.x.
- Dual-run (PS + Go) durante a transição exige testes de paridade.
- Introdução de SQLite muda o modelo operacional local (ainda isolado por repo).
- Escopo 0.5 exige disciplina de **não** abrir épicos agentic paralelos.

### Não-consequências (explícito)

- Não substitui o método ARAH nem o TechOrganism.
- Não torna Graphify obrigatório; torna-o um adapter de `KnowledgeProvider`.
- Não exige daemon (`arahd`) para um repo funcionar (H-08 permanece P2).

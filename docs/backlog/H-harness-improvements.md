# Épico H — Melhorias do harness

**Meta:** atacar overhead de escrita em arquivos e riscos ALTA antes de escalar.  
**Referência:** [Análise Técnica](../design/control-plane/design-files/Analise%20Tecnica.dc.html) §3–6.

| ID | Item | Status |
|----|------|--------|
| H-01 | Auditoria de superfície da CLI | **done** — `docs/CLI_SURFACE.md` |
| H-02 | Congelar e versionar tipos de sinal | **done** — schema `v` + compat |
| H-03 | Separação estado quente × evidência fria | **done** — `.arah/local/` + migrate |
| H-04 | Arquivo-por-evento + compactação | **done** — ULID + `arah compact` |
| H-05 | Scrubbing de secrets na evidência | **done** — scrub antes de persistir |
| H-06 | ADR linguagem da CLI portátil | **done** — ADR-001 (Go) |
| H-07 | CLI binária portátil (fase 1) | **done** — `cmd/arah` doctor/sync-check/version |
| H-08 | Daemon `arahd` (opcional) | **backlog** — médio prazo |
| H-09 | Pre-commit hooks + branch protection | **done** — `arah hooks install` |
| H-10 | Fonte única de status de capacidades | **done** — `capabilities.yaml` |
| H-11 | Modo mínimo de adoção | **done** — `install -Minimal` |
| H-12 | Knowledge Graph (Graphify) | **fase 0 done** — adapter opcional; fases 1–3 backlog |
| H-13 | `arah-core` tipado (Go) | **done** — domain + briefing + context budget + consultations |
| H-14 | Contratos JSON + `plan→validate→apply` | **done** — envelope + dry-run + diff + idempotência |
| H-15 | Kernel gerado (fim da segunda fonte) | **done** — sync/verify + go:embed zip + `arah kernel install` |
| H-16 | StateStore SQLite + migração | **done** — runtime.db WAL + schema v3 correlação |
| H-17 | MCP serve (mesmos use cases da CLI) | **done** — tools ECP + context + route + evidence + consultations |
| H-18 | Evidence Graph determinístico | **done** — graph + explain + event edges |
| H-19 | Timeline unificada task/run | **done** (base) — correlacionadores no Event; OTel later |
| H-20 | Conformance suite + fixtures | **done** — fixtures + CLI↔MCP + AC-10 |

**Direção 0.5:** [ADR-002](../adr/002-runtime-cohesion-0.5.md) · [RUNTIME_COHESION.md](../architecture/RUNTIME_COHESION.md)

---

### H-01 · Auditoria de superfície da CLI `P` — done
Tabela de gaps em [`docs/CLI_SURFACE.md`](../CLI_SURFACE.md). Comandos novos: `compact`, `migrate-state`, `hooks`.

### H-02 · Congelar e versionar tipos de sinal `M` — done
Enum estável + campo `v` no payload; [`docs/SIGNAL_COMPATIBILITY.md`](../SIGNAL_COMPATIBILITY.md). Bloqueia C-01.

### H-03 · Separação estado quente × evidência fria `G` — done
- `.arah/local/` (gitignored): telemetria e sinais operacionais  
- Versionado: `docs/_meta/runs/<run-id>/summary.json` + decisões  
- `arah migrate-state` · docs [`STATE_MODEL.md`](../STATE_MODEL.md)

### H-04 · Arquivo-por-evento + compactação `G` — done
Escrita `<ULID>.json` atômica; `arah compact` funde em JSONL por período.

### H-05 · Scrubbing de secrets na evidência `M` — done
Redação antes de persistir; gate security também varre `.arah/**` (quando versionado).

### H-06 · Decisão: linguagem da CLI portátil `P` — done
[`docs/adr/001-portable-cli-language.md`](../adr/001-portable-cli-language.md) — Go.

### H-07 · CLI binária portátil (fase 1) `G` — done
Binário em [`cmd/arah/`](../../cmd/arah/): `doctor`, `sync-check`, `version` com exit codes 0/1/2/4/10. PowerShell permanece canônico para fluxos de escrita/organismo; `export-graph` segue no PS até fase 2.

### H-08 · Daemon `arahd` (opcional) `G` — backlog
Watch + batch + stream WS; CLI degrada sem daemon.

### H-09 · Pre-commit hooks + branch protection guide `P` — done
`arah hooks install`; guia em [`docs/BRANCH_PROTECTION.md`](../BRANCH_PROTECTION.md).

### H-10 · Fonte única de status de capacidades `P` — done
[`capabilities.yaml`](../../capabilities.yaml) — available / experimental / planned.

### H-11 · Modo mínimo de adoção `M` — done
`arah install -Minimal`: manifests + gates; upgrade path em INSTALL.md.

### H-12 · Knowledge Graph via Graphify `M` — fase 0 done / fases 1–3 backlog
Adapter opcional (`arah knowledge-graph`, skill, manifesto). Veredito e fases em [`docs/architecture/GRAPHIFY.md`](../architecture/GRAPHIFY.md).

- [x] Fase 0 — ADR + adapter `--code-only` + CLI + capability experimental  
- [ ] Fase 1 — discover/evolve consomem comunidades Graphify  
- [ ] Fase 2 — C-12 dual-pane no Live Console  
- [ ] Fase 3 — MCP `graphify.serve` opcional  
- **Nota 0.5:** fases 1–3 só após porta `KnowledgeProvider` (H-13); Graphify vira adapter, não exceção.

---

## Épico H · 0.5 Runtime Cohesion

### H-13 · `arah-core` (Go) — **done** P0
Domínio tipado em `internal/core`: Task/Contract, transitions, evidence, TaskService,
briefing (`BRIEFING.md`), consultas tipadas, context budget.
Ports: StateStore, EventStore, BriefingWriter, ConsultationWriter, ChoreographyResolver.
Adapters `fsstore`, `choreography`, `sqlitestore`, `mcp`.

### H-14 · Envelope JSON + pipeline único — **done** P0
`--json` com `ok` / `code` / `message` / `trace_id` / `details` / `remediation`
(`internal/envelope`). Mutações aceitam `--dry-run` / `dry_run` (plan sem persistir).
Resposta inclui `diff` textual e `idempotent` (re-complete/re-block com a mesma
evidência/razão não persiste de novo).

### H-15 · Kernel gerado — **done** P0
Fonte canônica na raiz → `arah kernel sync` → `kernel/` + embed zip.
`arah kernel verify` + CI. `arah kernel install` extrai go:embed.
Capability `generated-kernel-package` experimental.

### H-16 · StateStore — done P0
Hot state em `.arah/local/runtime.db` (SQLite WAL). Schema v3 com colunas de
correlação em `task_events`.

### H-17 · MCP primeira classe — **done** P1
`arah mcp serve` + [mcp-tool-contract.md](../architecture/mcp-tool-contract.md).
Tools: capabilities, get/create/complete/block task, timeline, **task_context**,
**explain_route**, **get_evidence**, **submit_consultation**, evidence graph.
`dry_run` nas mutações.

### H-18 · Evidence Graph — **done** P1
Grafo determinístico (`internal/evidence`). Eventos → `validated_by` / `invokes`.
CLI `arah evidence graph|explain`. IDs estáveis via SHA-256.

### H-19 · Timeline por Run — **done** P1 (base)
`arah task timeline` + EventStore com `run_id` / `correlation_id` / `agent_id` /
`session_id`. OTel adapter permanece later.

### H-20 · Conformance suite — **done** P1
`internal/conformance` + fixtures (`valid-minimal`, `invalid-config`, `monorepo`,
`task-blocked`, …). Provas: dry-run, error codes, CLI↔MCP create/complete/block,
transições, briefing, context budget, kernel dogfood, correlação de eventos.
Spec AC-10 covered.

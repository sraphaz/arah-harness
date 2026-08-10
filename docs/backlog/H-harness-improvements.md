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
| H-13 | `arah-core` tipado (Go) | **in progress** — domain + task CLI + MCP stdio |
| H-14 | Contratos JSON + `plan→validate→apply` | **partial** — envelope JSON + invariantes; dry-run/diff TBD |
| H-15 | Kernel gerado (fim da segunda fonte) | **backlog 0.5 P0** |
| H-16 | StateStore SQLite + migração | **done** — `.arah/local/runtime.db` WAL + mirror YAML |
| H-17 | MCP serve (mesmos use cases da CLI) | **in progress** — tools ECP + timeline + evidence graph |
| H-18 | Evidence Graph determinístico | **in progress** — `arah evidence graph` + MCP tool |
| H-19 | Timeline unificada task/run | **partial** — `task timeline` + EventStore; correlacionadores amplos TBD |
| H-20 | Conformance suite + fixtures | **backlog 0.5 P1** |

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

### H-13 · `arah-core` (Go) — in progress P0
Domínio tipado em `internal/core`: Task/Contract, transitions, evidence, TaskService.
Ports no core; adapters `fsstore`, `choreography`, `mcp`. Inspirado em kern ADR-0001.

### H-14 · Envelope JSON + pipeline único — partial P0
`--json` com `ok` / `code` / `message` / `trace_id` / `details` / `remediation`
(`internal/envelope`). Dry-run/diff idempotente ainda TBD.

### H-15 · Kernel gerado — backlog P0
Fonte canônica → validação → pacote + hashes. Preferir `go:embed`. Eliminar edição
manual de `kernel/.agents` / `kernel/.skills` / `kernel/scripts` como segunda fonte.

### H-16 · StateStore — done P0
Hot state em `.arah/local/runtime.db` (SQLite WAL, `modernc.org/sqlite`).
Migração automática a partir de YAML; mirror filesystem para PS. EventStore
`task_events` append-only (timeline).

### H-17 · MCP primeira classe — in progress P1
`arah mcp serve` + [mcp-tool-contract.md](../architecture/mcp-tool-contract.md).
Tools: capabilities, get/create/complete/block task, timeline, evidence graph.

### H-18 · Evidence Graph — in progress P1
Grafo determinístico schemas→arestas (`internal/evidence`). CLI
`arah evidence graph`. Completa Agent ∪ Knowledge.

### H-19 · Timeline por Run — backlog P1
Correlacionadores `task_id` / `run_id` / `trace_id` / …; timeline create→done|blocked.
OTel opcional.

### H-20 · Conformance suite — backlog P1
Fixtures (vazio, web, API, monorepo, drift, gate pendente, …) + provas de
install/update/migração/paridade CLI≡MCP/error codes. Dogfood em arah-harness.

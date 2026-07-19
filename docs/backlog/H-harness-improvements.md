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

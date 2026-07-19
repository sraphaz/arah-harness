# Modelo de estado — quente × frio

**Versão:** 1.0 · **Data:** 2026-07-19

O passivo técnico do harness era versionar **todo** estado operacional (bus, ledger, telemetria). A partir de 0.3.1 separamos:

| Camada | O quê | Onde | Git |
|--------|-------|------|-----|
| **Quente** | Sinais operacionais, auditoria bruta, telemetria, **contratos de execução** | `.arah/local/` (+ `execution/`) | **Ignorado** |
| **Frio** | Resumo compacto por run + decisões | `docs/_meta/runs/<run-id>/summary.json` | Versionável |
| **Contratos (versão)** | Manifests, schemas, approvals humanas | `docs/_meta/`, `.arah/approvals.yaml`, `schemas/` | Versionável |

## Layout

```text
.arah/
  local/                      # HOT — gitignored
    bus/pending/<ULID>.json   # arquivo-por-evento (atômico)
    bus/archive/YYYY-MM.jsonl # compactado
    audit/pending/<ULID>.json
    audit/archive/YYYY-MM.jsonl
    execution/                # Execution Control Protocol
      active/<task-id>.yaml
      completed/
      blocked/
      <task-id>/consultations/
  observability/summary.yaml  # agregado quente (gitignored)
  organism/state.json         # runtime (gitignored)
  approvals.yaml              # gates humanos (versionável)
docs/_meta/runs/<run-id>/
  summary.json                # COLD evidence
```

## Comandos

```powershell
# Migrar JSONL legado → pending + summary frio
powershell -File cli/arah.ps1 migrate-state

# Compactar pending → archive (retenção default 90 dias)
powershell -File cli/arah.ps1 compact [-Kind all|bus|audit] [-RetainDays 90]
```

## Escrita

- Cada evento = **criar arquivo** `<ULID>.json` (sem append concorrente no mesmo arquivo).
- Payloads passam por **scrubbing** de secrets antes do disco.
- `evolve` e leitores agregam pending + archive + legado.

## Por quê

Append em JSONL versionado gera crescimento de repo, conflitos de merge e ruído em PRs. Evidência deve ser **anexa** à decisão (summary / PR), não entrelaçada ao diff de código.

Relacionado: [AUDIT.md](AUDIT.md), [OBSERVABILITY.md](OBSERVABILITY.md), [SIGNAL_COMPATIBILITY.md](SIGNAL_COMPATIBILITY.md).

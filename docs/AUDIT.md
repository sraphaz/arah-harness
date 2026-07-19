# AUDIT — ARAH Harness

**Versão**: 1.0 · **Data**: 2026-07-06

## Trilha de auditoria

Estado quente (gitignored) — arquivo-por-evento:

```
.arah/local/audit/pending/<ULID>.json
.arah/local/audit/archive/YYYY-MM.jsonl   # após arah compact
```

Legado (ainda lido; migrar com `arah migrate-state`):

```
.arah/audit/events.jsonl
```

Cada evento (JSON):

```json
{
  "v": 1,
  "ts": "2026-07-06T22:00:00Z",
  "correlation_id": "a1b2c3d4e5f6",
  "project": "meu-projeto",
  "agent_id": "pr-steward",
  "action": "skill.invoke",
  "autonomy_level": "invoke_skill",
  "outcome": "ok",
  "human_gate": null,
  "details": ""
}
```

Schema: [`schemas/arah-harness/audit-event.schema.yaml`](../schemas/arah-harness/audit-event.schema.yaml)  
Modelo: [`STATE_MODEL.md`](STATE_MODEL.md) · Scrubbing antes do disco.

## Outcomes

| Valor | Significado |
|-------|-------------|
| `ok` | Ação permitida e concluída |
| `blocked` | Autonomia ou gate bloqueou |
| `denied` | Humano negou gate |
| `error` | Falha na execução |
| `pending` | Aguardando gate humano |

## Registrar

```powershell
./scripts/agents/record-agent-event.ps1 -AgentId backend -Action session.write -Outcome ok
```

Bloqueios automáticos via `check-autonomy.ps1`:

```powershell
./scripts/agents/check-autonomy.ps1 -AgentId release -Action release.cut
# → registra outcome: blocked se gate pendente
```

## Resumo agregado

`.arah/observability/summary.yaml` — atualizado a cada evento:

```yaml
total_events: 42
last_agent: qa
last_action: skill.invoke
last_outcome: ok
```

## Gates humanos

Aprovações em `.arah/approvals.yaml`:

```yaml
spec_before_work:
  status: approved
  by: raphael
  at: 2026-07-06T20:00:00Z
release_approval:
  status: pending
```

## Retenção

- Hot state: `.arah/local/` sempre gitignored; `arah compact -RetainDays 90`
- Cold evidence: `docs/_meta/runs/<run-id>/summary.json` (versionável)
- Enterprise profile: retenção contractual (ver `harness/profiles/enterprise.yaml`)

## Consultar eventos

```powershell
Get-ChildItem .arah/local/audit/pending -Filter *.json | Select-Object -Last 20
powershell -File ./scripts/agents/signal-bus.ps1 -List   # bus
# ou migrate/compact e ler archive/*.jsonl
```

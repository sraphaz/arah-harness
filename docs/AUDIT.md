# AUDIT — ARAH Harness

**Versão**: 1.0 · **Data**: 2026-07-06

## Trilha de auditoria

Append-only por repositório:

```
.arah/audit/events.jsonl
```

Cada linha (JSON):

```json
{
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

- Desenvolvimento: local, gitignored opcional
- Enterprise profile: definir retenção com cliente (ver `harness/profiles/enterprise.yaml`)

## Consultar eventos

```powershell
Get-Content .arah/audit/events.jsonl | Select-Object -Last 20
Get-Content .arah/audit/events.jsonl | ConvertFrom-Json | Where-Object { $_.outcome -eq 'blocked' }
```

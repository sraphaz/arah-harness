# HARNESS MODEL — ARAH Harness

**Versão**: 0.1.0 · **Data**: 2026-07-06

O **harness-model** define o que todo repo governado declara — domain agents, governança, observabilidade e auditoria são **first-class**, não add-ons opcionais.

Schema: [`schemas/arah-harness/harness-model.schema.yaml`](../schemas/arah-harness/harness-model.schema.yaml)

## Blocos do modelo

### `domain_agents`

Slots consultivos com obrigatoriedade por tier:

| Agente | Função | Max autonomia |
|--------|--------|---------------|
| `clean-craft-advisor` | SOLID, Uncle Bob, boundaries | `consult` |
| `test-architect` | TEA, risk-based, gates CI | `consult` |
| `architecture-documenter` | C4, ADRs, jornadas | `consult` |

Manifest em `.agents/domain/{id}.agent.yaml`.

### `governance`

```yaml
governance:
  autonomy_source: .agents/autonomy.yaml
  autonomy_levels: [observe, consult, route, activate, invoke_skill, side_effect, public]
  human_gates: [spec_before_work, release_approval]
  blocked_actions: [merge, approve, force_push]
```

Níveis 0–6 — ver [GOVERNANCE.md](GOVERNANCE.md).

### `observability`

```yaml
observability:
  diagnostics: .cursor/arah-live/diagnostics.jsonl
  session_traces: .cursor/arah-live/sessions/
  metrics_summary: .arah/observability/summary.yaml
  agent_graph: docs/_meta/agent-graph.generated.json
```

Ver [OBSERVABILITY.md](OBSERVABILITY.md).

### `audit`

```yaml
audit:
  event_schema: arah-harness/audit-event
  ledger_path: .arah/audit/events.jsonl
  retention_policy: project_default  # local_dev | contractual
  record_script: scripts/agents/record-agent-event.ps1
```

Ver [AUDIT.md](AUDIT.md).

## Onde o modelo vive

| Artefato | Função |
|----------|--------|
| `harness/profiles/*.yaml` | Modelo por tier (fonte de instalação) |
| `harness-profile.yaml` | Modelo instalado no repo-alvo |
| `schemas/arah-harness/harness-profile.schema.yaml` | Contrato do arquivo instalado |

## Obrigatoriedade por tier

| Tier | Domain agents required |
|------|------------------------|
| `minimal` | architecture-documenter |
| `consulting` | todos os 3 |
| `product` | todos os 3 |
| `enterprise` | todos os 3 + audit contractual |
| `open-source` | architecture-documenter + clean-craft-advisor |

## Validação

```powershell
./harness/scripts/validate-specs.ps1 -Target .
./harness/scripts/validate-agent-graph.ps1 -Target .
./scripts/harness/validate-agent-graph.ps1   # self-hosted arah-harness
```

Ambos invocam `harness-model-lib.ps1` — rejeitam repos incompletos para o tier instalado.

## Biocomponente (extensão do modelo)

Desde v0.3 o harness declara também artefatos de organismo (fora do profile tier, por repo):

| Artefato | Schema |
|----------|--------|
| `docs/_meta/discovery.proposed.yaml` | arah-harness/discovery |
| `docs/_meta/organism.manifest.yaml` | arah-harness/organism |
| `.arah/bus/signals.jsonl` | arah-harness/signal |
| `docs/_meta/evolution.proposed.yaml` | arah-harness/evolution |

Ver [BIOCOMPONENT.md](BIOCOMPONENT.md).

## Referências

- [HARNESS_PROFILES.md](HARNESS_PROFILES.md)
- [GOVERNANCE.md](GOVERNANCE.md)
- [BIOCOMPONENT.md](BIOCOMPONENT.md)

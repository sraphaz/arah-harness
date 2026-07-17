# OBSERVABILITY — ARAH Harness

**Versão**: 1.1 · **Data**: 2026-07-17

## Camadas de observabilidade

| Camada | Local | Conteúdo |
|--------|-------|----------|
| Auditoria | `.arah/audit/events.jsonl` | Eventos de agente (append-only) |
| Scorecard | `.arah/observability/summary.yaml` | Economy Intelligence (`metrics-summary`) |
| Digest | `docs/_meta/metrics.digest.md` | Resumo humano opcional (`-Digest`) |
| Live (Cursor) | `.cursor/arah-live/diagnostics.jsonl` | Diagnósticos de sessão |
| Sessões | `.cursor/arah-live/sessions/*.diagnostics.jsonl` | Traces por conversa |
| Agent Graph | `docs/_meta/agent-graph.generated.json` | Grafo exportável |

## Economy Intelligence

```powershell
powershell -File cli/arah.ps1 metrics rollup
powershell -File cli/arah.ps1 metrics report
```

Agrega audit + signals + live → rates, semaphore (`productive|neutral|expensive|insufficient_data`) e `roi_hints`.  
Guia completo: [ECONOMY.md](ECONOMY.md).

## Telemetria de sessão

Hook Cursor → `scripts/agents/session-telemetry.ps1`:

```powershell
./scripts/agents/session-telemetry.ps1 -Action turn-stop
./scripts/agents/session-telemetry.ps1 -Action choreography-resolve
```

Eventos espelhados em `.cursor/arah-live/events.jsonl`.

## Registrar eventos manualmente

```powershell
./scripts/agents/record-agent-event.ps1 -AgentId qa -Action skill.invoke -Outcome ok
./scripts/agents/record-agent-event.ps1 -AgentId release -Action release.cut -Outcome blocked -Blocked
```

## Exportar Agent Graph

```powershell
./scripts/agents/export-agent-graph.ps1
./scripts/harness/validate-agent-graph.ps1 -Json
```

Artefato: `docs/_meta/agent-graph.generated.json` — nós (agentes, skills, rules, gates) e arestas.

## Coreografia observável

```powershell
./scripts/agents/choreograph-agents.ps1 -ChangedFiles docs/architecture/c4.md -Json
```

Retorna: rules matched, agentes operacionais, consultas de domínio, skills, gates pendentes.

## Integração CI

Workflow `.github/workflows/agents-validate.yml` roda:

- `validate-agent-graph.ps1`
- `validate-manifests.ps1`
- `validate-specs.ps1` (quando aplicável)

## Privacidade

Diagnósticos locais em `.cursor/arah-live/` — não commitar por padrão (ver `.gitignore`).

Ver também: [LIVE_SESSION.md](LIVE_SESSION.md), [AUDIT.md](AUDIT.md).

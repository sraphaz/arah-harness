# Economy Intelligence — ARAH Harness

**Versão**: 0.3.1 · **Spec-Id**: `arah-economy-metrics`

O harness deixa de ser só política de “economia de tokens” (comunicação passiva) e passa a **medir** eficiência do trabalho agentic — para o humano e para o TechOrganism.

## Promessa

```text
append local → rollup → score transparente → evolve propõe → humano aplica
```

Sem swarm de agentes falando de métricas. Sem auto-merge por score.

## Comandos

```powershell
powershell -File cli/arah.ps1 metrics rollup [-Target .] [-Last 500] [-Digest]
powershell -File cli/arah.ps1 metrics report [-Target .] [-Last 500]
```

| Artefato | Path | Versionar? |
|----------|------|------------|
| Scorecard | `.arah/observability/summary.yaml` | Não (runtime) |
| Digest humano | `docs/_meta/metrics.digest.md` | Opcional (`-Digest`) |
| Ledger | `.arah/audit/events.jsonl` | Não |

## Scorecard (semaphore)

| Semaphore | Significado |
|-----------|-------------|
| `productive` | Alta taxa de ok, pouca fricção, ciclo TechOrganism em uso |
| `neutral` | Atividade sem extremos — monitore tendências |
| `expensive` | blocked/error altos — harness pode custar mais do que entrega |
| `insufficient_data` | <5 eventos na janela — ainda cedo para julgar |

Indicadores-chave: `blocked_rate`, `ok_rate`, `error_rate`, `tokens_*` (quando houver), `roi_hints`.

## Instrumentação (M2)

Campos opcionais em `record-agent-event.ps1` / audit-event:

`SessionId`, `SkillId`, `Model`, `TokensIn`, `TokensOut`, `LatencyMs`, `CostUsd`

Sem esses campos, o scorecard usa **proxies** (outcomes, turns, signals) — já útil no dia 1.

```powershell
./scripts/agents/record-agent-event.ps1 `
  -AgentId qa -Action skill.invoke -Outcome ok `
  -SkillId craft-review -TokensIn 4200 -TokensOut 900 -CostUsd 0.04
```

## Evolve (M3)

`arah evolve` chama rollup (fail-open) e pode emitir propostas `kind: economy` com evidence do scorecard (ex.: semaphore expensive, tokens não observados, turn storm).

## Fases

| Fase | Entrega |
|------|---------|
| **M1** | Rollup + report com proxies |
| **M2** | Tokens/custo opcionais no audit |
| **M3** | Heurísticas no evolve + digest |

## Context budget (0.5)

O agente **não** deve receber AGENTS.md + contrato completo + histórico em todo turno.
Use progressive disclosure (inspirado nas harness APIs do NOOA):

```bash
arah task context --task-id task-… --budget minimal|standard|full --json
arah economy context-compare [--task-id task-…] --json
```

MCP: `arah_get_task_context` com `budget`.

| Budget | Conteúdo | Uso |
|--------|----------|-----|
| `minimal` | id, state, executor, objective truncado | roteamento / status |
| `standard` | + paths (cap), limits, event kinds recentes | execução típica |
| `full` | contrato + timeline + BRIEFING.md | debugging / handoff |

Proxy de tokens: `chars/4` (determinístico). Comparativo before/after em
`docs/_meta/runs/context-budget/summary.json`.

## O que não fazer

- Chat entre agentes sobre métricas a cada turn
- Auto-merge por semaphore
- Depender 100% de billing do Cursor no dia 1

Ver também: [OBSERVABILITY.md](OBSERVABILITY.md), [AUDIT.md](AUDIT.md), [TECHORGANISM.md](TECHORGANISM.md), [METHOD.md](METHOD.md).

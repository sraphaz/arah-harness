# Superfície da CLI — auditoria (H-01)

**Fonte:** `cli/arah.ps1` · **Data:** 2026-07-19

## Comandos expostos

| Comando | Flags principais | Script |
|---------|------------------|--------|
| `install` | `-Target` `-ProjectName` `-Force` `-Minimal` | `cli/install.ps1` |
| `init` | `-Target` `-ProjectName` `-Force` `-Minimal` | `cli/init.ps1` |
| `update` | `-Target` `-Force` | `cli/update.ps1` |
| `doctor` | `-Target` | `cli/doctor.ps1` |
| `sync-check` | `-Target` | `cli/sync-check.ps1` |
| `domain sync` | `-Target` `-DryRun` | `domain-sync.ps1` |
| `export-graph` | `-Target` | `export-agent-graph.ps1` |
| `validate-runtime` | `-Target` | `validate-solution-choreography.ps1` |
| `discover` | `-Target` `-Apply` `-DryRun` | `discover-repo.ps1` |
| `organism bootstrap\|status\|signal` | signal: `-From` `-SignalType` `-SignalTo` `-Topic` `-Payload` | `organism-bootstrap` / `signal-bus` |
| `evolve` | `-Target` `-Apply` `-DryRun` | `evolve-harness.ps1` |
| `metrics rollup|report` | `-Last` `-Digest` `-DryRun` | `metrics-rollup.ps1` |
| `regenerate` | `-UpdateKernel` `-Force` `-ApplyDiscovery` `-SkipDoctor` `-DryRun` | `cli/regenerate.ps1` |
| `compact` | `-Kind` `-RetainDays` `-DryRun` | `compact-state.ps1`
| `migrate-state` | `-DryRun` | `migrate-state.ps1` |
| `hooks install` | `-Target` `-Force` | `install-hooks.ps1` |

## CLI Go (fase 1 — H-07)

Binário paralelo em [`cmd/arah/`](../cmd/arah/) (não substitui `cli/arah.ps1`):

| Comando | Notas |
|---------|-------|
| `arah doctor -target` | Checks mínimos de layout |
| `arah sync-check -target` | Drift vs `.arah-version` / kernel |
| `arah version` | `0.3.1-phase1` |

Live Service (épico C): [`live/cmd/arah-live`](../live/README.md) — REST+WS read-only.

## Gaps conhecidos (aceitos)

| Gap | Notas |
|-----|-------|
| `signal-bus -List` | Disponível no script; use Live Console (`/api/feed`) ou o script |
| Paridade total Go | Write/organism ainda PowerShell; `export-graph` na fase 2 |
| `arahd` | H-08 backlog |

## Exit codes (alvo para H-07)

| Code | Significado |
|------|-------------|
| 0 | OK |
| 1 | Erro genérico / validação |
| 2 | Drift / sync-check |
| 3 | Gate bloqueado |
| 4 | Doctor unhealthy |
| 10 | Uso incorreto (flags) |

Hoje a maioria dos scripts usa 0/1; paridade completa chega com a CLI Go.

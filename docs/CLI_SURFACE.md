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
| `task create` | `-Objective` `-Area` `-Class` | `task-control.ps1` / `execute-task.ps1` |
| `task status\|validate\|complete\|block` | `-TaskId` (`-Evidence` / `-Reason`) | `task-control.ps1` |

## Gaps conhecidos (aceitos)

| Gap | Notas |
|-----|-------|
| `signal-bus -List` | Disponível no script; não espelhado na CLI (use script ou Live Console futuro) |
| Bash parity | ROADMAP / ADR-001 — CLI portátil em Go |
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

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
| `knowledge-graph` | `[status\|code-only\|full]` `-Require` `-DryRun` | `graphify-knowledge-graph.ps1` |
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
| `update-check` | `-Notify` `-FailIfOutdated` `-LatestVersion` | `check-harness-update.ps1` |
| `release cut` | `-DryRun` | `cut-release.ps1` |
| `assess-repo` / `bootstrap-vision` | `-OutDir` `-Agents` `-Force` `-Refresh` `-BacklogOnly` `-DryRun` | `assess-repo.ps1` (**experimental** v2) |
| `vision update` | mesmo que assess-repo `-Refresh` | merge backlog + evento |
| `backlog sync` | assess-repo `-Refresh -BacklogOnly` | merge backlog |
| `slice plan` | `-SliceId` `-Suggestions` `-VisionDir` `-OutDir` `-Executor` `-Force` `-DryRun` | `slice-compose.ps1` (**experimental**) |

## CLI Go — arah-core (0.5 foundation)

Módulo raiz `github.com/sraphaz/arah-harness` · binário [`cmd/arah`](../cmd/arah/):

| Comando | Notas |
|---------|-------|
| `arah doctor -target [--json]` | Checks mínimos de layout |
| `arah sync-check -target [--json]` | Drift vs `.arah-version` / graph |
| `arah version [--json]` | `0.5.0-dev` |
| `arah task create\|status\|complete\|block` | Execution Control via `internal/core` |
| `arah mcp serve` | MCP stdio — ver [mcp-tool-contract.md](architecture/mcp-tool-contract.md) |

CLI e MCP compartilham os mesmos use cases. PowerShell permanece canônico para
organism/discover/evolve/regenerate até migração por estrangulamento.

Inspiração de engenharia: [rafaelnicolett/kern](https://github.com/rafaelnicolett/kern)
(hexagonal + MCP; sem RAG/ontologia).

## Gaps conhecidos (aceitos)

| Gap | Notas |
|-----|-------|
| `signal-bus -List` | Disponível no script; use Live Console (`/api/feed`) ou o script |
| Paridade total Go | Organism/write amplo ainda PowerShell; task ECP já no Go |
| SQLite StateStore | H-16 — filesystem adapter agora |
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

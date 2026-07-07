# Agent Operation

Operação detalhada além do resumo em `AGENTS.md`.

## Autonomia

Níveis declarados em `.agents/autonomy.yaml`. Domain agents ficam no máximo em `consult`.

## Auditoria

Eventos append-only em `.arah/audit/events.jsonl` via `scripts/agents/record-agent-event.ps1`.

## Observabilidade

- Diagnósticos: `.cursor/arah-live/diagnostics.jsonl`
- Sessões: `.cursor/arah-live/sessions/`
- Resumo: `.arah/observability/summary.yaml`

## Doctor

```powershell
$env:ARAH_HARNESS_PATH = "C:\path\to\arah-harness"
& "$env:ARAH_HARNESS_PATH\harness\scripts\doctor-harness.ps1" -Target .
```

## CI

Workflows `validate-specs.yml` e `harness-checks.yml` rodam em PR. Configure `vars.ARAH_HARNESS_PATH` no GitHub para drift check (opcional).

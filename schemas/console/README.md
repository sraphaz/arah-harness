# schemas/console/

Contratos de dados lidos pelo **ARAH Live Console** (épico C).

O console é **somente-leitura**: estes schemas descrevem artefatos do harness, nunca mutações.

| Schema | Artefato |
|--------|----------|
| [`console-event.schema.json`](./console-event.schema.json) | Evento tipado no feed (bus/audit/live) |
| [`summary.schema.json`](./summary.schema.json) | `GET /api/summary` |
| [`gate-run.schema.json`](./gate-run.schema.json) | Última execução de gates |
| [`domain-health.schema.json`](./domain-health.schema.json) | Território / domínio |

## Prefixos de evento (`type`)

| Prefixo | Cor (UI) | Origem típica |
|---------|----------|---------------|
| `consultation.*` | ciano | signal-bus consult/attract |
| `gates.passed` / `gates.failed` | verde / vermelho | record-agent-event / CI |
| `change.*` | âmbar | sessões / file edits |
| `evolution.*` | `#B49BE0` | evolve / propose |
| `session.*` | âmbar | `.cursor/arah-live` |

Wire format alinhado a `arah-harness/signal` v0.2 (`v` aditivo). Ver [SIGNAL_COMPATIBILITY.md](../../docs/SIGNAL_COMPATIBILITY.md).

## Paths observados

- Quente: `.arah/local/bus/`, `.arah/local/audit/`
- Legado: `.arah/bus/signals.jsonl`, `.arah/audit/events.jsonl`
- Meta: `docs/_meta/agent-graph.generated.json`, `discovery.proposed.yaml`, `evolution.proposed.yaml`
- Config: `arah.config.yaml`, `.arah-version`

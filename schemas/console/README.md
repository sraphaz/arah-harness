# schemas/console/

Contratos de dados lidos pelo ARAH Live Console (épico C).

**Status:** placeholder — implementação completa em [C-01](../../docs/backlog/C-live-console.md).

Pré-requisito satisfeito: tipos de sinal versionados (`arah-harness/signal` v0.2.0 + `docs/SIGNAL_COMPATIBILITY.md`).

Quando C-01 for aberto, versionar aqui:

- eventos do bus (`.arah/local/bus/**`, legado `.arah/bus/*.jsonl`)
- ledger (`.arah/local/audit/**`)
- `docs/_meta/{domains.yaml,graph.json,discovery.proposed.yaml}`
- `arah.config.yaml` (subset consumido pelo summary)

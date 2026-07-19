# ARAH Live Service

Serviço local **read-only** para o ARAH Live Console (épico C).

## Quick start

```bash
cd live
go mod tidy
go run ./cmd/arah-live -repo .. -addr 127.0.0.1:8787
```

Opcional:

```bash
go run ./cmd/arah-live -repo .. -reindex
export ARAH_GITHUB_REPO=sraphaz/arah-harness
export GITHUB_TOKEN=ghp_...   # read-only
```

## API

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/api/health` | liveness |
| GET | `/api/summary` | KPIs + drift + kernel |
| GET | `/api/feed?filter=` | eventos (`consulta\|gates\|mudancas\|evolucao`) |
| GET | `/api/gates` | última execução agregada |
| GET | `/api/domains` | territórios |
| GET | `/api/queue` | PRs abertos (GitHub) |
| GET | `/api/proposals` | `evolution.proposed.yaml` |
| GET | `/api/graph` | agent-graph JSON |
| WS | `/events` | ticks tipados |

**Sem endpoints de escrita** — POST/PUT/PATCH/DELETE → 405.

## Storage

- Índice descartável: `.arah/local/index/live-index.sqlite`
- Fontes: `.arah/local/{bus,audit}/**`, legado jsonl, `.cursor/arah-live/`

Schemas: [`../schemas/console/`](../schemas/console/).

## Console web

No site (`website/`), defina:

```bash
NEXT_PUBLIC_LIVE_API=http://127.0.0.1:8787 pnpm dev
```

Sem a variável, o console continua no mock embutido.

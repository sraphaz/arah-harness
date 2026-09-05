# Context budget — before/after

Medição determinística (`chars/4`) do tamanho de contexto enviado ao agente.

| Modo | Tokens ≈ |
|------|----------|
| **before** (AGENTS.md + EXECUTION_CONTROL + contrato + briefing) | ~3020 |
| **after** minimal | ~26 |
| **after** standard | ~44 |
| **after** full | ~301 |

Economia **standard vs before ≈ 98.5%**.

Gerar de novo:

```bash
go run ./cmd/arah economy context-compare --json --target .
```

Ver `docs/ECONOMY.md` · MCP `arah_get_task_context`.

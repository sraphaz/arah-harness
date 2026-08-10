# MCP Tool Contract — Arah Harness

> Primary agent surface for arah-core (0.5 foundation), inspired by
> [kern ADR-0007](https://github.com/rafaelnicolett/kern/blob/main/docs/adr/0007-mcp-as-the-primary-contract.md)
> and [kern mcp-tool-contract](https://github.com/rafaelnicolett/kern/blob/main/docs/architecture/mcp-tool-contract.md).
>
> CLI and MCP call the **same** `TaskService` use cases. Transport: MCP JSON-RPC
> over **stdio** (`arah mcp serve`).

## Error / result shape

Unlike kern (plain error strings), every tool returns a JSON **envelope** inside
the MCP text content:

```json
{
  "ok": false,
  "code": "EXECUTION.COMPLETION_EVIDENCE_REQUIRED",
  "message": "A tarefa não pode ser concluída sem evidência.",
  "trace_id": "…",
  "details": { "task_id": "task-…" },
  "remediation": ["Informe arquivos alterados", "Registre os testes executados"]
}
```

`isError: true` on the MCP result when `ok` is false.

---

## `arah_get_capabilities`

No required input. Returns runtime version, surfaces, and command list.

## `arah_get_task`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |

## `arah_create_task`

| Field | Type | Required |
|-------|------|----------|
| `objective` | string | yes |
| `area` | string | no (default `backend`) |
| `work_class` | string | no (`trivial\|standard\|architectural\|release`) |
| `intent_type` | string | no (`execution` default) |

Creates the contract, routes choreography, and starts `executing`.

## `arah_complete_task`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |
| `evidence` | string[] | yes (concrete) |

## `arah_block_task`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |
| `reason` | string | yes |

## `arah_get_timeline`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |

Returns append-only `task_events` from SQLite (`task.created`, `task.started`, …).

## `arah_get_evidence_graph`

No required input. Deterministic graph from specs (`covers`), tasks
(`assigned_to`, `evidenced_by`, `implements`, …). No LLM.

---

## Non-goals (v0 of this contract)

- Generic shell execution
- RAG / ontology / vector search (kern product surface — not absorbed)
- REST/OpenAPI as primary contract

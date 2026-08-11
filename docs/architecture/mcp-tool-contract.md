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
| `dry_run` | boolean | no — plan without persist / emit |

Creates the contract, routes choreography, and starts `executing` in the returned plan.
With `dry_run=true`, `TaskService.Create` still returns that executing plan but skips
persistence and event emission (path `dry-run`). Non-dry-run behaviour is unchanged.
Response includes textual `diff` and `idempotent` (always false for create).

## `arah_complete_task`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |
| `evidence` | string[] | yes (concrete) |
| `dry_run` | boolean | no |

Returns `diff` (state/evidence). Re-complete with the same evidence on an already-done
task is **idempotent** (`idempotent: true`, empty diff, no Save/emit).

## `arah_block_task`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |
| `reason` | string | yes |
| `dry_run` | boolean | no |

Re-block with the same reason is **idempotent** (no Save/emit).

## `arah_get_timeline`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |

Returns append-only `task_events` from SQLite (`task.created`, `task.started`, …)
including correlation fields `run_id`, `correlation_id`, `agent_id`, `session_id`.

## `arah_get_task_context`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |
| `budget` | string | no (`minimal` \| `standard` \| `full`, default `standard`) |

Progressive disclosure of task state with `estimated_tokens` (chars/4 proxy).
Prefer this over dumping AGENTS.md + full contract every turn.

## `arah_explain_route`

| Field | Type | Required |
|-------|------|----------|
| `area` | string | no |
| `preferred` | string | no |

Returns choreography decision (primary_executor, consultants, allowed_paths).

## `arah_get_evidence`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |

Evidence Graph slice for one task (`evidence explain`).

## `arah_submit_consultation`

| Field | Type | Required |
|-------|------|----------|
| `task_id` | string | yes |
| `consultant_id` | string | yes |
| `summary` | string | yes |
| `recommendations` | string[] | no |
| `blockers` | string[] | no |

Writes structured YAML under `.arah/local/execution/<task-id>/consultations/`.
Increments consultation counters; respects work-class limits.

## `arah_get_evidence_graph`

No required input. Deterministic graph from specs (`covers`), tasks
(`assigned_to`, `evidenced_by`, `implements`, …), and runtime events
(`validated_by`). No LLM.

---

## Non-goals (v0 of this contract)

- Generic shell execution
- RAG / ontology / vector search (kern product surface — not absorbed)
- REST/OpenAPI as primary contract
- CodeAct / unrestricted REPL (NOOA research surface — principles only)
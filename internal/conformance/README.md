# Conformance suite (H-20)

Proves Runtime Cohesion 0.5 invariants against fixture repositories.

## Fixtures

| Fixture | Purpose |
|---------|---------|
| `valid-minimal/` | Healthy choreography + config |
| `invalid-config/` | Missing `primary_executor` |
| `monorepo/` | Multi-path routing skeleton |
| `task-blocked/` | Seeded blocked-task flow |
| `empty/` | Bare directory |
| `drift-kernel/` | Built in-test from valid-minimal + stale kernel |

## Proofs

- Dry-run create/complete does not persist
- Stable `EXECUTION.*` error codes
- CLI ↔ MCP parity for create/complete/block
- Transition: `executing → routed` forbidden
- BRIEFING.md written on create
- Context budget MCP tool
- Kernel verify dogfood on module root
- Event correlation fields (`run_id`, `correlation_id`, `agent_id`)

```bash
go test ./internal/conformance/ -count=1
```

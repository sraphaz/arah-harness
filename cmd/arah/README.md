# ARAH portable CLI (Go) — arah-core

ADR-001 + **0.5 Runtime Cohesion** foundation. CLI and MCP share the same
`internal/core` use cases (engineering patterns inspired by
[rafaelnicolett/kern](https://github.com/rafaelnicolett/kern) — hexagonal core,
MCP as agent contract, local per-repo state — **not** its RAG/ontology product).

```bash
# from repo root
go test ./...
go build -o arah ./cmd/arah

./arah doctor -target .
./arah sync-check -target .
./arah version --json

./arah task create -objective "…" -area backend --json
./arah task status -task-id task-… --json
./arah task complete -task-id task-… -evidence "path updated; tests passed" --json
./arah task block -task-id task-… -reason "missing credential X" --json

./arah mcp serve -target .
```

| Exit | Meaning |
|------|---------|
| 0 | OK |
| 1 | Error (see `--json` `code`) |
| 2 | Drift (sync-check) |
| 4 | Doctor unhealthy |
| 10 | Usage |

PowerShell (`cli/arah.ps1`) remains the full organism surface during strangler
migration; writes for **Execution Control tasks** are available in Go now.

# ARAH portable CLI (Go) — arah-core

ADR-001 + **0.5 Runtime Cohesion**. CLI and MCP share `internal/core` use cases
(engineering patterns from [rafaelnicolett/kern](https://github.com/rafaelnicolett/kern)).

```bash
go test ./...
go build -o arah ./cmd/arah

./arah doctor -target .
./arah task create -objective "…" -area backend --json
./arah task timeline -task-id task-… --json
./arah evidence graph --json
./arah mcp serve -target .
./arah kernel sync -target .
./arah kernel verify -target .
```

Hot state: `.arah/local/runtime.db` (SQLite WAL) with YAML mirror under
`.arah/local/execution/` for PowerShell compatibility.

Kernel package (H-15): edit root sources, then `kernel sync`. Do not hand-edit `kernel/`.

| Exit | Meaning |
|------|---------|
| 0 | OK |
| 1 | Error (see `--json` `code`) |
| 2 | Drift (sync-check) |
| 4 | Doctor unhealthy |
| 10 | Usage |

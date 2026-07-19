# ARAH portable CLI (Go) — phase 1

H-07 / ADR-001. Read-only commands with exit-code parity; PowerShell remains canonical for write/organism flows.

```bash
cd cmd/arah
go build -o arah .
./arah doctor -target ../..
./arah sync-check -target ../..
./arah version
```

| Exit | Meaning |
|------|---------|
| 0 | OK |
| 1 | Error |
| 2 | Drift (sync-check) |
| 4 | Doctor unhealthy |
| 10 | Usage |

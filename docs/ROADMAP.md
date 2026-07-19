# Roadmap — arah-harness

## Now · v0.3.1 Control Plane foundations

- Estado quente × evidência fria (`.arah/local/` + `docs/_meta/runs/`)
- Arquivo-por-evento + `arah compact` / `migrate-state`
- Scrubbing de secrets na evidência; hooks pre-commit
- `capabilities.yaml`; ADR-001 (CLI Go); install `-Minimal`
- Backlog visual: site (W) + Live Console (C) em `docs/backlog/`

## Next

- `propose_and_draft_pr` — Apply discovery/evolution abre PR draft opcional  
- Doctor com diff de managed-blocks (profile install)  
- MCP: agent-graph + leitura do bus de sinais  
- Heurísticas de discovery mais ricas (PowerShell, monorepos profundos, README NLP leve)  
- **H-07** CLI binária portátil (Go) — doctor / sync-check / export-graph  
- C-01 schemas do console (contrato de artefatos)

## Later · v1.0

- Produto Arah consome harness como dependência versionada (prova real)  
- **W** Site de produto + portal docs MDX  
- **C** ARAH Live Console MVP + **H-08** `arahd`  
- Releases semver com CHANGELOG contínuo e notas de migração automáticas  
- Profiles enterprise com retention contractual + dashboards de evolução  

## Norte

**ARAH Harness · TechOrganism** como control plane de repositórios sérios: autonomia crescente, ledger intacto, humano no merge.

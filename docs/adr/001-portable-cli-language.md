# ADR-001 — Linguagem da CLI portátil

- **Status:** Accepted
- **Date:** 2026-07-19
- **Deciders:** maintainers (proposta agente; merge humano)

## Context

A CLI atual é PowerShell (`cli/arah.ps1` + scripts). Isso acopla a distribuição a `pwsh`, com startup lento e fricção em macOS/Linux. O Live Service (épico C) e um daemon opcional (`arahd`, H-08) precisam da mesma leitura de artefatos. Critérios: startup, single-binary, reuso no Live Service, manutenção do contrato “arquivos são a API”.

## Options

| Opção | Startup | Single-binary | Reuso Live Service | Custo de migração |
|-------|---------|---------------|--------------------|-------------------|
| **Go** | Excelente | Nativo | Mesmo módulo reader | Médio |
| Rust | Excelente | Nativo | Crates separados | Alto |
| Node SEA | Bom | Possível | Alto reuso TS | Médio (runtime) |

## Decision

Adotar **Go** para a CLI portátil fase 1 (H-07) e como base do Live Service reader (C-02/C-03), mantendo PowerShell como implementação de referência durante a transição.

## Consequences

- Paridade de exit codes 0/1/2/3/4/10 (ver CLI_SURFACE.md).
- Fase 1: `doctor`, `sync-check`, `export-graph` (leitura/validação).
- Escrita (signal-bus, record, compact) permanece em PowerShell até H-07/H-08.
- Contratos YAML/JSON não mudam — a CLI é substituível.

# AGENTS.md — Manual de operação por agentes (ARAH)

**Projeto**: {{project_name}}
**Harness**: ARAH (profile consulting)

Fonte de verdade para agentes (Cursor, CI). Procedimentos em `.skills/`; regras em `.cursor/rules/`; operação completa em `docs/ops/AGENT_OPERATION.md`.

## Princípios

1. **Humano comanda, agente executa** — merge sempre humano.
2. **Tudo via Pull Request** — sem commit direto em `main`.
3. **Escopo mínimo** — cada agente só toca paths permitidos.
4. **Doc como código** — documentação atualizada no mesmo PR.
5. **Spec-before-work** — mudanças cobertas por spec em `docs/specs/`.
6. **Contexto sob demanda** — comunicação entre agentes é passiva (arquivo + CI).

## Fluxo

```
Intenção (humano) → Orquestrador → Agente + skills → PR → CI + PR Steward → ready-for-merge → Merge (humano)
```

## Catálogo

Operacionais: [`.agents/README.md`](.agents/README.md).

Consultivos de domínio: `.agents/domain/` (advisors do harness + domínios de negócio via `arah.config.yaml`).

Coreografia: [`.agents/choreography.yaml`](.agents/choreography.yaml).

## Skills

```powershell
./scripts/agents/invoke-skill.ps1 -Skill <nome>
./scripts/agents/validate-manifests.ps1
```

## Referências

- [ARAH Harness](https://github.com/sraphaz/arah-harness) — kernel e método
- `docs/specs/` — specs SDD
- `docs/governance/DEFINITION_OF_DONE.md` — Definition of Done

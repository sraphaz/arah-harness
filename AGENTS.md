# AGENTS.md — Manual de operação por agentes (ARAH)

**Projeto**: arah-harness
**Harness**: ARAH 0.2.2

Fonte de verdade para agentes (Cursor, CI). Procedimentos em `.skills/`; regras em `.cursor/rules/`; operação completa em `docs/ops/AGENT_OPERATION.md` (crie se necessário).

## Princípios

1. **Humano comanda, agente executa** — merge sempre humano.
2. **Tudo via Pull Request** — sem commit direto em `main`.
3. **Escopo mínimo** — cada agente só toca paths permitidos (`.agents/*.agent.yaml`).
4. **Doc como código** — documentação atualizada no mesmo PR.
5. **Spec-before-code** — fases S0+ exigem spec em `docs/specs/` e `Spec-Id:` no PR.
6. **Contexto sob demanda** — comunicação entre agentes é passiva (arquivo + CI).

## Fluxo

```
Intenção (humano) → Orquestrador → Agente + skills → PR → CI + PR Steward → ready-for-merge → Merge (humano)
```

## Catálogo

Operacionais: ver [`.agents/README.md`](.agents/README.md).

Consultivos de domínio: `.agents/domain/` (gerados via `arah.config.yaml`).

Coreografia: [`.agents/choreography.yaml`](.agents/choreography.yaml).

## Skills

```powershell
./scripts/agents/invoke-skill.ps1 -Skill <nome> [-Area backend|frontend]
./scripts/agents/validate-manifests.ps1
```

## Configuração

Edite [`arah.config.yaml`](arah.config.yaml) para comandos de teste, domínios e especialistas.

## Harness profiles (Onda 2)

Profiles instaláveis em `harness/profiles/` — ver `harness/profiles/consulting.yaml` e [ARAH_HARNESS_EXTRACTION_PLAN.md](docs/_meta/ARAH_HARNESS_EXTRACTION_PLAN.md).

```powershell
./harness/scripts/doctor-harness.ps1 -Target <repo-path>
```

Schemas: `docs/schemas/`

## Referências

- [ARAH Harness](https://github.com/sraphaz/arah-harness) — kernel e método
- `docs/specs/` — specs SDD
- `docs/governance/DEFINITION_OF_DONE.md` — crie conforme necessidade


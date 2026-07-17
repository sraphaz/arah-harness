# AGENTS.md — Manual de operação por agentes (ARAH)

**Projeto**: arah-harness
**Harness**: ARAH 0.3.0

Fonte de verdade para agentes (Cursor, CI). Procedimentos em `.skills/`; regras em `.cursor/rules/`; operação completa em `docs/ops/AGENT_OPERATION.md` (crie se necessário).

## Princípios

1. **Humano comanda, agente executa** — merge sempre humano.
2. **Tudo via Pull Request** — sem commit direto em `main`.
3. **Escopo mínimo** — cada agente só toca paths permitidos (`.agents/*.agent.yaml`).
4. **Doc como código** — documentação atualizada no mesmo PR.
5. **Spec-before-code** — fases S0+ exigem spec em `docs/specs/` e `Spec-Id:` no PR.
6. **Contexto sob demanda** — comunicação passiva (arquivo + CI) e sinais tipados no bus.
7. **Agentes propõem; humanos aplicam** — discovery/evolve nunca spawnam células em silêncio.

## Fluxo

```
Intenção (humano) → Orquestrador → Agente + skills → PR → CI + PR Steward → ready-for-merge → Merge (humano)
```

Biocomponente (homeostase):

```
discover → organism bootstrap → sinais → evolve → regenerate → PR
```

## Catálogo

Operacionais: ver [`.agents/README.md`](.agents/README.md).

Consultivos de domínio: `.agents/domain/` (gerados via `arah.config.yaml`).

Coreografia: [`.agents/choreography.yaml`](.agents/choreography.yaml).

Organismo: `docs/_meta/organism.manifest.yaml` — ver [docs/BIOCOMPONENT.md](docs/BIOCOMPONENT.md).

## Skills

```powershell
./scripts/agents/invoke-skill.ps1 -Skill <nome> [-Area backend|frontend]
./scripts/agents/validate-manifests.ps1
```

Skills biocomponente: `discover-repo`, `evolve-harness`, `regenerate-harness`.

## Configuração

Edite [`arah.config.yaml`](arah.config.yaml) para comandos de teste, domínios e especialistas.

```powershell
powershell -File ./cli/arah.ps1 discover
powershell -File ./cli/arah.ps1 organism bootstrap
powershell -File ./cli/arah.ps1 evolve
powershell -File ./cli/arah.ps1 regenerate
```

## Harness profiles (Onda 2)

Profiles instaláveis em `harness/profiles/` — ver `harness/profiles/consulting.yaml` e [ARAH_HARNESS_EXTRACTION_PLAN.md](docs/_meta/ARAH_HARNESS_EXTRACTION_PLAN.md).

```powershell
./harness/scripts/doctor-harness.ps1 -Target <repo-path>
```

Schemas: `docs/schemas/` · `schemas/arah-harness/`

## Referências

- [ARAH Harness](https://github.com/sraphaz/arah-harness) — kernel e método
- `docs/BIOCOMPONENT.md` — dimensão autônoma
- `docs/specs/` — specs SDD
- `docs/governance/DEFINITION_OF_DONE.md` — crie conforme necessidade


# AGENTS.md — Operação por agentes (ARAH)

**Projeto:** arah-harness  
**Harness:** ARAH **0.3.0** · Biocomponente ativo

Fonte de verdade para agentes (Cursor, CI). Procedimentos em `.skills/`; operação profunda em `docs/ops/` (crie se necessário).

---

## Princípios

1. **Humano comanda, agente executa** — merge sempre humano  
2. **Tudo via Pull Request** — sem commit direto em `main`  
3. **Escopo mínimo** — cada agente só toca paths do seu manifest  
4. **Doc como código** — documentação no mesmo PR  
5. **Spec-before-code** — fases S0+ com `Spec-Id:` no PR  
6. **Contexto sob demanda** — arquivo + CI + sinais tipados  
7. **Agentes propõem; humanos aplicam** — sem spawn silencioso  

---

## Fluxos

### Entrega

```text
Intenção → Orquestrador → Célula + skills → PR → CI + PR Steward → ready-for-merge → Merge
```

### Organismo (homeostase)

```text
discover → organism bootstrap → sinais → evolve → regenerate → PR
```

---

## Catálogo

| Tipo | Onde |
|------|------|
| Operacionais | [`.agents/README.md`](.agents/README.md) |
| Domínio | `.agents/domain/` (via `arah.config.yaml` + `domain sync`) |
| Specialists | `.agents/specialists/` |
| Coreografia | [`.agents/choreography.yaml`](.agents/choreography.yaml) |
| Organismo | [`docs/_meta/organism.manifest.yaml`](docs/_meta/organism.manifest.yaml) |

Dimensão viva: **[docs/BIOCOMPONENT.md](docs/BIOCOMPONENT.md)**

---

## Skills

```powershell
./scripts/agents/invoke-skill.ps1 -Skill <nome> [-Area backend|frontend]
./scripts/agents/validate-manifests.ps1
```

Biocomponente: `discover-repo` · `evolve-harness` · `regenerate-harness`

---

## CLI rápida (este repo)

```powershell
powershell -File ./cli/arah.ps1 discover
powershell -File ./cli/arah.ps1 organism bootstrap
powershell -File ./cli/arah.ps1 organism status
powershell -File ./cli/arah.ps1 evolve
powershell -File ./cli/arah.ps1 regenerate
powershell -File ./scripts/self-test.ps1
```

Config: [`arah.config.yaml`](arah.config.yaml)

---

## Profiles

`harness/profiles/` — ver `consulting.yaml` e [HARNESS_PROFILES.md](docs/HARNESS_PROFILES.md).

```powershell
./harness/scripts/doctor-harness.ps1 -Target <repo-path>
```

Schemas: `schemas/arah-harness/` · `docs/schemas/`

---

## Referências

| Doc | Uso |
|-----|-----|
| [docs/BIOCOMPONENT.md](docs/BIOCOMPONENT.md) | Organismo, sinais, evolução |
| [docs/METHOD.md](docs/METHOD.md) | Método completo |
| [docs/GOVERNANCE.md](docs/GOVERNANCE.md) | Autonomia e gates |
| [docs/specs/](docs/specs/) | Specs SDD |
| [docs/governance/DEFINITION_OF_DONE.md](docs/governance/DEFINITION_OF_DONE.md) | DoD |

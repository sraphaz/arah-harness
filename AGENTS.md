# AGENTS.md — Operação por agentes (ARAH)

**Projeto:** arah-harness  
**Harness:** ARAH **0.4.0** · TechOrganism + Execution Control

Fonte de verdade para agentes (Cursor, CI). Procedimentos em `.skills/`; operação profunda em `docs/ops/` (crie se necessário).

---

## Princípios

1. **Humano comanda, agente executa** — merge sempre humano  
2. **Tudo via Pull Request** — sem commit direto em `main`  
3. **Escopo mínimo** — cada agente só toca paths do seu manifest  
4. **Doc como código** — documentação no mesmo PR  
5. **Spec-before-code** — proporcional à classe de trabalho (`trivial` dispensa ritual completo)  
6. **Contexto sob demanda** — arquivo + CI + sinais tipados  
7. **Agentes propõem; humanos aplicam** — sem spawn silencioso  
8. **Um executor primário** — vários participantes, uma responsabilidade de entrega  

---

## Execution Control Protocol

Toda tarefa executável deve possuir exatamente um **primary_executor**.

O orquestrador encerra seu papel após:
1. classificar a intenção;
2. resolver a coreografia;
3. selecionar o executor;
4. criar o contrato de execução.

Depois do estado `executing`, a tarefa não retorna para `routed`.

Consultores:
- retornam parecer estruturado ao executor;
- não conversam entre si;
- não redefinem a tarefa;
- não executam alterações;
- não bloqueiam a entrega sem evidência crítica.

O executor:
- altera os arquivos necessários;
- executa verificações;
- registra evidências;
- encerra a tarefa como `done` ou `blocked`.

Uma análise não equivale a uma entrega quando a intenção é executável.

```powershell
arah task create -Objective "…" -Area backend -Class standard
arah task complete -TaskId task-… -Evidence "path updated; tests passed"
```

Detalhes: [`docs/EXECUTION_CONTROL.md`](docs/EXECUTION_CONTROL.md)

---

## Fluxos

### Entrega

```text
Intenção → Orquestrador → contrato + executor → consultas limitadas → alteração → verificação → done|blocked → PR → CI → Merge
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

Dimensão viva: **[docs/TECHORGANISM.md](docs/TECHORGANISM.md)**

---

## Skills

```powershell
./scripts/agents/invoke-skill.ps1 -Skill <nome> [-Area backend|frontend]
./scripts/agents/validate-manifests.ps1
```

TechOrganism: `discover-repo` · `evolve-harness` · `regenerate-harness`

---

## CLI rápida (este repo)

```powershell
powershell -File ./cli/arah.ps1 discover
powershell -File ./cli/arah.ps1 organism bootstrap
powershell -File ./cli/arah.ps1 organism status
powershell -File ./cli/arah.ps1 evolve
powershell -File ./cli/arah.ps1 regenerate
powershell -File ./cli/arah.ps1 task create -Objective "…" -Area backend
powershell -File ./scripts/self-test.ps1
powershell -File ./scripts/harness/test-execution-control.ps1
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
| [docs/TECHORGANISM.md](docs/TECHORGANISM.md) | Organismo, sinais, evolução |
| [docs/METHOD.md](docs/METHOD.md) | Método completo |
| [docs/GOVERNANCE.md](docs/GOVERNANCE.md) | Autonomia e gates |
| [docs/EXECUTION_CONTROL.md](docs/EXECUTION_CONTROL.md) | Terminalidade e executor primário |
| [docs/specs/](docs/specs/) | Specs SDD |
| [docs/governance/DEFINITION_OF_DONE.md](docs/governance/DEFINITION_OF_DONE.md) | DoD |

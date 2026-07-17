# AGENTS.md — Operação por agentes (ARAH)

**Projeto:** {{PROJECT_NAME}}  
**Harness:** ARAH {{HARNESS_VERSION}}

Fonte de verdade para agentes (Cursor, CI). Procedimentos em `.skills/`; operação profunda em `docs/ops/AGENT_OPERATION.md` (crie se necessário).

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
| Operacionais | `.agents/README.md` |
| Domínio | `.agents/domain/` (via `arah.config.yaml` + `domain sync`) |
| Specialists | `.agents/specialists/` |
| Coreografia | `.agents/choreography.yaml` (+ overlays locais) |
| Organismo | `docs/_meta/organism.manifest.yaml` |

---

## Skills

```powershell
./scripts/agents/invoke-skill.ps1 -Skill <nome> [-Area backend|frontend]
./scripts/agents/validate-manifests.ps1
```

Biocomponente: `discover-repo` · `evolve-harness` · `regenerate-harness`

---

## Configuração

Edite `arah.config.yaml` (testes, domínios, specialists).

```powershell
# Com o clone do arah-harness disponível:
arah discover
arah organism bootstrap
arah evolve
arah regenerate -UpdateKernel
```

Documentação do biocomponente: no repositório upstream `docs/BIOCOMPONENT.md`.

---

## Referências

- Upstream: https://github.com/sraphaz/arah-harness  
- Specs: `docs/specs/`  
- Definition of Done: `docs/governance/DEFINITION_OF_DONE.md`  

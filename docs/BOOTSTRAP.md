# Bootstrap — checklist pós-install

Checklist premium após `arah install` (ou `init`). Objetivo: organismo saudável, domínio alinhado, CI verde.

---

## 1. Identidade do projeto

Em `arah.config.yaml`:

- [ ] `project.name` e `project.stack`
- [ ] `harness.version` / `release` apontando para **v0.3.0+**
- [ ] `tests.*` com comandos reais da suíte

---

## 2. Descobrir e declarar domínio

**Opção A — TechOrganism (recomendada)**

```powershell
arah discover
# revise docs/_meta/discovery.proposed.yaml
arah discover -Apply          # se concordar com as candidaturas
arah domain sync
arah organism bootstrap
arah evolve
```

**Opção B — manual**

- [ ] Preencher `domains[]` e `specialists[]` em `arah.config.yaml`
- [ ] `arah domain sync`
- [ ] `arah organism bootstrap`

- [ ] `docs/_meta/organism.manifest.yaml` existe
- [ ] Tecidos (`delivery`, `governance`, `craft`, `domain-sense`) fazem sentido

---

## 3. Validar instalação

```powershell
./scripts/agents/validate-manifests.ps1
arah export-graph
arah doctor
```

- [ ] Doctor OK  
- [ ] Agent Graph gerado em `docs/_meta/`  
- [ ] Sem skills órfãs críticas  

---

## 4. Overlay e paths reais

- [ ] Overlay `.agents/choreography.<projeto>.yaml` (monorepo)
- [ ] Paths dos domínios batem com a árvore real
- [ ] Separação clara: agentes SDLC vs agentes runtime do produto (se houver)

---

## 5. Documentação mínima

- [ ] `docs/governance/DEFINITION_OF_DONE.md`
- [ ] `docs/ops/AGENT_OPERATION.md` (detalhes além do `AGENTS.md`)
- [ ] `docs/specs/README.md` (+ registry se profile consulting)
- [ ] Link para [TECHORGANISM.md](TECHORGANISM.md) no README do produto (opcional)

---

## 6. Cursor / IDE

- [ ] `.cursor/hooks.json` ativo
- [ ] Rules escopadas em `.cursor/rules/` conforme stack
- [ ] Extensão [ARAH Live](LIVE_SESSION.md) (opcional, telemetria)

---

## 7. CI e branch protection

- [ ] Workflow `agents-validate.yml` verde
- [ ] Branch protection em `main`
- [ ] `ARAH_HARNESS_PATH` no CI se usar `sync-check`

---

## 8. Greenfield vs brownfield

| Tipo | Ação |
|------|------|
| Greenfield | `install` → regenerate → primeiro Spec → primeira fase |
| Brownfield | `install` sem `-Force` → mesclar AGENTS → discover → PR só de propostas |

---

## 9. Ritmo de homeostase

| Frequência | Ação |
|------------|------|
| A cada release do harness | `regenerate -UpdateKernel` |
| Após mudança grande de stack/domínio | `discover` + `organism bootstrap -Force` |
| Periodicamente / pós-fase | `evolve` → revisar propostas → PR |

```powershell
arah regenerate -UpdateKernel -Force
arah sync-check
```

Commit `.arah-version` após update.

---

## Referências

- [INSTALL.md](INSTALL.md) — instalação completa  
- [TECHORGANISM.md](TECHORGANISM.md) — dimensão viva  
- [METHOD.md](METHOD.md) — método  
- [GOVERNANCE.md](GOVERNANCE.md) — autonomia  

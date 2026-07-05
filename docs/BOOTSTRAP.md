# Bootstrap — checklist pós-install

Após `powershell -File path/to/arah-harness/cli/arah.ps1 install` (ou `init`):

## 1. Configurar projeto

Edite `arah.config.yaml`:

- `project.name` e `project.stack`
- `tests.backend`, `tests.frontend` com comandos reais
- `domains[]` com agentes consultivos de negócio

## 2. Gerar domínios

```powershell
powershell -File path/to/arah-harness/cli/arah.ps1 domain sync
powershell -File path/to/arah-harness/cli/arah.ps1 export-graph
```

## 3. Validar instalação

```powershell
powershell -File ./scripts/agents/validate-manifests.ps1
powershell -File path/to/arah-harness/cli/arah.ps1 doctor
```

## 4. Stack e CI

O workflow `.github/workflows/agents-validate.yml` vem no `init`. Confirme que roda:

- `validate-manifests.ps1`
- `validate-agent-graph.ps1` (via export ou CI)
- `validate-specs.ps1` (se usar SDD)

Configure branch protection em `main`.

## 5. Documentação mínima

Crie se não existir:

- `docs/governance/DEFINITION_OF_DONE.md`
- `docs/ops/AGENT_OPERATION.md` (detalhes além do AGENTS.md enxuto)
- `docs/specs/README.md`

## 6. Cursor

- Confirme `.cursor/hooks.json` ativo
- Adicione rules escopadas em `.cursor/rules/` conforme stack

## 7. Novo projeto vs repo existente

**Greenfield**: `install` → config → `domain sync` → primeiro spec → primeira fase.

**Brownfield**: `install` sem `-Force` (não sobrescreve `AGENTS.md`/`README` existentes);
mescle manualmente a seção ARAH; rode `doctor` após merge.

## 8. Manter kernel atualizado

```powershell
powershell -File path/to/arah-harness/cli/arah.ps1 update -Force
powershell -File path/to/arah-harness/cli/arah.ps1 sync-check
```

Commit `.arah-version` após update.

Veja também: [INSTALL.md](INSTALL.md)

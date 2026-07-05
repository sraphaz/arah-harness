# Bootstrap — checklist pós-init

Após `pwsh path/to/arah-harness/cli/arah.ps1 init`:

## 1. Configurar projeto

Edite `arah.config.yaml`:

- `project.name` e `project.stack`
- `tests.backend`, `tests.frontend` com comandos reais
- `domains[]` com agentes consultivos de negócio

## 2. Validar instalação

```powershell
pwsh ./scripts/agents/validate-manifests.ps1
pwsh path/to/arah-harness/cli/arah.ps1 doctor
```

## 6. Domínio de negócio

Edite `domains` em `arah.config.yaml`, depois:

```powershell
powershell -File path/to/arah-harness/cli/arah.ps1 domain sync
powershell -File path/to/arah-harness/cli/arah.ps1 export-graph
```

## 4. Stack e CI

- Adicione workflow `.github/workflows/agents-validate.yml` que rode:
  - `validate-manifests.ps1`
  - `arah sync-check` (drift do kernel)
  - `validate-specs.ps1` (se usar SDD)
- Configure branch protection em `main`.

## 5. Documentação mínima

Crie se não existir:

- `docs/governance/DEFINITION_OF_DONE.md`
- `docs/ops/AGENT_OPERATION.md` (detalhes além do AGENTS.md enxuto)
- `docs/specs/README.md`

## 6. Cursor

- Confirme `.cursor/hooks.json` ativo
- Adicione rules escopadas em `.cursor/rules/` conforme stack

## 7. Novo projeto vs repo existente

**Greenfield**: init → config → primeiro spec → primeira fase.

**Brownfield**: init com cuidado (não sobrescrever AGENTS.md/README sem backup);
use `-Force` só quando intencional; rode `doctor` após merge manual.

## 8. Manter kernel atualizado

```powershell
pwsh path/to/arah-harness/cli/arah.ps1 update -Force
pwsh path/to/arah-harness/cli/arah.ps1 sync-check
```

Commit `.arah-version` após update.

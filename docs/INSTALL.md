# Instalar ARAH Harness

Guia premium para adicionar o kernel ARAH a **qualquer repositório** — greenfield ou brownfield — e ativar a dimensão TechOrganism.

**Upstream:** [sraphaz/arah-harness](https://github.com/sraphaz/arah-harness) · **Release:** **v0.3.0**

---

## Em 3 comandos

```powershell
git clone https://github.com/sraphaz/arah-harness.git $env:USERPROFILE\arah-harness
cd C:\caminho\para\meu-projeto

powershell -ExecutionPolicy Bypass -File $env:USERPROFILE\arah-harness\cli\arah.ps1 install `
  -ProjectName "meu-projeto"
```

### Modo mínimo (manifests + gates)

Para adoção leve sem organismo/bus no primeiro dia:

```powershell
powershell -File $env:USERPROFILE\arah-harness\cli\arah.ps1 install `
  -ProjectName "meu-projeto" -Minimal
```

Isso anota `organism.enabled: false` em `arah.config.yaml`. **Upgrade path:**

```powershell
# remova o bloco organism: minimal (ou enabled: true) e:
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 regenerate -Target . -UpdateKernel
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 discover -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 organism bootstrap -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 hooks install -Target .
```

Ative o organismo (recomendado na v0.3):

```powershell
$env:ARAH_HARNESS_PATH = "$env:USERPROFILE\arah-harness"
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 regenerate -Target . -UpdateKernel
```

Revise `docs/_meta/*.proposed.yaml` → ajuste `arah.config.yaml` → commit → PR.

---

## Variantes de layout

### Clone irmão (padrão IAutos)

```powershell
cd C:\caminho\para
git clone https://github.com/sraphaz/arah-harness.git arah-harness
cd meu-projeto
powershell -ExecutionPolicy Bypass -File ..\arah-harness\cli\arah.ps1 install -ProjectName "meu-projeto"
```

### Path fixo

```powershell
$env:ARAH_HARNESS_PATH = "C:\caminho\para\arah-harness"
powershell -ExecutionPolicy Bypass -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 install -ProjectName "meu-projeto"
```

---

## O que o `install` coloca no repo

| Artefato | Descrição |
|----------|-----------|
| `.agents/` | 11 operacionais + domain advisors + coreografia |
| `.skills/` | Skills executáveis (incl. TechOrganism) |
| `scripts/agents/` · `scripts/harness/` | Orquestração, gates, discover/evolve/regenerate |
| `.cursor/hooks.json` | Domain review + live session |
| `arah.config.yaml` | Config do projeto (**edite**) |
| `AGENTS.md` | Manual operacional (só se não existir) |
| `.github/workflows/agents-validate.yml` | CI de manifests + graph |
| `.arah-version` | Pin da versão do harness |

**Brownfield:** sem `-Force`, arquivos existentes não são sobrescritos. Mescle a seção ARAH manualmente se já houver `AGENTS.md`.

---

## Configurar `arah.config.yaml`

```yaml
harness:
  source: sraphaz/arah-harness
  version: "0.3.0"
  repository: https://github.com/sraphaz/arah-harness
  release: v0.3.0

project:
  name: meu-projeto
  stack: fullstack

tests:
  backend: "dotnet test"
  frontend: "npm test"
  all: "npm test && dotnet test"

domains:
  - id: meu-dominio
    name: Meu Domínio
    description: Regras de negócio desta área
    paths:
      - src/meu-dominio/**
    enrich: |
      Contexto que agentes devem respeitar.
    validate: |
      Invariantes a verificar no PR.
```

### Deixar o TechOrganism propor

```powershell
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 discover -Target .
# revise docs/_meta/discovery.proposed.yaml
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 discover -Target . -Apply   # opcional
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 domain sync -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 organism bootstrap -Target .
```

Detalhe: [TECHORGANISM.md](TECHORGANISM.md).

---

## Overlay local (monorepos)

Crie `.agents/choreography.<projeto>.yaml` com regras de path do seu monorepo (`packages/**`, `apps/**`, …).  
Referência: [choreography.iautos.yaml](https://github.com/sraphaz/iautos/blob/main/.agents/choreography.iautos.yaml).

Overlays sobrevivem a `update` / `regenerate`.

---

## Validar

```powershell
./scripts/agents/validate-manifests.ps1
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 export-graph -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 doctor -Target .
```

---

## Primeiro commit no repo-alvo

```powershell
git add .agents .skills scripts .cursor arah.config.yaml .arah-version .github AGENTS.md docs
git commit -m "chore: bootstrap ARAH Harness v0.3.1"
```

---

## Manutenção e evolução

| Objetivo | Comando |
|----------|---------|
| Homeostase completa (recomendado) | `arah regenerate -UpdateKernel -Force` |
| Só reaplicar kernel | `arah update -Force` |
| Detectar drift | `arah sync-check` |
| Self-learning | `arah evolve` |
| Redefinir organismo | `arah organism bootstrap -Force` |
| Migrar estado legado → quente/frio | `arah migrate-state` |
| Compactar eventos pending | `arah compact` |
| Instalar pre-commit hooks | `arah hooks install` |

```powershell
# Receber v0.3+ em consumidor existente
git -C $env:ARAH_HARNESS_PATH pull
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 regenerate `
  -Target . -UpdateKernel -Force
```

Isso reaplica o kernel, rediscovery, rebootstrap, evolve, grafo e doctor — e deixa sugestões em `docs/_meta/` para o repositório evoluir via PR.

---

## CI no repo-alvo

O workflow `agents-validate.yml` vem no `init`. Para drift-check opcional:

| Variável | Exemplo |
|----------|---------|
| `ARAH_HARNESS_PATH` | `../arah-harness` (checkout paralelo no CI) |

---

## Próximos passos

1. Checklist: [BOOTSTRAP.md](BOOTSTRAP.md)  
2. Método: [METHOD.md](METHOD.md)  
3. TechOrganism: [TECHORGANISM.md](TECHORGANISM.md)  
4. Migração Arah legado: [MIGRATION_FROM_ARAH.md](MIGRATION_FROM_ARAH.md)  
5. Exemplo real: [sraphaz/iautos](https://github.com/sraphaz/iautos)  

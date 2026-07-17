# Instalar ARAH Harness em outro repositório

Guia copy-paste para adicionar o kernel ARAH a **qualquer repo** (greenfield ou brownfield).

Repositório upstream: **https://github.com/sraphaz/arah-harness** (release **v0.2.0**)

---

## Comando rápido (recomendado)

```powershell
# 1) Clone o harness (uma vez por máquina ou por pasta de projetos)
git clone https://github.com/sraphaz/arah-harness.git $env:USERPROFILE\arah-harness

# 2) Entre no repo que receberá o ARAH
cd C:\caminho\para\meu-projeto

# 3) Instale (init + doctor + instruções pós-install)
powershell -ExecutionPolicy Bypass -File $env:USERPROFILE\arah-harness\cli\arah.ps1 install -ProjectName "meu-projeto"
```

Alternativa com clone **irmão** do projeto (padrão usado no [IAutos](https://github.com/sraphaz/iautos)):

```powershell
cd C:\caminho\para
git clone https://github.com/sraphaz/arah-harness.git arah-harness
cd meu-projeto
powershell -ExecutionPolicy Bypass -File ..\arah-harness\cli\arah.ps1 install -ProjectName "meu-projeto"
```

Ou defina o caminho uma vez:

```powershell
$env:ARAH_HARNESS_PATH = "C:\caminho\para\arah-harness"
powershell -ExecutionPolicy Bypass -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 install -ProjectName "meu-projeto"
```

---

## Passo a passo completo

### 1. `install` (kernel + templates + CI)

```powershell
powershell -ExecutionPolicy Bypass -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 install -Target . -ProjectName "meu-projeto"
```

Instala no repo-alvo:

| Artefato | Descrição |
|----------|-----------|
| `.agents/` | 11 agentes operacionais + checklists |
| `.skills/` | 18 skills |
| `scripts/agents/` + `scripts/harness/` | Orquestração, gates, agent graph |
| `.cursor/hooks.json` | Domain review passivo |
| `arah.config.yaml` | Config do projeto (edite!) |
| `AGENTS.md` | Manual (só se não existir) |
| `.github/workflows/agents-validate.yml` | CI de manifests + graph |
| `.arah-version` | Pin da versão do harness |

**Brownfield** (já tem `AGENTS.md`): o `init` **não sobrescreve** arquivos existentes. Mescle manualmente a seção ARAH do template ou do [IAutos](https://github.com/sraphaz/iautos/blob/main/AGENTS.md).

### 2. Configurar `arah.config.yaml`

Edite testes e domínios de negócio:

```yaml
harness:
  source: sraphaz/arah-harness
  version: "0.2.0"
  repository: https://github.com/sraphaz/arah-harness
  release: v0.2.0

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

### 3. Gerar agentes de domínio (ou deixar o biocomponente propor)

```powershell
# Manual:
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 domain sync -Target .

# Ou discovery + homeostase (v0.3+):
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 discover -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 organism bootstrap -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 domain sync -Target .
```

Cria `.agents/domain/*.agent.yaml` e `.agents/choreography.domains.yaml`.
Propostas ficam em `docs/_meta/discovery.proposed.yaml` — ver [BIOCOMPONENT.md](BIOCOMPONENT.md).

### 4. Overlay local (opcional, recomendado)

Crie `.agents/choreography.<projeto>.yaml` com regras de paths do seu monorepo (ex.: `packages/**`, `apps/**`). Ver [IAutos](https://github.com/sraphaz/iautos/blob/main/.agents/choreography.iautos.yaml).

### 5. Validar e exportar grafo

```powershell
./scripts/agents/validate-manifests.ps1
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 export-graph -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 doctor -Target .
```

### 6. Commit no repo-alvo

```powershell
git add .agents .skills scripts .cursor arah.config.yaml .arah-version .github AGENTS.md docs
git commit -m "chore: bootstrap ARAH Harness v0.3.0"
```

---

## Manutenção

```powershell
# Homeostase completa (recomendado a partir de v0.3)
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 regenerate -Target . -UpdateKernel -Force

# Ou só reaplicar kernel (preserva arah.config.yaml e overlays)
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 update -Target . -Force

# Detectar drift vs upstream
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 sync-check -Target .

# Ciclo de autoaprendizado
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 evolve -Target .
```

---

## CI no repo-alvo

O workflow `agents-validate.yml` já vem no `init`. Para drift-check opcional, configure a variável de repositório GitHub:

| Variável | Valor exemplo |
|----------|----------------|
| `ARAH_HARNESS_PATH` | `../arah-harness` (se checkout em CI) |

---

## Referências

- [BOOTSTRAP.md](BOOTSTRAP.md) — checklist pós-init
- [METHOD.md](METHOD.md) — arquitetura do método
- [MIGRATION_FROM_ARAH.md](MIGRATION_FROM_ARAH.md) — migrar repo Arah existente
- Exemplo real: [sraphaz/iautos](https://github.com/sraphaz/iautos)

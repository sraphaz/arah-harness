# Contribuindo

Obrigado por considerar contribuir com o **ARAH Harness**.

## Princípios

- **Kernel genérico** — sem paths ou domínios de um produto específico no `kernel/`
- **Uma fonte** — edite `.agents` / `.skills` / `.cursor` / `scripts/{agents,harness}` na raiz; rode `go run ./cmd/arah kernel sync` (não edite `kernel/` à mão). CI falha em drift (`arah kernel verify`). Consumidores sem checkout: `arah kernel install`
- **Overlays locais** — customização via `arah.config.yaml` e `choreography.*.yaml` no projeto-alvo
- **Determinismo** — scripts PowerShell 5.1+, sem dependências npm para o core
- **Economia de tokens** — comunicação passiva entre agentes (arquivo + CI)

## Setup

```powershell
git clone https://github.com/sraphaz/arah-harness.git
cd arah-harness
./scripts/self-test.ps1
```

## Pull requests

1. Fork + branch descritiva (`feat/domain-sync`, `fix/config-parser`)
2. `./scripts/self-test.ps1` verde
3. Descreva impacto em projetos que usam `arah init` / `update`
4. Atualize `CHANGELOG.md` e `VERSION` se mudança de release

### Releases (autônomo)

Não é necessário `git tag` manual. Após o merge em `main` de um PR que bumpa `VERSION` + `CHANGELOG.md`, o workflow **Release** publica sozinho a tag `vX.Y.Z` e o GitHub Release (`scripts/agents/cut-release.ps1`). Ver [docs/UPDATE_NOTIFICATIONS.md](docs/UPDATE_NOTIFICATIONS.md).

## Reportar bugs

Use [Issues](https://github.com/sraphaz/arah-harness/issues) com:

- Versão (`VERSION` ou `.arah-version` no projeto)
- Comando executado
- Saída de `doctor` / `validate-manifests`

## Código de conduta

Seja respeitoso. Foco em clareza operacional e auditabilidade — alinhado ao método ARAH.

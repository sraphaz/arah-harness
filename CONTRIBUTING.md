# Contribuindo

Obrigado por considerar contribuir com o **ARAH Harness**.

## Princípios

- **Kernel genérico** — sem paths ou domínios de um produto específico no `kernel/`
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

## Reportar bugs

Use [Issues](https://github.com/sraphaz/arah-harness/issues) com:

- Versão (`VERSION` ou `.arah-version` no projeto)
- Comando executado
- Saída de `doctor` / `validate-manifests`

## Código de conduta

Seja respeitoso. Foco em clareza operacional e auditabilidade — alinhado ao método ARAH.

# AGENTS.md — ARAH Harness (este repositório)

**Repo**: `sraphaz/arah-harness` · **Versão**: ver [VERSION](VERSION)

Este repositório **distribui** o método ARAH; não é um app de produto. O kernel vive em `kernel/`; projetos-alvo recebem cópia via `cli/arah.ps1 init`.

## Escopo de edição

| Área | Pode alterar | Notas |
|------|--------------|-------|
| `kernel/` | Sim | Bump `VERSION` + `CHANGELOG.md` em releases |
| `cli/` | Sim | Comandos init/update/doctor/sync-check/domain/export-graph |
| `templates/` | Sim | arah.config.yaml, AGENTS.md.tpl, workflows |
| `docs/` | Sim | METHOD, mercado, bootstrap, migração |
| `scripts/self-test.ps1` | Sim | CI e validação local |

## Antes de PR neste repo

1. `./scripts/self-test.ps1` — verde
2. Diff em `kernel/` revisado (impacta todos os projetos derivados)
3. `CHANGELOG.md` atualizado se release
4. `VERSION` alinhado com `cli/init.ps1` e `templates/arah.config.yaml`

## Comandos

```powershell
./scripts/self-test.ps1
powershell -File cli/arah.ps1 help
```

## Referências

- [docs/METHOD.md](docs/METHOD.md)
- [README.md](README.md)

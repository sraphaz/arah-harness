# Checklist — Release

- [ ] `VERSION` bumpado e seção `## [X.Y.Z]` em `CHANGELOG.md`
- [ ] PR mergeado em `main` (humano) — a Action **Release** publica tag + GitHub Release
- [ ] Confirmar em https://github.com/sraphaz/arah-harness/releases
- [ ] Workflows CI/CD válidos (lint sintaxe YAML)
- [ ] Sem secrets em workflows ou compose
- [ ] Plano de rollback documentado para deploys críticos

Opcional local: `arah release cut -DryRun` antes do merge.

Inclui [_shared.conduct.md](_shared.conduct.md).

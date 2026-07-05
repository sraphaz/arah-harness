# Migração do repositório Arah para ARAH Harness

Guia para extrair o app `arah` para consumir o kernel `arah-harness` sem perder capacidades.

## Situação atual

O repo `CursorRepos/arah` contém **duas camadas misturadas**:

| Camada | Exemplos | Destino |
|--------|----------|---------|
| Kernel ARAH | orchestrator, qa, invoke-skill, validate-manifests, export-agent-graph | `arah-harness/kernel/` |
| Domínio Arah | territorio-membership, mercado-economia, paths `backend/Arah.*` | Permanece no repo `arah` |
| App | código .NET/Flutter, docker, CI de produto | Permanece no repo `arah` |

## Estratégia recomendada (incremental)

### Fase 1 — Paralelo (agora)

1. Novos projetos usam `arah init` do harness.
2. Repo `arah` continua operando como hoje.
3. Evoluções do kernel vão para `arah-harness`; backport manual quando relevante.

### Fase 2 — Pin + drift-check

1. Adicionar `.arah-version` no repo `arah` apontando para harness `0.2.0`.
2. CI: `arah sync-check` contra clone do harness (variável `ARAH_HARNESS_PATH`).
3. Documentar overrides locais em `arah.config.yaml` → seção `overrides:` (futuro).

### Fase 3 — Split de manifests

```
arah/
├── arah.config.yaml          # domínios Arah, comandos de teste reais
├── .agents/
│   ├── *.agent.yaml          # operacionais → removidos (vêm do kernel via update)
│   ├── domain/               # 11 agentes de negócio Arah (permanecem)
│   ├── specialists/          # dotnet, flutter, etc. (permanecem)
│   ├── choreography.yaml     # só regras específicas Arah (ou merge)
│   └── choreography.domains.yaml  # gerado por domain sync
└── scripts/agents/
    └── overrides/            # scripts específicos Arah (BFF journey, LikeC4, etc.)
```

### Fase 4 — Consumo como dependência

Opções (escolher uma):

- **Git submodule**: `arah-harness` em `tools/arah-harness/` + `arah update` no CI
- **Copier**: overlay do harness com `.copier-answers.harness.yml` (updates automáticos)
- **npm/pwsh package global**: `arah init` / `arah update` via PATH

## Mapeamento de domínios Arah → arah.config.yaml

Exemplo de entrada para `territorio-membership`:

```yaml
domains:
  - id: territorio-membership
    name: Território & Membership
    description: Território-primeiro, papéis visitante/morador/curador
    paths:
      - backend/**/Territories/**
      - backend/**/Memberships/**
      - backend/Arah.Domain/Territories/**
    enrich: |
      Sugere regras território-primeiro e casos de papel.
    validate: |
      Escopo por territoryId; visibilidade imposta no servidor.
    references:
      - docs/05_GLOSSARY.md
      - docs/12_DOMAIN_MODEL.md
```

Depois: `powershell -File path/to/arah-harness/cli/arah.ps1 domain sync -Target .`

## Scripts que permanecem no Arah (não no kernel)

- `register-bff-journey-check.ps1`
- `design-gate-check.ps1`
- `likec4-export` / diagramas
- `parallel-attempt.ps1`
- `agent-metrics.ps1`
- Discord agent runner
- Workflows deploy-staging/production

Estes viram **skills locais** ou entradas em `scripts/agents/overrides/` referenciadas por skills custom no repo.

## Comandos de teste reais (arah.config.yaml)

```yaml
tests:
  backend: dotnet test backend/Tests/Arah.Tests/Arah.Tests.csproj --configuration Release
  frontend: flutter test
  bff: dotnet test backend/Tests/Arah.Tests.Bff/Arah.Tests.Bff.csproj --configuration Release
```

Estender `invoke-skill.ps1` localmente se precisar de áreas extras (`bff`, `flutter`).

## Checklist de migração

- [ ] `arah init` em branch de experimento
- [ ] Copiar `domains` para `arah.config.yaml`
- [ ] `domain sync`
- [ ] Comparar `validate-manifests` + `export-agent-graph` vs estado atual
- [ ] Portar workflows CI (agents-validate já vem do template)
- [ ] Remover duplicatas do kernel após validação
- [ ] Documentar overrides em `docs/ops/AGENT_OPERATION.md`

## Riscos

- **Drift**: kernel atualizado quebra customizações locais → mitigar com `sync-check` no CI.
- **Paths hardcoded**: agentes operacionais do Arah usam paths `Arah.*` → ficam no domínio/choreography local.
- **Skills duplicadas**: `.skills/` local deve declarar só skills extras; kernel fornece o resto.

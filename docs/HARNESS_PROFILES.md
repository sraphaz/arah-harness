# HARNESS PROFILES — ARAH Harness

**Versão**: 1.0 · **Data**: 2026-07-06

Profiles declarativos em `harness/profiles/` para instalação via `install-harness.ps1`.

| ID | Herda | Uso |
|----|-------|-----|
| `minimal` | — | Base: AGENTS.md, specs, CI mínima |
| `consulting` | minimal | Spec-before-work, gates de handoff |
| `product` | minimal | Roadmap, releases, ADRs, qualidade |
| `enterprise` | product | LGPD, threat model, audit log |
| `open-source` | product | CONTRIBUTING, governança pública |

## Instalação

```powershell
# Do repo arah-harness
./harness/scripts/install-harness.ps1 -Target ../meu-repo -Profile consulting

# Via CLI
./cli/arah.ps1 init -Target ../meu-repo -ProjectName meu-projeto
./cli/arah.ps1 domain sync -Target ../meu-repo
```

## O que cada profile instala

### minimal

- `AGENTS.md` com blocos gerenciados
- `docs/specs/_template.spec.yaml`
- `.agents/` kernel (orquestrador, spec-steward, qa)
- `harness-profile.yaml` pinado

### consulting

- Gates: `proposal_before_implementation`, `handoff_acceptance`
- PR checks: spec-before-work, no-secrets
- Domínio: craft + architecture consultivos em PRs

### product

- ADRs (`docs/adr/`), CHANGELOG, ROADMAP
- Agentes: pr-steward, domain advisors (craft, test, architecture)
- Workflows: validate-specs, quality, harness-checks
- Gate: `release_approval`

### enterprise

- Herda product + audit retention, LGPD checklist
- Scan reforçado de segredos
- Trilha `.arah/audit/` obrigatória

### open-source

- CONTRIBUTING, CODE_OF_CONDUCT templates
- Gates públicos documentados

## Validação pós-instalação

```powershell
./harness/scripts/doctor-harness.ps1 -Target ../meu-repo
./harness/scripts/validate-specs.ps1 -Target ../meu-repo
./harness/scripts/validate-agent-graph.ps1 -Target ../meu-repo
```

## Schema

[`schemas/arah-harness/harness-model.schema.yaml`](../schemas/arah-harness/harness-model.schema.yaml) · [`harness-profile.schema.yaml`](../schemas/arah-harness/harness-profile.schema.yaml)

Ver também: [MODEL.md](MODEL.md), [GOVERNANCE.md](GOVERNANCE.md), [INSTALL.md](INSTALL.md).

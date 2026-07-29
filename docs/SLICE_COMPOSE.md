# Slice Compose — fatia composta (experimental)

**Status:** experimental · **CLI:** `arah slice plan` · **Skill:** `slice-compose`  
**Desde:** 0.4.4

## Visão

Cada agent (ex. **qa** / **test-architect**) mantém um **backlog próprio** com gaps na sua lente (ex.: sem % de coverage) e metas mensuráveis (`goal` / `acceptance`).

Na hora de planejar uma fatia de produto (“faz outra slice”), o fluxo **não** é só sidecar passivo: ele **lê e propõe** itens dos agent-BL.

```text
fatia composta =
    backlog produto (E1-Sx)
  ∩ prioridades de lançamento dos agents (P0/P1 + goals)
  ∩ sugestões do usuário (“quero X/Y”)
```

O que não cabe agora vai para **Deferred** (ou fatia irmã).

## Input → Output

| Input | Fonte |
|-------|--------|
| Product story | `docs/BACKLOG*.md` / `ROADMAP.md` (heading `### E1-S4 — …`) |
| Agent vision backlogs | `docs/_arah/visions/*.backlog.yaml` (ou `.arah/visions/`) |
| User suggestions | `-Suggestions` / `-SuggestionsFile` |

| Output | Path |
|--------|------|
| Slice plan | `docs/_arah/slice-plans/<SliceId>.md` |

### Seções do plano

1. **Product** — story, priority, DoD  
2. **Agent priorities pulled** — itens puxados / “se couber”  
3. **User suggestions**  
4. **Deferred**  
5. **Executor / consultants** — 1 executor + consultas (ECP)

## Metas mensuráveis no agent-BL

Campos opcionais no sidecar YAML:

```yaml
  - id: qa-BL-005
    title: "Coverage mínimo unitário no domínio (MVP)"
    status: todo
    priority: P1
    goal: "coverage ≥ 50% em packages/domain (configurável; default 50%)"
    acceptance: "Script test:coverage + relatório c8/nyc; baseline; CI warn opcional"
    links: []
```

Default sugerido: **≥50%** no escopo unitário de `packages/domain` primeiro; expandir depois. Configurável por repo (ajuste o `goal` no backlog).

## CLI

```powershell
# Compor próxima fatia
powershell -File path\to\arah-harness\cli\arah.ps1 slice plan `
  -Target . `
  -SliceId E1-S4 `
  -Suggestions "quero coverage no domain nesta fatia se couber" `
  -VisionDir docs/_arah/visions `
  -OutDir docs/_arah/slice-plans `
  -Executor backend `
  -Consultants "solutions-architect,qa,test-architect" `
  -Force

# Dry-run (imprime o markdown)
powershell -File …\arah.ps1 slice plan -SliceId E1-S4 -DryRun
```

Convenção mínima se a CLI do consumidor ainda não tiver o comando: chamar o script direto:

```powershell
powershell -File scripts/agents/slice-compose.ps1 -SliceId E1-S4 -RepoRoot .
```

## Relação com Execution Control e Visions

| Fase | Artefato | Entrega? |
|------|----------|----------|
| Assessment | `.arah/visions/` / `docs/_arah/visions/` | Não |
| **Slice compose** | `docs/_arah/slice-plans/` | Não — só plano |
| Entrega | `arah task create` + executor | Sim |

Priorização pelo **lançamento**: P0/P1 e itens com `goal` entram como candidatos; P2+ e fora de alinhamento vão para Deferred.

## Heurística de pull (aceitável no experimental)

- Match de tokens das sugestões do usuário com title/goal  
- Coverage/TEA + stories E1/domain/ports/adapter → **se couber**  
- P0 → **puxar**; P1 em épico E1 → **se couber**

Revisão humana / executor decide o escopo final da fatia.

## Testes

```powershell
powershell -File scripts/harness/test-slice-compose.ps1
```

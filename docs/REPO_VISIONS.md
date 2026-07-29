# Repo Perspective Assessment (experimental)

**Status:** experimental · **CLI:** `arah assess-repo` · **Skill:** `repo-perspective-assess`  
**Desde:** 0.4.2 · **Schema visão:** v2 (0.4.3 — opinion + backlog + memory)

## Problema

Na primeira interação com um repositório, cada personalidade (QA, arquitetura, backend, security, domínio…) precisa responder:

1. **Qual minha opinião / escopo neste tipo de aplicação?** (não só checklist de pastas)
2. **Como está (As-Is) do meu ponto de vista?**
3. **Quais gaps vejo?**
4. **Qual meu action plan?**
5. **Qual meu backlog próprio** (stories com IDs estáveis) — e como atualizá-lo depois?

Isso vale também para **repo vazio** (visão de bootstrap por papel). Time dinâmico com memória, premissas Clean Architecture e especialidade de domínio.

## Solução

`assess-repo` varre evidências leves (stack, pastas, testes, CI, specs, ADRs, app_type) e, **em série**, gera artefatos por agent:

```text
.arah/visions/
  README.md
  summary.yaml
  qa.md
  qa.backlog.yaml      # merge-safe
  qa.events.yaml       # append-only
  solutions-architect.md
  …
```

Opcional: `-OutDir docs/_arah/visions` se quiser versionar sob `docs/`.

### Schema v2 por visão (Markdown)

| Seção | Conteúdo |
|-------|----------|
| **Opinião** | Perspectiva do papel no `app_type` + lente Clean Architecture / especialidade |
| **As-Is** | Sinais scoped ao mandato |
| **Gaps** | Déficits do ponto de vista do papel |
| **Action plan** | Plano curto (ex-To-Be) |
| **Backlog** | Tabela `{agent}-BL-NNN` · priority · status · title |
| **Memory / events** | Últimos eventos do feedback loop |

## Relação com Execution Control

| Fase | Modelo |
|------|--------|
| **Bootstrap / assessment** | Vários pareceres **em série** (exceção controlada) — visão + backlog, não entrega |
| **Entrega de produto** | 1 `primary_executor` + consultas limitadas ([EXECUTION_CONTROL.md](EXECUTION_CONTROL.md)) |
| **Feedback loop** | `assess-repo -Refresh` / `vision update` / `backlog sync` mescla backlog e registra evento |

A assessment **não** é uma entrega. Não altera código de produto.

## CLI

```powershell
# Gerar / forçar reescrita
powershell -File path\to\arah-harness\cli\arah.ps1 assess-repo -Target .

# Filtrar agents
powershell -File …\arah.ps1 assess-repo -Agents "qa,backend,security,solutions-architect"

# Refresh: regenera narrativa + mescla backlog (não apaga done/cancelled) + evento
powershell -File …\arah.ps1 assess-repo -Refresh -OutDir docs/_arah/visions

# Aliases do feedback loop
powershell -File …\arah.ps1 vision update -Target .
powershell -File …\arah.ps1 backlog sync -Target . -Agents qa

# Dry-run
powershell -File …\arah.ps1 assess-repo -DryRun
```

Flags do script: `-IncludeSpecialists`, `-SkipIndex`, `-RepoRoot`, `-Force`, `-Refresh`, `-BacklogOnly`.

Domain agents em `.agents/domain/` entram por padrão.

### Merge de backlog

- IDs estáveis: `qa-BL-001`, `security-BL-003`, …
- Status `done` / `cancelled` **nunca** são sobrescritos pelo Refresh
- Novos seeds entram; itens `todo`/`doing` podem atualizar título/priority
- Sidecar `*.backlog.yaml` é a fonte de merge; MD espelha a tabela

## Pós-install (opt-in)

`arah install` **não** dispara assessment automaticamente. Após install:

```powershell
arah domain sync
arah assess-repo -Force
# depois, em ciclos:
arah vision update
```

## Templates por personalidade (lentes CA)

| Agent | Lente |
|-------|--------|
| `qa` | Pirâmide TEA — unit domínio / integração ports / e2e adapters |
| `test-architect` | TEA risk-based + NFRs |
| `solutions-architect` | Boundaries CA + ADRs + contratos |
| `backend` | Ports & adapters |
| `frontend` | UI = adapter delivery |
| `security` | ASVS lite + threat model nas bordas |
| `clean-craft-advisor` | Regras sem UI; DIP |
| `architecture-documenter` | C4 + glossário |
| domains | Invariantes no núcleo; adapters não ditam regras |

Demais agents usam lente default. Cada lente traz `opinion_seed`, `action_plan` e `backlog_seeds`.

## App type (heurística)

`bootstrap-empty` · `multi-surface-product` · `web-frontend-heavy` · `api-backend` · `domain-centric-monorepo` · `docs-first` · `general-codebase`

Alimenta a seção **Opinião** de cada papel.

## Limitações (aceitáveis no experimental)

- Heurísticas por arquivos/pastas — não é LLM semântico
- a11y e threat model são conservadores (costumam marcar gap)
- Paths de escopo checam só o 1º segmento
- Opinião é template + app_type — revisão humana reforça o parecer
- Não substitui `discover` / `organism bootstrap`

## Testes

```powershell
powershell -File scripts/harness/test-assess-repo.ps1
```

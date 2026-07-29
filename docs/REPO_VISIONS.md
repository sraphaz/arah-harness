# Repo Perspective Assessment (experimental)

**Status:** experimental · **CLI:** `arah assess-repo` · **Skill:** `repo-perspective-assess`  
**Desde:** 0.4.2 (branch experimental; pin local até release)

## Problema

Na primeira interação com um repositório, cada personalidade (QA, arquitetura, backend, security, domínio…) precisa responder:

1. **O que é este repo na minha cabeça?**
2. **Quais gaps vejo do meu ponto de vista?**
3. **Qual o plano As-Is → To-Be?**

Isso vale também para **repo vazio** (visão de bootstrap por papel).

## Solução

`assess-repo` varre evidências leves (stack, pastas, testes, CI, specs, ADRs) e, **em série**, gera um artefato por agent:

```text
.arah/visions/
  README.md
  summary.yaml
  qa.md
  solutions-architect.md
  backend.md
  security.md
  test-architect.md
  …
```

Opcional: `-OutDir docs/_arah/visions` se quiser versionar sob `docs/`.

## Relação com Execution Control

| Fase | Modelo |
|------|--------|
| **Bootstrap / assessment** | Vários pareceres **em série** (exceção controlada) — só artefatos de visão |
| **Entrega de produto** | 1 `primary_executor` + consultas limitadas ([EXECUTION_CONTROL.md](EXECUTION_CONTROL.md)) |

A assessment **não** é uma entrega. Não altera código de produto. Não faz merge.

## CLI

```powershell
# No consumidor (cwd = repo) ou com -Target
powershell -File path\to\arah-harness\cli\arah.ps1 assess-repo -Target .

# Filtrar agents
powershell -File …\arah.ps1 assess-repo -Agents "qa,backend,security,solutions-architect"

# Forçar reescrita + saída sob docs
powershell -File …\arah.ps1 assess-repo -Force -OutDir docs/_arah/visions

# Dry-run
powershell -File …\arah.ps1 assess-repo -DryRun
```

Flags do script: `-IncludeSpecialists`, `-SkipIndex`, `-RepoRoot`.

Domain agents em `.agents/domain/` entram por padrão.

## Pós-install (opt-in)

`arah install` **não** dispara assessment automaticamente (evita ruído). Após install:

```powershell
arah domain sync
arah assess-repo -Force
```

Sugestão impressa nos next steps do install.

## Template por personalidade

Cada visão inclui:

- Headline do papel (“na cabeça do teste / da arquitetura / …”)
- Perguntas-guia
- **As-Is** (sinais positivos)
- **Gaps** (sinais negativos + paths de escopo ausentes)
- **To-Be** (plano curto)
- Autonomia sugerida (gates/padrões vs execução ECP)

Lentes especiais: `qa`, `test-architect`, `solutions-architect`, `backend`, `frontend`, `security`, `docs-steward`, `planner`, `orchestrator`, `spec-steward`, `pr-steward`, `release`, `clean-craft-advisor`, `architecture-documenter`. Demais agents usam lente default; domains usam lente de domínio (+ override se o id for conhecido).

## Repo vazio

Se não houver código/stack detectável (`emptyish`), o modo vira `bootstrap-empty`: gaps enfatizam bootstrap e cada papel propõe To-Be inicial.

## Limitações (aceitáveis no experimental)

- Heurísticas por arquivos/pastas — não é LLM semântico
- a11y e threat model são conservadores (costumam marcar gap)
- Paths de escopo checam só o 1º segmento
- Não substitui `discover` / `organism bootstrap` — complementa na dimensão “perspectiva de papel”

## Testes

```powershell
powershell -File scripts/harness/test-assess-repo.ps1
```

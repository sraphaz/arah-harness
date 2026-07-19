# Referências de mercado — ARAH Harness

Síntese das referências usadas para desenhar o ARAH Harness e o posicionamento vs o ecossistema Arah e o mercado agentic (2026).

**Leitura complementar:** [TECHORGANISM.md](TECHORGANISM.md) — como a v0.3 traduz autonomia sem perder auditoria.

## Referências principais

### GitHub Spec Kit
- **URL**: https://github.com/github/spec-kit
- **Distribuição**: CLI `specify init` / `uvx specify-cli init`
- **Força**: Constitution → Specify → Plan → Tasks → Implement; specs executáveis; agent-agnostic
- **Gap**: Sem multi-agente, coreografia, domínio consultivo
- **ARAH adota**: SDD com gates; modelo de bootstrap via CLI

### BMAD-METHOD
- **URL**: https://github.com/bmad-code-org/BMAD-METHOD (48k+ stars)
- **Força**: Squad de 21+ agentes por papel (Analyst, PM, Architect, Dev)
- **Gap**: Handoffs manuais; sem coreografia por paths
- **ARAH adota**: Papéis operacionais distintos; rejeita hierarquia profunda

### autonomous-sdlc
- **URL**: https://github.com/bitbitcodes/autonomous-sdlc
- **Força**: `sdlc init` instala 40 agentes; 11 quality gates; CONTINUITY.md
- **Gap**: Genérico; sem agentes de domínio de negócio
- **ARAH adota**: Bootstrap em `.agents/`; gates por fase

### harnessforge
- **URL**: https://github.com/jcaiagent7143-ui/harnessforge
- **Força**: `harness init` determinístico; multi-IDE; `sync --check` drift
- **Gap**: Só camada de contexto, sem orquestração
- **ARAH adota**: `sync-check` no CI; adaptadores Cursor

### keegoid/agentic-project-scaffold
- **URL**: https://github.com/keegoid/agentic-project-scaffold
- **Força**: Separação agentes (razão) vs scripts (determinismo) vs playbooks
- **ARAH adota**: `.skills/` = procedimentos; `scripts/` = execução determinística

### copier-coding-harness / product-template
- **Força**: Overlay em repos existentes; `copier update` para evoluir derivados
- **ARAH adota**: `update` + pin `.arah-version`; drift-check; **notificação** via GitHub Releases + cron no consumidor ([UPDATE_NOTIFICATIONS.md](UPDATE_NOTIFICATIONS.md)) — equivalente prático a “há update disponível” sem Dependabot no kernel

### Thread AI — Agentic SDLC Harness
- **URL**: https://www.threadai.com/blog/an-inside-look-how-we-built-our-agentic-sdlc-harness
- **Força**: Harness = superfície determinística (schemas, manifests, ledgers, routers)
- **ARAH adota**: Manifests YAML + gates + agent graph como vocabulário auditável

## Matriz comparativa

| Capacidade | Spec Kit | BMAD | autonomous-sdlc | harnessforge | **ARAH** |
|---|---|---|---|---|---|
| CLI bootstrap | ✅ | ✅ | ✅ | ✅ | ✅ |
| Multi-agente | ❌ | ✅ | ✅ | ❌ | ✅ |
| Coreografia paths | ❌ | ❌ | ❌ | ❌ | ✅ |
| Domínio consultivo | ❌ | ❌ | ❌ | ❌ | ✅ |
| SDD + harness | ✅ | parcial | ✅ | ❌ | ✅ |
| Drift check | ❌ | ❌ | parcial | ✅ | ✅ |
| Economia tokens | parcial | ❌ | parcial | ✅ | ✅ |
| Discovery + evolve (TechOrganism) | ❌ | parcial | parcial | ❌ | ✅ |

## Validação no Arah (repo origem)

O repositório `CursorRepos/arah` já documenta alinhamento com mercado em
`docs/ops/AGENT_STRATEGY_VALIDATION.md` — veredito: **alinhado 2026, à frente em coreografia e domínio**.

O ARAH Harness generaliza esse modelo para qualquer stack/projeto.

## Decisões conscientes (não adotar)

- **Tessl spec-as-source** (código 100% gerado): não determinístico para domínios críticos.
- **Merge automático**: contraria "humano comanda".
- **Agentes que criam agentes em silêncio**: degrada contexto e auditabilidade.

### Reinterpretação (v0.3 TechOrganism)

ARAH **não** permite spawn livre de agentes. Em vez disso:

1. `discover` / `evolve` **propõem** células, skills e coreografia em `docs/_meta/*.proposed.yaml`
2. `-Apply` só mescla candidaturas revisáveis em `arah.config.yaml`
3. Humano abre PR → CI → merge

Isso dá autonomia de descoberta sem perder o ledger auditável — alinhado a Thread AI (superfície determinística) e oposto a hierarquias opacas.

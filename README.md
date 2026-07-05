# ARAH Harness

**ARAH** (*Agent Runtime Autonomous Harness*) é o kernel reutilizável extraído do ecossistema [Arah](https://github.com/sraphaz/arah). Ele instala, em qualquer repositório, a camada que hoje vocês copiam manualmente: agentes, skills, coreografia, gates, harness SDD, agent graph, hooks e scripts de orquestração.

O repositório de produto fica enxuto; o método ARAH fica versionado aqui e chega via `init` / `update`.

## Problema que resolve

Cada projeto novo exigia replicar:

- manifests YAML (`.agents/`, `.skills/`)
- coreografia path-based (`choreography.yaml`)
- scripts PowerShell (`scripts/agents/`, `scripts/harness/`)
- hooks Cursor, rules escopadas, workflows CI
- specs SDD, gates, Definition of Done

Com ARAH Harness, isso vira **uma instalação + configuração de domínio**, não um copy-paste.

## O que o mercado faz vs o que ARAH traz

| Referência | Distribuição | Multi-agente | Coreografia por paths | Domínio consultivo | Gates + evidência |
|---|---|---|---|---|---|
| GitHub Spec Kit | CLI `specify init` | ❌ | ❌ | ❌ | parcial |
| BMAD-METHOD | npm / clone | ✅ (21+ roles) | ❌ | ❌ | parcial |
| autonomous-sdlc | `sdlc init` | ✅ (40 agents) | ❌ | ❌ | ✅ |
| harnessforge | `harness init` | ❌ | ❌ | ❌ | drift-check |
| **ARAH Harness** | `arah init` / `update` | ✅ | ✅ | ✅ | ✅ |

ARAH combina o que nenhum scaffold genérico oferece hoje: **coreografia path-based**, **agentes de domínio consultivos** (parecer sem PR), **agent graph auditável** e **economia de tokens** (contexto sob demanda).

## Quick start

```powershell
# Clonar o harness (ou usar como submodule)
git clone https://github.com/sraphaz/arah-harness.git

# Instalar no seu projeto
cd C:\path\to\meu-projeto
pwsh C:\path\to\arah-harness\cli\arah.ps1 init

# Validar instalação
pwsh .\scripts\agents\validate-manifests.ps1
pwsh C:\path\to\arah-harness\cli\arah.ps1 doctor
```

Depois de `init`, edite `arah.config.yaml` (nome, stack, comandos de teste, domínios) e rode:

```powershell
powershell -File path/to/arah-harness/cli/arah.ps1 domain sync
powershell -File ./scripts/agents/validate-manifests.ps1
powershell -File path/to/arah-harness/cli/arah.ps1 export-graph
```

## Estrutura deste repositório

```
arah-harness/
├── kernel/           # Camada copiada para projetos-alvo (versionada)
│   ├── .agents/
│   ├── .skills/
│   ├── .cursor/
│   └── scripts/
├── templates/        # Templates renderizados no init (domínio, AGENTS.md, specs)
├── cli/              # arah init | update | doctor | sync-check | domain add
└── docs/             # METHOD.md, MARKET_REFERENCE.md, BOOTSTRAP.md
```

## Princípios (herdados do Arah, generalizados)

1. **Humano comanda, agente executa** — merge sempre humano.
2. **Tudo via Pull Request** — sem commit direto em `main`.
3. **Escopo mínimo** — cada agente só toca paths permitidos.
4. **Spec-before-code** — fases exigem spec em `docs/specs/` e `Spec-Id:` no PR.
5. **Contexto sob demanda** — regras/skills carregadas quando relevantes; comunicação entre agentes é passiva (arquivo + CI).
6. **Kernel imutável localmente** — customizações vão em `arah.config.yaml` e `.agents/domain/`, não editando scripts do kernel sem override explícito.

## Documentação

- [docs/METHOD.md](docs/METHOD.md) — método ARAH completo
- [docs/MARKET_REFERENCE.md](docs/MARKET_REFERENCE.md) — referências de mercado e decisões
- [docs/BOOTSTRAP.md](docs/BOOTSTRAP.md) — checklist pós-init para projetos novos
- [docs/MIGRATION_FROM_ARAH.md](docs/MIGRATION_FROM_ARAH.md) — migrar o repo Arah para consumir o harness
- [kernel/.agents/README.md](kernel/.agents/README.md) — catálogo de agentes do kernel

## Origem

Extraído e generalizado a partir de `CursorRepos/arah` (validação vs mercado em `docs/ops/AGENT_STRATEGY_VALIDATION.md` no repo Arah).

## Licença

MIT — ver [LICENSE](LICENSE).

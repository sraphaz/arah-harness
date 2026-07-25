# Graphify × ARAH — avaliação e arquitetura

**Status:** accepted (integração opcional, faseada)  
**Data:** 2026-07-25  
**Spec-Id:** `arah-graphify-knowledge-graph`  
**Fonte externa:** [safishamsi/graphify](https://github.com/safishamsi/graphify) / PyPI `graphifyy` · CLI `graphify`

---

## Veredito

**Sim — integrar Graphify como grafo de conhecimento do corpus (código/docs/mídia), complementar ao Agent Graph.**  
Não substituir `docs/_meta/agent-graph.generated.json`. Os dois grafos respondem perguntas diferentes.

| Grafo | Pergunta | Fonte | Dependência |
|-------|----------|-------|-------------|
| **Agent Graph** | Quem colabora com quem no harness? | manifests, skills, coreografia, gates | nativo ARAH |
| **Knowledge Graph (Graphify)** | O que o repositório *contém* e como se conecta? | AST (tree-sitter) + opcional LLM em docs/mídia | opcional (`graphify`) |

---

## Por que faz sentido

1. **Discover fraco em monorepos** — `discover-repo.ps1` é heurístico (arquivos-raiz). Graphify entrega comunidades e arestas estruturais reais sem gastar token em código (`--code-only`).
2. **Lacuna `semantic-discovery`** — já planejada em `capabilities.yaml`; Graphify é o candidato concreto para preenchê-la.
3. **C-12 Graph Explorer** — o console já serve Agent Graph; um segundo artefato navegável (comunidades + query) encaixa no backlog sem redesenhar o modelo de agentes.
4. **Local-first alinhado à governança** — código 100% local; passagem semântica só quando o operador optar (API key / Ollama).
5. **Skill ecosystem** — Graphify já se registra em Cursor (`graphify cursor install`); ARAH pode orquestrar o mesmo artefato no ciclo TechOrganism.

## Por que *não* acoplar no núcleo

- Dependência Python/tree-sitter fora do kernel PS 5.1.
- Pacote PyPI `graphifyy` (nome transitório) — pin e detecção de CLI, não import embutido.
- Passagem LLM em docs/imagens implica custo, privacidade e chaves — deve ser opt-in.
- CI não deve falhar se Graphify estiver ausente (exit skip).

---

## Arquitetura alvo

```text
                    ┌─────────────────────────┐
  manifests ───────►│ Agent Graph (nativo)    │──► docs/_meta/agent-graph.generated.*
                    └─────────────────────────┘
  código/docs ─────►│ Knowledge Graph (Graphify) │──► graphify-out/graph.json
                    │  + manifest ARAH           │──► docs/_meta/knowledge-graph.manifest.yaml
                    └─────────────────────────┘
                              │
              discover / evolve / Live Console (C-12)
```

### Contratos

| Artefato | Versionar? | Notas |
|----------|------------|-------|
| `docs/_meta/knowledge-graph.manifest.yaml` | Sim | Ponte ARAH ↔ Graphify (status, modo, paths) |
| `graphify-out/graph.json` | Opcional (equipe) | Mapa compartilhado; ver docs Graphify |
| `graphify-out/GRAPH_REPORT.md` | Opcional | Highlights para agentes/humanos |
| `graphify-out/cost.json`, `cache/` | Não | gitignore |

### Superfície CLI

```powershell
arah knowledge-graph              # code-only se CLI presente; senão skip (0)
arah knowledge-graph status       # presença CLI + artefatos
arah knowledge-graph -Mode full   # opt-in semântico (requer backend LLM)
```

### Fases

| Fase | Entrega | Gate |
|------|---------|------|
| **0 (este PR)** | ADR + spec + adapter opcional + capability experimental | self-test sem Graphify instalado |
| **1** | Discover/evolve leem comunidades do manifesto / report | propostas em `*.proposed.yaml` |
| **2** | Live Console / C-12 dual-pane (Agent ∪ Knowledge) | schema console |
| **3** | MCP `graphify.serve` opcional no control plane | AuthZ |

---

## Riscos e mitigação

| Risco | Mitigação |
|-------|-----------|
| Dep Python / PATH | Detecção de `graphify` / `uv tool`; `-Require` só quando o operador exige |
| Custo LLM | Default `--code-only`; `Mode full` explícito |
| Drift de API Graphify | Manifest versionado; adapter fino; sem fork |
| Confusão Agent vs Knowledge | Docs + nomes distintos; console rotula os dois |
| Prompt-cache IDE | Ignorar `cost.json`/`cache/`; documentar `.claudeignore` |

---

## Decisão

Aceitar Graphify como **provedor opcional de knowledge graph** sob o ciclo TechOrganism (`discover` → homeostase → exploração). Manter Agent Graph como fonte de verdade da *colaboração* entre células.

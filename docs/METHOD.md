# Método ARAH

**ARAH** — *Agent Runtime Autonomous Harness*

Framework de bootstrap para repositórios gerenciados por agentes: autonomia coreografada, auditável, observável, com economia de tokens.

## Camadas

```
┌─────────────────────────────────────────────────────────┐
│  Projeto (código de produto)                            │
├─────────────────────────────────────────────────────────┤
│  Camada de domínio (.agents/domain/, choreography)      │  ← específico do negócio
├─────────────────────────────────────────────────────────┤
│  Kernel ARAH (.agents operacionais, .skills, scripts)   │  ← arah-harness (versionado)
├─────────────────────────────────────────────────────────┤
│  Config (arah.config.yaml, AGENTS.md)                   │  ← por repositório
└─────────────────────────────────────────────────────────┘
```

## Fluxo operacional

1. Humano define intenção (issue, fase, pedido).
2. **Orquestrador** roteia para agente operacional.
3. **Coreografia** co-ativa agentes de domínio (parecer consultivo) e skills.
4. Agente implementa dentro do escopo → abre PR.
5. **QA** + **PR Steward** + CI + gates SDD.
6. Humano faz merge → **next-phase** (opcional).

## Tipos de agente

| Tipo | Codifica? | Papel |
|------|-----------|-------|
| Operacional | Sim (PR) | backend, frontend, spec-steward, release… |
| Domínio | Não | Parecer de negócio via coreografia path-based |
| Specialist | Não | Profundidade técnica (stack) |

## SDD (Spec-Driven Development)

- Specs em `docs/specs/*.spec.yaml` com acceptance EARS.
- PR de fase exige `Spec-Id:` no corpo.
- Harness executa comandos declarados na spec.
- Gate `validate-specs.ps1`: AC `covered` exige `covered_by`.

## Comunicação passiva (economia de tokens)

- Hook Cursor `stop` → `.cursor/domain-review.md` (sem followup_message).
- CI publica pareceres de domínio como comentários no PR.
- `AGENTS.md` enxuto; detalhes em `docs/ops/`.
- **Sinais tipados** (v0.3+): `.arah/bus/signals.jsonl` — attract/consult/propose/coalesce/evolve — auditáveis, não chat obrigatório.

## Biocomponente (v0.3+)

O harness observa o repositório, propõe células e evolui por seleção via PR:

| Comando | Função |
|---------|--------|
| `arah discover` | Percebe stack/domínio → `discovery.proposed.yaml` |
| `arah organism bootstrap` | Ritual ontogênico → `organism.manifest.yaml` |
| `arah organism signal` | Emite sinal tipado no bus |
| `arah evolve` | Self-learning → `evolution.proposed.yaml` |
| `arah regenerate` | Homeostase: update + sync + discover + organism + evolve + graph + doctor |

Princípio: **agentes propõem; humanos aplicam.** Ver [BIOCOMPONENT.md](BIOCOMPONENT.md).

## Agent Graph

O grafo (agentes ↔ skills ↔ paths ↔ gates) é exportável para auditoria.
No Arah completo: `export-agent-graph.ps1` + `docs/_meta/agent-graph.generated.json`.
No kernel v0.1: coreografia + validate-manifests.

## Distribuição

| Comando | Função |
|---------|--------|
| `arah init` | Instala kernel no repo-alvo |
| `arah domain sync` | Gera agentes de domínio + `choreography.domains.yaml` |
| `arah export-graph` | Exporta agent graph (JSON + Mermaid) |
| `arah update` | Reaplica kernel (preserva config/domínio) |
| `arah sync-check` | Detecta drift vs kernel (CI) |
| `arah doctor` | Valida instalação |
| `arah regenerate` | Homeostase biocomponente no consumidor |

## O que NÃO fazer

- Merge automático com CI verde.
- Hierarquia multi-nível de agentes (degrada auditabilidade).
- Criação silenciosa de agentes (use discovery/evolve + Apply + PR).
- Código 100% gerado da spec sem revisão humana (domínios financeiros/críticos).

## Evolução

1. **v0.1**: kernel genérico + CLI init/update/sync-check.
2. **v0.2**: `domain sync`; agent graph; Live Session; harness-model.
3. **v0.3** (atual): biocomponente — discover, organism, signal bus, evolve, regenerate.
4. **v1.0**: Migrar repo Arah para consumir harness como dependência (prova real).

# Método ARAH

**ARAH** — *Agent Runtime Autonomous Harness*  
**Versão do método** alinhada ao harness **0.3.0**

Framework de bootstrap para repositórios gerenciados por agentes: autonomia coreografada, auditável, observável, com economia de tokens — e dimensão **TechOrganism** para discovery, organização e evolução contínua.

---

## Promessa

Um humano define a intenção. O organismo ARAH roteia, co-ativa, implementa, valida e propõe a próxima melhoria. O humano faz o merge.

```text
Intenção → Orquestrador → Célula + skills → PR → CI + gates → ready-for-merge → Merge (humano)
                │
                └── TechOrganism: discover · organism · signals · evolve · regenerate
```

---

## Camadas

```
┌──────────────────────────────────────────────────────────┐
│  Projeto — código de produto                             │
├──────────────────────────────────────────────────────────┤
│  Domínio — .agents/domain/, specialists, overlays        │  ← negócio / stack
├──────────────────────────────────────────────────────────┤
│  Organismo — docs/_meta/*, .arah/bus, evolution          │  ← TechOrganism v0.3
├──────────────────────────────────────────────────────────┤
│  Kernel ARAH — operacionais, .skills, scripts, hooks     │  ← versionado
├──────────────────────────────────────────────────────────┤
│  Config — arah.config.yaml, AGENTS.md, harness-profile   │  ← por repositório
└──────────────────────────────────────────────────────────┘
```

| Camada | Mutável por | Persistência |
|--------|-------------|--------------|
| Kernel | `arah update` / `regenerate -UpdateKernel` | Pin `.arah-version` |
| Config | Humano (+ Apply revisável) | Git |
| Organismo | TechOrganism (propostas) + PR | `docs/_meta/` versionado |
| Domínio | `domain sync` a partir da config | `.agents/domain/` gerado |
| Produto | Agentes operacionais em PR | Código da app |

---

## Fluxo operacional

1. Humano define intenção (issue, fase, pedido).
2. **Orquestrador** roteia para a célula operacional.
3. **Coreografia** co-ativa domínio/specialists e skills.
4. Célula implementa no escopo → abre PR.
5. **QA** + **PR Steward** + CI + gates SDD.
6. Humano faz merge → **next-phase** (opcional).
7. Periodicamente: **evolve** / **regenerate** → novas propostas.

---

## Tipos de agente (células)

| Tipo | Codifica? | Papel | Autonomia máx. típica |
|------|-----------|-------|------------------------|
| Operacional | Sim (via PR) | backend, frontend, qa, release… | `activate` / `invoke_skill` |
| Domínio | Não | Parecer de negócio path-based | `consult` |
| Specialist | Não | Profundidade de stack | `consult` |

Catálogo operacional: [`.agents/README.md`](../.agents/README.md).  
Mapa vivo: `docs/_meta/organism.manifest.yaml`.

---

## Spec-Driven Development

- Specs em `docs/specs/` com acceptance EARS.
- PR de fase exige `Spec-Id:` no corpo.
- Harness executa comandos declarados na spec.
- Gate `validate-specs`: AC `covered` exige `covered_by`.

Registry: [`docs/specs/REGISTRY.md`](specs/REGISTRY.md).

---

## Comunicação — economia de tokens

ARAH privilegia **contexto sob demanda**, não chat entre agentes:

| Canal | Quando |
|-------|--------|
| Hook `stop` → `.cursor/domain-review.md` | Fim de turno Cursor |
| Comentário de PR (`post-domain-consult`) | CI / parecer de domínio |
| Agent Graph | Auditoria estrutural |
| **Sinais tipados** (v0.3) | `.arah/bus/signals.jsonl` |

Sinais (`attract`, `consult`, `propose`, `coalesce`, `evolve`, …) são artefatos append-only — não um swarm conversacional.

Detalhe: [TECHORGANISM.md](TECHORGANISM.md).

---

## TechOrganism

Capacidade que diferencia a v0.3: o harness **observa**, **organiza** e **evolui** o próprio arranjo de agentes.

| Comando | Função |
|---------|--------|
| `discover` | Stack + domínio → propostas |
| `organism bootstrap` | Células, tecidos, vias |
| `organism signal` | Comunicação tipada |
| `evolve` | Self-learning → propostas |
| `regenerate` | Homeostase no consumidor |

**Agentes propõem; humanos aplicam.**

---

## Distribuição e homeostase

| Comando | Função |
|---------|--------|
| `install` / `init` | Bootstrap do kernel |
| `domain sync` | Gera domínio + `choreography.domains.yaml` |
| `export-graph` | Agent Graph |
| `update` | Reaplica kernel (preserva config) |
| `sync-check` | Drift vs upstream |
| `doctor` | Saúde da instalação |
| `regenerate` | Update + sync + discover + organism + evolve + graph + doctor |

---

## Agent Graph

Grafo exportável: agentes ↔ skills ↔ paths ↔ gates.

- Artefato: `docs/_meta/agent-graph.generated.json`
- Diagrama: `docs/_meta/agent-graph.generated.mmd`
- Validação: `validate-agent-graph.ps1`

---

## O que o método rejeita

| Anti-padrão | Por quê |
|-------------|---------|
| Merge automático | Contraria “humano comanda” |
| Hierarquia profunda de agentes | Degrada contexto e auditoria |
| Spawn silencioso de agentes | Use discover/evolve + Apply + PR |
| Código 100% gerado da spec sem revisão | Inaceitável em domínios críticos |

---

## Evolução do harness

| Versão | Marco |
|--------|-------|
| v0.1 | Kernel + CLI init/update/sync-check |
| v0.2 | Domain sync, agent graph, Live Session, harness-model |
| **v0.3** | **TechOrganism** — discover, organism, signal bus, evolve, regenerate |
| v1.0 | Produto Arah consome harness como dependência (prova real) |

---

## Referências rápidas

- [TECHORGANISM.md](TECHORGANISM.md) — dimensão viva  
- [GOVERNANCE.md](GOVERNANCE.md) — autonomia e gates  
- [MODEL.md](MODEL.md) — contratos first-class  
- [INSTALL.md](INSTALL.md) — instalação  
- [MARKET_REFERENCE.md](MARKET_REFERENCE.md) — mercado  

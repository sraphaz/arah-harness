# Biocomponente ARAH

**Versão** 0.3.0 · **Spec** [`arah-biocomponent`](specs/arah-biocomponent.spec.yaml) · **Status** active

O ARAH Harness não é apenas um kit de arquivos. É um **biocomponente tecnológico**: um organismo versionado que se instala no repositório, observa como a aplicação funciona, propõe a distribuição certa de agentes e evolui com o próprio projeto — sem abrir mão de auditoria nem do merge humano.

---

## Visão

Quando o harness chega a um repositório, ele precisa responder a perguntas que antes só o humano respondia:

- Qual é o **domínio** desta aplicação?
- Quais **tecnologias** e superfícies importam?
- Quais **agentes** devem existir — e com qual papel?
- Como eles devem **se comunicar** e formar grupos?
- Como o organismo **melhora** a cada execução?

A v0.3 responde com um ciclo vivo: perceber → definir → sinalizar → evoluir → regenerar.

```text
                 ┌──────────────┐
                 │   humano     │
                 │  (intenção)  │
                 └──────┬───────┘
                        │
        ┌───────────────▼────────────────┐
        │         ORGANISMO ARAH         │
        │                                │
        │  discover ──► bootstrap        │
        │      │              │          │
        │      ▼              ▼          │
        │   propostas      células       │
        │                  tecidos       │
        │      │              │          │
        │      └──────┬───────┘          │
        │             ▼                  │
        │         sinais ──► evolve      │
        │             │                  │
        │             ▼                  │
        │        regenerate              │
        └─────────────┬──────────────────┘
                      │
                      ▼
              PR → CI → merge humano
                 (seleção natural)
```

---

## Metáfora operacional

A natureza inspira a *forma*; o ledger garante a *prova*.

| Natureza | ARAH | Artefato |
|----------|------|----------|
| Organismo | Harness instalado no repo | kernel + config + meta |
| Célula | Agente (operacional / domínio / specialist) | `.agents/**/*.agent.yaml` |
| Tecido | Grupo com tópico compartilhado | `organism.manifest.yaml` → `tissues` |
| Sinal químico | Mensagem tipada, append-only | `.arah/bus/signals.jsonl` |
| Ontogenia | Primeiro momento de definição | `arah organism bootstrap` |
| Homeostase | Manter o organismo saudável | `arah regenerate` |
| Evolução | Seleção via proposta + PR | `evolution.proposed.yaml` |

Não é metáfora decorativa: cada conceito mapeia para um comando, um schema e um caminho no disco.

---

## Princípio de governança

### Agentes propõem. Humanos aplicam.

| Permitido | Proibido |
|-----------|----------|
| Escrever propostas em `docs/_meta/` | Criar agentes em silêncio |
| Emitir sinais `propose` / `evolve` | Reescrever o kernel sem `update` |
| `-Apply` mesclar candidaturas revisáveis | Auto-merge com CI verde |
| Abrir PR com evolução | Hierarquia opaca de agentes |

Isso reinterpretá o anti-padrão de mercado *“agentes que criam agentes”* como um **loop de proposta + gate** — autonomia com auditabilidade.

Gates relevantes: `proposal_before_implementation`, `spec_before_work`, `release_approval`.

---

## Ciclo de vida

### 1. Percepção — `discover`

Observa manifests, linguagens, frameworks e estrutura de pastas.  
Gera `docs/_meta/discovery.proposed.yaml` com:

- stack detectada (evidências)
- domínios candidatos
- specialists candidatos
- nível de confiança

```powershell
arah discover              # só propõe
arah discover -Apply       # mescla ausentes em arah.config.yaml
arah discover -DryRun      # imprime sem gravar
```

### 2. Ontogenia — `organism bootstrap`

Ritual do **primeiro momento**: lê agentes existentes + discovery e declara o mapa vivo do repositório.

```powershell
arah organism bootstrap
arah organism bootstrap -Force   # redefinir após mudança grande
arah organism status
```

Saída: `docs/_meta/organism.manifest.yaml` + `.arah/organism/state.json`.

### 3. Comunicação — `organism signal`

Sinais tipados no barramento. Default **passivo** (economia de tokens): não há chat multi-turno obrigatório.

| Tipo | Semântica |
|------|-----------|
| `attract` | Solicita atenção / co-ativação |
| `consult` | Pedido ou entrega de parecer |
| `propose` | Proposta de mudança (requer Apply humano) |
| `acknowledge` | Confirma recebimento |
| `coalesce` | Forma ou reforça um tecido |
| `evolve` | Dispara ciclo de aprendizado |
| `status` | Heartbeat / homeostase |

```powershell
arah organism signal `
  -From orchestrator -To backend `
  -Type attract -Topic delivery
```

Ledger: `.arah/bus/signals.jsonl` (schema `arah-harness/signal`).

### 4. Aprendizado — `evolve`

Consome auditoria, sinais e telemetria Live.  
Gera `docs/_meta/evolution.proposed.yaml` com propostas de:

- definição de células
- skills
- coreografia
- workflows
- comunicação
- config de domínio

```powershell
arah evolve
arah evolve -Apply    # emite sinal evolve + reforça trilha (sem reescrever kernel)
```

### 5. Homeostase — `regenerate`

Pipeline unificado para o consumidor:

1. `update` do kernel *(opcional, `-UpdateKernel`)*
2. `domain sync`
3. `discover`
4. `organism bootstrap`
5. `evolve`
6. `export-graph`
7. `doctor`

```powershell
# Quem já usa ARAH — receber v0.3 e regenerar o organismo
powershell -File cli/arah.ps1 regenerate `
  -Target C:\meu-projeto `
  -UpdateKernel -Force
```

---

## Artefatos

| Artefato | Função | Versionar? |
|----------|--------|------------|
| `docs/_meta/discovery.proposed.yaml` | Observação + propostas | Sim (revisável) |
| `docs/_meta/organism.manifest.yaml` | Mapa de células e tecidos | Sim |
| `docs/_meta/evolution.proposed.yaml` | Propostas de self-learning | Sim |
| `docs/_meta/agent-graph.generated.json` | Grafo auditável | Sim (gerado) |
| `.arah/bus/signals.jsonl` | Barramento de sinais | Não (runtime) |
| `.arah/organism/state.json` | Estado ontogênico | Não (runtime) |
| `.arah/audit/events.jsonl` | Ledger de ações | Não (runtime) |

---

## Schemas canônicos

| Schema | Path |
|--------|------|
| discovery | [`schemas/arah-harness/discovery.schema.yaml`](../schemas/arah-harness/discovery.schema.yaml) |
| organism | [`schemas/arah-harness/organism.schema.yaml`](../schemas/arah-harness/organism.schema.yaml) |
| signal | [`schemas/arah-harness/signal.schema.yaml`](../schemas/arah-harness/signal.schema.yaml) |
| evolution | [`schemas/arah-harness/evolution.schema.yaml`](../schemas/arah-harness/evolution.schema.yaml) |

---

## Skills do biocomponente

| Skill | Script |
|-------|--------|
| `discover-repo` | `scripts/agents/discover-repo.ps1` |
| `evolve-harness` | `scripts/agents/evolve-harness.ps1` |
| `regenerate-harness` | `scripts/agents/regenerate-harness.ps1` |

Invocação: `./scripts/agents/invoke-skill.ps1 -Skill discover-repo`

---

## Fluxo recomendado por tipo de repo

### Greenfield

```text
install → regenerate → revisar docs/_meta → Apply seletivo → domain sync → primeiro Spec → primeira fase
```

### Brownfield

```text
install (sem -Force) → discover → organism bootstrap → evolve
→ PR só com propostas → humano escolhe → domain sync → regenerate periódico
```

### Consumidor já em ARAH &lt; 0.3

```text
git pull arah-harness → regenerate -UpdateKernel -Force
→ revisar discovery/evolution → PR de evolução do repo
```

---

## O que o biocomponente *não* é

- Não é um swarm conversacional sem ledger  
- Não é auto-merge  
- Não é geração cega de agentes a cada prompt  
- Não substitui spec-before-work nem Definition of Done  

É **autonomia com superfície determinística** — a mesma filosofia Thread AI / Spec Kit, com a vantagem ARAH de coreografia e domínio.

---

## Referências

| Doc | Por quê |
|-----|---------|
| [METHOD.md](METHOD.md) | Método e camadas |
| [GOVERNANCE.md](GOVERNANCE.md) | Níveis de autonomia e gates |
| [MODEL.md](MODEL.md) | Harness-model first-class |
| [OBSERVABILITY.md](OBSERVABILITY.md) | Telemetria e Live Session |
| [AUDIT.md](AUDIT.md) | Ledger de eventos |
| [MARKET_REFERENCE.md](MARKET_REFERENCE.md) | Posicionamento vs mercado |
| [INSTALL.md](INSTALL.md) | Como instalar / atualizar |
| Spec | [`arah-biocomponent.spec.yaml`](specs/arah-biocomponent.spec.yaml) |

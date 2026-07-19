# TechOrganism

**ARAH Harness · TechOrganism**  
**Versão** 0.3.0 · **Spec** [`arah-biocomponent`](specs/arah-biocomponent.spec.yaml) · **Status** active

**TechOrganism** é o nome da dimensão viva do ARAH Harness: um organismo tecnológico versionado que se instala no repositório, observa domínio e stack, propõe a distribuição certa de agentes, define como eles se comunicam e evolui com o próprio projeto — sem abrir mão de auditoria nem do merge humano.

> *Harness = kernel. TechOrganism = o ser que esse kernel torna vivo no seu repo.*

---

## Visão

Quando o TechOrganism chega a um repositório, ele responde ao que antes só o humano respondia:

- Qual é o **domínio** desta aplicação?
- Quais **tecnologias** e superfícies importam?
- Quais **agentes** devem existir — e com qual papel?
- Como eles devem **se comunicar** e formar grupos?
- Como o organismo **melhora** a cada execução?

Ciclo vivo: **perceber → definir → sinalizar → evoluir → regenerar**.

```text
                 ┌──────────────┐
                 │   humano     │
                 │  (intenção)  │
                 └──────┬───────┘
                        │
        ┌───────────────▼────────────────┐
        │         TECHORGANISM           │
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

| Natureza | TechOrganism | Artefato |
|----------|--------------|----------|
| Organismo | Harness instalado no repo | kernel + config + meta |
| Célula | Agente (operacional / domínio / specialist) | `.agents/**/*.agent.yaml` |
| Tecido | Grupo com tópico compartilhado | `organism.manifest.yaml` → `tissues` |
| Sinal químico | Mensagem tipada, arquivo-por-evento | `.arah/local/bus/` |
| Ontogenia | Primeiro momento de definição | `arah organism bootstrap` |
| Homeostase | Manter o organismo saudável | `arah regenerate` |
| Evolução | Seleção via proposta + PR | `evolution.proposed.yaml` |

Cada conceito mapeia para um comando, um schema e um caminho no disco.

---

## Princípio de governança

### Agentes propõem. Humanos aplicam.

| Permitido | Proibido |
|-----------|----------|
| Escrever propostas em `docs/_meta/` | Criar agentes em silêncio |
| Emitir sinais `propose` / `evolve` | Reescrever o kernel sem `update` |
| `-Apply` mesclar candidaturas revisáveis | Auto-merge com CI verde |
| Abrir PR com evolução | Hierarquia opaca de agentes |

O anti-padrão de mercado *“agentes que criam agentes”* vira um **loop de proposta + gate** — autonomia com auditabilidade.

Gates: `proposal_before_implementation` · `spec_before_work` · `release_approval`.

---

## Ciclo de vida

### 1. Percepção — `discover`

Observa manifests, linguagens, frameworks e estrutura.  
Gera `docs/_meta/discovery.proposed.yaml`.

```powershell
arah discover              # só propõe
arah discover -Apply       # mescla ausentes em arah.config.yaml
arah discover -DryRun      # imprime sem gravar
```

### 2. Ontogenia — `organism bootstrap`

Ritual do **primeiro momento**: declara o mapa vivo do repositório.

```powershell
arah organism bootstrap
arah organism bootstrap -Force
arah organism status
```

Saída: `docs/_meta/organism.manifest.yaml` + `.arah/organism/state.json`.

### 3. Comunicação — `organism signal`

Sinais tipados. Default **passivo** (economia de tokens).

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
  -From orchestrator -SignalTo backend `
  -SignalType attract -Topic delivery
```

Ledger (quente): `.arah/local/bus/` — ver [STATE_MODEL.md](STATE_MODEL.md).

### 4. Aprendizado — `evolve`

Consome auditoria, sinais e telemetria Live → `docs/_meta/evolution.proposed.yaml`.

```powershell
arah evolve
arah evolve -Apply
```

### 5. Economy Intelligence — `metrics`

```powershell
powershell -File cli/arah.ps1 metrics report
powershell -File cli/arah.ps1 metrics rollup -Digest
```

Scorecard em `.arah/observability/summary.yaml`; evolve consome o semaphore para propor eficiência. Guia: [ECONOMY.md](ECONOMY.md).

### 6. Homeostase — `regenerate`

```powershell
powershell -File cli/arah.ps1 regenerate `
  -Target C:\meu-projeto `
  -UpdateKernel -Force
```

Pipeline: update → domain sync → discover → organism → evolve → metrics rollup → graph → doctor.

---

## Artefatos

| Artefato | Função | Versionar? |
|----------|--------|------------|
| `docs/_meta/discovery.proposed.yaml` | Observação + propostas | Sim |
| `docs/_meta/organism.manifest.yaml` | Mapa de células e tecidos | Sim |
| `docs/_meta/evolution.proposed.yaml` | Propostas de self-learning | Sim |
| `docs/_meta/metrics.digest.md` | Digest de eficiência (opcional) | Opcional |
| `docs/_meta/agent-graph.generated.json` | Grafo auditável | Sim (gerado) |
| `.arah/local/bus/` | Barramento de sinais (quente) | Não (runtime) |
| `.arah/local/audit/` | Ledger de ações (quente) | Não (runtime) |
| `docs/_meta/runs/*/summary.json` | Evidência fria por run | Sim |
| `.arah/organism/state.json` | Estado ontogênico | Não (runtime) |
| `.arah/observability/summary.yaml` | Scorecard Economy Intelligence | Não (runtime) |

---

## Schemas

| Schema | Path |
|--------|------|
| discovery | [`schemas/arah-harness/discovery.schema.yaml`](../schemas/arah-harness/discovery.schema.yaml) |
| organism | [`schemas/arah-harness/organism.schema.yaml`](../schemas/arah-harness/organism.schema.yaml) |
| signal | [`schemas/arah-harness/signal.schema.yaml`](../schemas/arah-harness/signal.schema.yaml) |
| evolution | [`schemas/arah-harness/evolution.schema.yaml`](../schemas/arah-harness/evolution.schema.yaml) |

---

## Skills

| Skill | Script |
|-------|--------|
| `discover-repo` | `scripts/agents/discover-repo.ps1` |
| `evolve-harness` | `scripts/agents/evolve-harness.ps1` |
| `metrics-rollup` | `scripts/agents/metrics-rollup.ps1` |
| `regenerate-harness` | `scripts/agents/regenerate-harness.ps1` |

---

## Fluxos por tipo de repo

### Greenfield

```text
install → regenerate → revisar docs/_meta → Apply seletivo → domain sync → Spec → fase
```

### Brownfield

```text
install (sem -Force) → discover → organism bootstrap → evolve → PR de propostas → regenerate
```

### Consumidor ARAH &lt; 0.3

```text
git pull arah-harness → regenerate -UpdateKernel -Force → PR de evolução
```

---

## O que TechOrganism *não* é

- Swarm conversacional sem ledger  
- Auto-merge  
- Spawn cego de agentes a cada prompt  
- Substituição de spec-before-work ou Definition of Done  

É **autonomia com superfície determinística** — coreografia + domínio + evolução auditável.

---

## Nome e legado

| Termo | Uso |
|-------|-----|
| **TechOrganism** | Nome de produto da dimensão viva |
| *biocomponente* | Sinônimo técnico / legado em specs e commits |
| CLI `organism` | Superfície de comando estável |

---

## Referências

| Doc | Por quê |
|-----|---------|
| [METHOD.md](METHOD.md) | Método e camadas |
| [GOVERNANCE.md](GOVERNANCE.md) | Autonomia e gates |
| [MODEL.md](MODEL.md) | Harness-model |
| [OBSERVABILITY.md](OBSERVABILITY.md) | Telemetria |
| [AUDIT.md](AUDIT.md) | Ledger |
| [MARKET_REFERENCE.md](MARKET_REFERENCE.md) | Mercado |
| [INSTALL.md](INSTALL.md) | Instalação |
| Spec | [`arah-biocomponent.spec.yaml`](specs/arah-biocomponent.spec.yaml) |

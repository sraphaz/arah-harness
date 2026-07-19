# ARAH Harness · TechOrganism

[![CI](https://github.com/sraphaz/arah-harness/actions/workflows/ci.yml/badge.svg)](https://github.com/sraphaz/arah-harness/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.4.1-green.svg)](VERSION)
[![TechOrganism](https://img.shields.io/badge/TechOrganism-v0.3-0A7A5A.svg)](docs/TECHORGANISM.md)

**ARAH** — *Agent Runtime Autonomous Harness*  
**TechOrganism** — a dimensão viva: descobrir, organizar, comunicar, evoluir.

O kernel open-source que transforma qualquer repositório em um **organismo tecnológico de agentes** — coreografado, auditável, observável e autônomo o bastante para se descobrir e melhorar a cada ciclo.

> Extraído do ecossistema [Arah](https://github.com/sraphaz/arah). Validado em produção e no monorepo legaltech **[IAutos](https://github.com/sraphaz/iautos)**.

---

## A ideia em uma frase

Instale o harness uma vez. O **TechOrganism** observa domínio e stack, propõe os agentes certos, define como eles se comunicam e evolui com o repositório — **sempre com você no comando do merge**.

```text
  humano define intenção
        │
        ▼
  ┌─────────────────────────────────────────┐
  │           TECHORGANISM                  │
  │  discover → cells → signals → evolve    │
  │  coreografia · skills · gates · audit   │
  └─────────────────────────────────────────┘
        │
        ▼
  Pull Request → CI → ready-for-merge → merge humano
```

---

## Por que existe

Sem um harness, cada repositório novo replica à mão:

- Manifests de agentes e skills  
- Coreografia path-based  
- Scripts de orquestração e gates  
- Hooks, rules, workflows CI  
- Specs SDD, Definition of Done, Agent Graph  

**ARAH versiona isso uma vez.**  
**TechOrganism** torna esse kernel *vivo* no seu repo: percepção, organização e evolução contínuas.

---

## O que torna o ARAH diferente

| Capacidade | Spec Kit | BMAD | autonomous-sdlc | harnessforge | **ARAH** |
|---|:---:|:---:|:---:|:---:|:---:|
| CLI bootstrap | ✅ | ✅ | ✅ | ✅ | ✅ |
| Multi-agente SDLC | ❌ | ✅ | ✅ | ❌ | ✅ |
| Coreografia por paths | ❌ | ❌ | ❌ | ❌ | ✅ |
| Agentes de domínio consultivos | ❌ | ❌ | ❌ | ❌ | ✅ |
| Agent Graph auditável | ❌ | parcial | parcial | ❌ | ✅ |
| Drift-check (`sync-check`) | ❌ | ❌ | parcial | ✅ | ✅ |
| Comunicação passiva (tokens) | parcial | ❌ | parcial | ✅ | ✅ |
| **TechOrganism** (discover → evolve) | ❌ | parcial | parcial | ❌ | ✅ |

### TechOrganism — a dimensão viva

Desde a **v0.3**, o harness não é só um pacote de arquivos. É um **TechOrganism** instalado no repositório:

| Fase | Comando | O que acontece |
|------|---------|----------------|
| Percepção | `discover` | Lê stack, estrutura e pistas de domínio |
| Ontogenia | `organism bootstrap` | Define células, tecidos e vias de sinal |
| Comunicação | `organism signal` | Sinais tipados, append-only, auditáveis |
| Aprendizado | `evolve` | Propõe melhorias a partir do ledger |
| Homeostase | `regenerate` | Atualiza kernel + regenera o organismo |

**Princípio de ouro:** agentes *propõem*; humanos *aplicam*.  
Nada de spawn silencioso. Nada de merge automático. Evolução por seleção via PR.

→ Guia completo: **[docs/TECHORGANISM.md](docs/TECHORGANISM.md)**

---

## Começar em 60 segundos

```powershell
git clone https://github.com/sraphaz/arah-harness.git $env:USERPROFILE\arah-harness
cd C:\caminho\para\meu-projeto

powershell -ExecutionPolicy Bypass -File $env:USERPROFILE\arah-harness\cli\arah.ps1 install `
  -ProjectName "meu-projeto"

# Ative o TechOrganism
powershell -File $env:USERPROFILE\arah-harness\cli\arah.ps1 regenerate -Target . -UpdateKernel
```

Revise as propostas em `docs/_meta/`, ajuste `arah.config.yaml`, abra o PR.

Guia: **[docs/INSTALL.md](docs/INSTALL.md)** · Checklist: **[docs/BOOTSTRAP.md](docs/BOOTSTRAP.md)**

---

## Arquitetura

```mermaid
flowchart TB
  subgraph harness["arah-harness"]
    CLI["cli/arah.ps1"]
    KERNEL["kernel/"]
    TPL["templates/"]
    TO["TechOrganism<br/>discover · organism · evolve"]
  end
  subgraph target["seu repositório"]
    CFG["arah.config.yaml"]
    AGENTS[".agents/ + .skills/"]
    META["docs/_meta/<br/>discovery · organism · evolution"]
    BUS[".arah/local/bus/"]
    OVERLAY["choreography.*.yaml"]
  end
  CLI -->|install / update / regenerate| KERNEL
  CLI --> TO
  KERNEL --> AGENTS
  TPL --> CFG
  TO --> META
  TO --> BUS
  CFG -->|domain sync| OVERLAY
  META -->|propostas → PR| CFG
```

### Camadas

| Camada | O quê | Quem mexe |
|--------|-------|-----------|
| Produto | Código da aplicação | Agentes operacionais + humanos |
| Domínio | Pareceres de negócio | Agentes consultivos (gerados) |
| Kernel | Operacionais, skills, scripts, hooks | `arah update` / `regenerate` |
| Config | `arah.config.yaml`, overlays, AGENTS.md | Humano (+ Apply revisável) |
| TechOrganism | Manifesto, sinais, evolução | Organismo + PR |

---

## CLI

### Essencial

| Comando | Função |
|---------|--------|
| `install` | Bootstrap completo (recomendado) |
| `init` | Kernel + templates + CI |
| `update [-Force]` | Reaplica kernel (preserva config) |
| `doctor` | Valida instalação |
| `sync-check` | Drift vs upstream (CI) |
| `domain sync` | Gera agentes de domínio |
| `export-graph` | Agent Graph (JSON + Mermaid) |

### TechOrganism

| Comando | Função |
|---------|--------|
| `discover [-Apply]` | Observa repo → propostas de domínio/stack |
| `organism bootstrap` | Ritual do primeiro momento |
| `organism status` | Estado do organismo |
| `organism signal` | Emite sinal tipado no bus |
| `evolve [-Apply]` | Ciclo de self-learning |
| `metrics rollup\|report` | Economy Intelligence (scorecard) |
| `task create\|status\|validate\|complete\|block` | Execution Control Protocol |
| `update-check` | Notifica se `.arah-version` &lt; latest Release |
| `regenerate [-UpdateKernel]` | Homeostase completa no consumidor |

```powershell
# Ativar / atualizar TechOrganism em um consumidor
powershell -File cli/arah.ps1 regenerate -Target C:\meu-projeto -UpdateKernel -Force
```

---

## Estrutura do repositório

```
arah-harness/
├── kernel/                 # Superfície distribuída (versionada)
│   ├── .agents/            # Operacionais + domain advisors + coreografia
│   ├── .skills/            # Procedimentos determinísticos
│   ├── .cursor/            # Hooks (domain review + live session)
│   └── scripts/            # Orquestração, gates, TechOrganism, telemetria
├── cli/                    # install · discover · organism · evolve · metrics · regenerate
├── extension/arah-live/    # Painel Cursor/VS Code em tempo real
├── harness/profiles/       # Tiers: minimal → enterprise
├── schemas/arah-harness/   # Contratos canônicos
├── templates/              # Config, AGENTS.md, CI
├── docs/                   # Método, TechOrganism, Economy, governança
└── scripts/self-test.ps1
```

---

## Princípios

1. **Humano comanda** — merge sempre humano  
2. **Tudo via Pull Request**  
3. **Escopo mínimo** por manifest  
4. **Spec-before-code** quando a fase exige  
5. **Contexto sob demanda** — arquivo + CI + sinais tipados  
6. **Agentes propõem; humanos aplicam** — autonomia com ledger  
7. **Kernel imutável no consumidor** — customização em config e overlays  

---

## Documentação

| Documento | Conteúdo |
|-----------|----------|
| **[TECHORGANISM.md](docs/TECHORGANISM.md)** | Dimensão viva — discovery, organismo, sinais, evolução |
| **[ECONOMY.md](docs/ECONOMY.md)** | Economy Intelligence — scorecard e eficiência |
| **[EXECUTION_CONTROL.md](docs/EXECUTION_CONTROL.md)** | Terminalidade — um executor, done\|blocked |
| **[UPDATE_NOTIFICATIONS.md](docs/UPDATE_NOTIFICATIONS.md)** | Releases + cron → issue nos consumidores |
| [STATE_MODEL.md](docs/STATE_MODEL.md) | Estado quente × evidência fria |
| [backlog/](docs/backlog/) | Control Plane — épicos W/C/H |
| [website/](website/) | Site EN/PT + docs + Live Console (UI) → [GitHub Pages](https://sraphaz.github.io/arah-harness/) |
| [live/](live/) | Live Service Go (`arah-live`) — API read-only para o console |
| [cmd/arah/](cmd/arah/) | CLI portátil Go fase 1 (`doctor` / `sync-check`) |
| [METHOD.md](docs/METHOD.md) | Método ARAH completo |
| [INSTALL.md](docs/INSTALL.md) | Instalar em qualquer repo |
| [BOOTSTRAP.md](docs/BOOTSTRAP.md) | Checklist pós-install |
| [GOVERNANCE.md](docs/GOVERNANCE.md) | Autonomia e gates humanos |
| [MODEL.md](docs/MODEL.md) | Harness-model first-class |
| [LIVE_SESSION.md](docs/LIVE_SESSION.md) | Telemetria + extensão |
| [MARKET_REFERENCE.md](docs/MARKET_REFERENCE.md) | Posicionamento vs mercado |
| [CHANGELOG.md](CHANGELOG.md) | Histórico de versões |

---

## Exemplo real

**[IAutos](https://github.com/sraphaz/iautos)** — monorepo legaltech white-label:

- Domínios consultivos (`core-cases`, `compliance`, `auth-tenant`, …)
- Overlay `choreography.iautos.yaml` para monorepo
- Separação clara: SDLC ARAH no repo vs agentes runtime do produto

---

## Desenvolvimento deste repo

```powershell
./scripts/self-test.ps1
```

Contribuindo: [CONTRIBUTING.md](CONTRIBUTING.md)

---

## Licença

[MIT](LICENSE) — Copyright (c) 2026 Raphael / Arah contributors

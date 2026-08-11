# ADR-003 — Fronteira Kernel × Control Plane: `org-control-plane` sai do escopo

- **Status:** Proposed — **decisão arquitetural; requer aprovação humana explícita (merge é o ato de aceite)**
- **Date:** 2026-08-11
- **Deciders:** maintainers (proposta agente; merge humano)
- **Supersedes / relates:** ADR-002 (Runtime Cohesion) — já listava "control plane multi-repo" como deliberadamente fora do 0.5; este ADR torna a exclusão **permanente**, não apenas um adiamento

## Context

O ARAH Harness é o **Engineering Kernel** do portfólio: define como se desenvolve em
**UM** repositório (agentes, gates, Execution Control, evidência, telemetria local).
Ao mesmo tempo, o repositório acumulou sinais de uma segunda ambição — virar
agregador org-level:

- `capabilities.yaml` lista `org-control-plane` ("Control plane multi-repo") em `planned`;
- `docs/ROADMAP.md` menciona "Control plane multi-repo (org)" no horizonte Later
  e usa "control plane de repositórios" como norte;
- o website anuncia "multi-repository control" e "org-wide governance" como planned
  (`website/content/home/{en,pt}.json`);
- há protótipos de UI em `docs/design/control-plane/` e um backlog inteiro
  (`docs/backlog/`, épicos C/W) nomeado "ARAH Control Plane".

A análise transversal do ecossistema (auditoria 2026-08-10, `overlaps-and-contracts.md`
§D1, e plano de evolução do kernel, doc J §2) identificou aqui o **maior risco de
duplicação de responsabilidade do portfólio**: o **Surya Labs Workspace já é o
Control Plane operacional** — portfólio (`data/portfolio/`), Decision Center,
task router, health score derivado de probe do harness. Se o item `org-control-plane`
fosse executado no kernel, os dois ativos colidiriam frontalmente na mesma
responsabilidade (agregação multi-repo, governança de portfólio), com dois bancos,
duas UIs e duas trilhas de decisão para o mesmo domínio.

A fronteira **de fato** hoje já é saudável: harness = intra-repo; Workspace = inter-repo.
Falta registrá-la como decisão para impedir regressão por inércia de backlog.

## Options

| Opção | Coesão do ecossistema | Risco | Valor |
|-------|----------------------|-------|-------|
| **A — Kernel single-repo por contrato** — o harness expõe dados de um repo; o Workspace agrega | Alta | Baixo | Alto — cada plano com uma responsabilidade |
| B — Manter `org-control-plane` planned no kernel | Baixa | Alto (duplicação do Workspace, dois control planes) | Baixo |
| C — Kernel absorve o Workspace | Média | Muito alto (reescrita de portfólio/decisões dentro de um kernel instalável) | Baixo |

## Decision

Adotar a **opção A**. O ARAH Harness **nunca agrega múltiplos repositórios**.

1. **O kernel expõe; o Workspace agrega.** O ARAH Harness publica dados de UM repo —
   `doctor`, Evidence Graph, timeline, ExecutionEvidence, OTLP (futuro) — em contratos
   versionados (`schemas/contracts/`, diretório introduzido pelo pacote de contratos da
   federação — PR #29). Qualquer visão de portfólio, agregação org-level,
   governança multi-repo e roteamento de demandas é responsabilidade do
   **Surya Labs Workspace** (Control Plane do portfólio).
2. **`org-control-plane` é removido** de `capabilities.yaml` (planned) e do
   `docs/ROADMAP.md`. "org-wide governance" e "multi-repository control" saem da
   narrativa do website.
3. **O console `live/` permanece single-repo** — visualização local read-only do
   próprio repositório. AuthZ (C-10) segue no backlog do console.
4. **Protótipos e épicos que assumem agregação** (`docs/design/control-plane/`,
   partes dos épicos C/W do backlog) ficam **arquivados como referência histórica**;
   se alguma feature deles voltar, volta como feature do Workspace, não do kernel.
5. Nome do backlog/handoff "ARAH Control Plane" passa a ser lido como legado
   histórico do épico de site+console — não como direção de produto do kernel.

## Consequences

### Positivas

- Elimina o maior risco de colisão de responsabilidade do portfólio (auditoria §D1).
- Cada capability tem um dono único: shaping (Sky Forge) → portfólio (Workspace) →
  engenharia intra-repo (ARAH Harness).
- O kernel fica menor, instalável e testável — coerente com ADR-002.
- A integração vira contrato explícito (ExecutionRequest/ExecutionEvidence/
  CanonicalEvent/scorecards) em vez de UI/banco compartilhado.

### Negativas / trade-offs

- Ideias de UI org-level já prototipadas em `docs/design/control-plane/` não serão
  construídas aqui; migrá-las ao Workspace exige retrabalho de design.
- O website perde dois itens de narrativa "planned" (honestidade > promessa).

### Não-consequências (explícito)

- Não remove o console `live/` nem a extensão ARAH Live (ambos single-repo).
- Não afeta o Economy Intelligence por repo — o rollup continua; a **agregação**
  dos scorecards entre repos é do Workspace.
- Não impede que o Workspace consuma dados do kernel — é exatamente o desenho.

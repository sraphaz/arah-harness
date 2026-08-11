# Backlog executável — ARAH Harness

Três épicos paralelizáveis: **W** (website+docs), **C** (Live Console), **H** (melhorias do harness). Issues no formato copiável para GitHub Issues. Ordem dentro de cada épico = ordem de execução recomendada; dependências explícitas quando cruzam épicos.

Legenda de estimativa: P (≤½ dia) · M (1–2 dias) · G (3–5 dias).

---

## Épico W — Site de produto + portal de docs

**Meta:** publicar o site bilíngue com docs em MDX mantidas no repositório. Stack: Next.js (App Router) + TypeScript + Tailwind + MDX. Referências: `design-files/*.dc.html`.

### W-01 · Scaffold do site `M`
Next.js + TS + Tailwind + MDX (contentlayer ou next-mdx-remote). ESLint/Prettier. Deploy preview (Vercel ou Pages).
- [ ] `pnpm dev` roda; página exemplo com fontes Schibsted Grotesk + IBM Plex Mono via `next/font`
- [ ] Tokens do README (cores/raios/espaços) em `tailwind.config.ts` (`arah.bg`, `arah.surface`, `arah.line`, `arah.accent`…)
- [ ] CI: lint + typecheck + build em PR

### W-02 · Layout global: Nav, Footer, seletor de idioma `M`
- [ ] Nav sticky com blur, wrap em telas estreitas, logo SVG do grafo, CTAs GitHub/Get Started
- [ ] Seletor EN|PT liga rotas `/en/*` ↔ `/pt/*` (mesma página, outro locale)
- [ ] Footer com tagline + GitHub/Releases/Changelog/Contributing/License
- [ ] Acessível por teclado; foco visível; contraste AA nos tons `#9AA5B1`+

### W-03 · i18n `M`
Dicionários por página (JSON/TS) para EN e PT-BR — textos integrais já estão nos `.dc.html`.
- [ ] `/en` e `/pt` estáticos (SSG), `hreflang` correto
- [ ] Código/comandos/paths nunca traduzidos
- [ ] 404 localizada

### W-04 · Home §1–4: hero animado, shift, problema, inversão `G`
- [ ] Pipeline de 7 estágios com `flowDot` + `stagePulse` sincronizados (CSS puro; respeitar `prefers-reduced-motion`)
- [ ] Grades responsivas idênticas ao design (auto-fit/minmax)
- [ ] Lighthouse: sem CLS das animações

### W-05 · Home §5: Harness Explorer (13 componentes) `M`
Dados extraídos do array `parts` (Home PT) para `content/harness-parts.{en,pt}.json`.
- [ ] Lista + painel sticky com FUNÇÃO/MATURIDADE/EXEMPLO
- [ ] Navegável por teclado (roving tabindex ou radiogroup)
- [ ] Deep-link `?part=slug`

### W-06 · Home §6–9: camadas, princípios, lifecycle, capability map `M`
- [ ] Camadas e capability map interativos com estado em URL-hash
- [ ] Chips do lifecycle linkam para as etapas do How It Works

### W-07 · Home §10–14: matriz, not-list, status, quick start, CTA `M`
- [ ] Matriz com scroll horizontal e cabeçalho sticky; coluna ARAH destacada
- [ ] Botões Copy com clipboard + feedback 1.5s (componente `CopyButton` reutilizável)
- [ ] Painéis Available/Experimental/Planned gerados de `content/status.json` (fonte única — ver H-10)

### W-08 · Página Architecture `M`
- [ ] Diagrama de camadas + duas árvores de diretórios (componente `FileTree` com comentários)
- [ ] 6 cards de contratos + fluxo de update em 4 passos

### W-09 · Página How It Works `G`
- [ ] Timeline de 12 etapas (dados em `content/lifecycle.{en,pt}.json`), etapa na URL
- [ ] Walkthrough de 9 passos com Voltar/Avançar, dots, teclas ←/→, passo na URL
- [ ] Terminal pane com `<pre>` estilo `#080A0D`

### W-10 · Páginas TechOrganism e Use Cases `M`
- [ ] Ciclo com `cyclePulse` em cadeia; grade de 11 correspondências; not-claims
- [ ] 8 cards de use cases

### W-11 · Portal de docs (MDX) `G`
Converter os arrays `docs` (EN e PT) em `content/docs/{en,pt}/<section>/<slug>.mdx` com frontmatter (title, section, order).
- [ ] Sidebar gerada do filesystem; item ativo com borda ciano; colapsa em drawer no mobile (resolve pendência do protótipo)
- [ ] Breadcrumb, prev/next, "Edit this page on GitHub" (link para o path do MDX), badge de versão documentada
- [ ] Headings com anchor links; TOC à direita em ≥1280px
- [ ] Blocos de código com syntax highlighting (shiki) + CopyButton
- [ ] URLs permanentes `/docs/<section>/<slug>`

### W-12 · Busca das docs `M`
- [ ] Índice local (FlexSearch) construído no build sobre título+headings+corpo dos MDX
- [ ] Atalho `/` foca a busca; resultados com seção; estado vazio
- [ ] Sem dependência de serviço externo

### W-13 · CLI Reference interativa `M`
Dados dos 14 comandos em `content/cli.{en,pt}.json` (extrair do array `cli`).
- [ ] Lista + painel (sintaxe, LÊ/ESCREVE, exemplo/resultado); comando na URL
- [ ] Validar conteúdo contra `cli/arah.ps1` real antes de publicar (par com H-01)

### W-14 · SEO, OG e qualidade `M`
- [ ] Metadata por página, OG images (template dark com título), sitemap.xml, robots
- [ ] 404 com navegação; estados vazios
- [ ] Lighthouse ≥90 em performance/a11y/SEO; navegação 100% por teclado
- [ ] axe sem violações críticas

### W-15 · Conteúdo: calibrar mocks com o repo real `M`
- [ ] Substituir `v1.5.0`, contagens (14 agentes/22 skills), exemplos de manifests e outputs pelos reais de `sraphaz/arah-harness`
- [ ] Revisão técnica dos textos por mantenedor
- Depende de: acesso ao repo; bloqueia release público

---

## Épico C — ARAH Live Console

**Meta:** MVP do console read-only plugável. Referência: `design-files/ARAH Live Console.dc.html` (UI + C4). Decisões travadas: read-only; artefatos como contrato; índice SQLite descartável; 3 implantações (extensão IDE → app local → serviço org — MVP = app local).

### C-01 · Especificar contrato de dados dos artefatos `M`
Documentar schemas do que o console lê: eventos do bus (`.arah/bus/*.jsonl`), ledger (`.arah/audit/`), `docs/_meta/{domains.yaml,graph.json,discovery.proposed.yaml}`, `arah.config.yaml`.
- [ ] JSON Schemas versionados em `schemas/console/` no repo do harness
- [ ] Tipos de evento e prefixos (consultation.*, gates.*, change.*, evolution.*, session.*) com payloads mínimos
- Depende de: H-02 (tipos de sinal estáveis). Bloqueia C-03+.

### C-02 · Scaffold do Live Service `M`
Serviço local (Node/TS ou Go — decidir com H-06) com REST + WebSocket, config por flag `--repo <path>`.
- [ ] `GET /api/summary`, `/api/feed?filter=`, `/api/gates`, `/api/domains`, `/api/queue`, `/api/proposals`
- [ ] `WS /events` com eventos tipados
- [ ] Sem endpoint de escrita — lint de rota garante (Decisão 01)

### C-03 · FS Watcher + Bus Reader `G`
- [ ] Watch com debounce em `.arah/bus` e `.arah/audit`; parse incremental (offset por arquivo)
- [ ] Validação contra schemas C-01; eventos malformados vão para fila de erro visível, sem derrubar o stream
- [ ] Testes com fixtures de JSONL reais

### C-04 · Ledger Indexer (SQLite derivado) `G`
- [ ] Projeção dos eventos em SQLite local (tabelas: events, runs, gates, consultations)
- [ ] Regenerável do zero (`--reindex`) — apagar o .db nunca perde dados
- [ ] Consultas dos KPIs (<50ms em ledger de 100k eventos)

### C-05 · Graph Builder + Gate Monitor `M`
- [ ] Carrega `graph.json` + detecta drift (hash das fontes vs grafo)
- [ ] Agrega última execução de gates + taxa de aprovação 24h/30d

### C-06 · GitHub Adapter (fila de seleção) `M`
- [ ] Poll/webhook de PRs abertos com evidência vinculada (`run-*`) e status de checks
- [ ] Token só com escopo read; ações são links para o PR no GitHub

### C-07 · Console Web — shell + KPIs + seletor de repo `M`
Reusar tokens/componentes do site (W-01/W-02 como pacote compartilhado se possível).
- [ ] Header com repo picker, chips kernel/drift, indicador ao-vivo
- [ ] 5 KPIs alimentados por `/api/summary`, atualizando via WS

### C-08 · Console Web — Signal Feed + Gate Panel `M`
- [ ] Feed com filtros (todos/consulta/gates/mudanças/evolução), cores por prefixo, autoscroll pausável
- [ ] Painel de gates ✓/✗ com duração e sumário

### C-09 · Console Web — Territory Map + Selection Queue + Proposals `M`
- [ ] Territórios com saúde/autonomia; fila com estado vazio; propostas com evidência
- [ ] Barras de distribuição de autonomia

### C-10 · AuthZ + empacotamento app local `M`
- [ ] OAuth GitHub (escopo espelha acesso ao repo) quando exposto além de localhost; sem auth em localhost puro
- [ ] Distribuição: binário único ou `npx arah-live`
- [ ] README de operação

### C-11 · (Fase 2) Extensão de IDE `G`
- [ ] Webview VS Code/Cursor embarcando o Console Web apontado ao workspace
- Depende de: C-07–C-09 estáveis

### C-12 · (Fase 2) Graph Explorer `G`
- [ ] Visualização navegável do grafo (agentes-skills-domínios) — único painel do design ainda não prototipado em UI

---

## Épico H — Melhorias do harness (da Análise Técnica)

**Meta:** atacar o overhead de escrita em arquivos e os riscos ALTA antes de escalar. Referência: `design-files/Analise Tecnica.dc.html` §3–6.

### H-01 · Auditoria de superfície da CLI `P`
Confirmar comandos/flags reais vs documentados (14 comandos do site).
- [ ] Tabela de gaps; corrigir site (W-13/W-15) ou CLI
- Bloqueia: W-13, W-15

### H-02 · Congelar e versionar tipos de sinal `M`
- [ ] Enum + JSON Schema por tipo de evento; campo `v` no payload
- [ ] Documento de compatibilidade (aditivo apenas)
- Bloqueia: C-01

### H-03 · Separação estado quente × evidência fria `G`
- [ ] `.arah/local/` (gitignored): telemetria e sinais operacionais
- [ ] Versionado: apenas resumo compacto por run (`run-*/summary.json`) + eventos de decisão
- [ ] Migração: comando `arah migrate-state` move o histórico atual
- [ ] Docs atualizadas (W-11 ganha página "State model")

### H-04 · Arquivo-por-evento + compactação `G`
- [ ] Escrita = criar arquivo `<ULID>.json` (atômico, sem lock, sem conflito de merge)
- [ ] `arah compact` funde eventos antigos em JSONL por período; retenção configurável
- [ ] Benchmark antes/depois em cenário multiagente (documentar no PR)

### H-05 · Scrubbing de secrets na evidência `M` **(risco ALTA)**
- [ ] Redação por regex/entropia antes de persistir qualquer payload
- [ ] Gate de segurança roda também sobre `.arah/**`
- [ ] Teste: secret plantado nunca chega ao disco

### H-06 · Decisão: linguagem da CLI portátil `P`
ADR Go vs Rust vs Node SEA; critérios: startup, distribuição single-binary, reuso no Live Service (C-02).
- [ ] ADR aprovado antes de H-07/C-02 avançarem juntos

### H-07 · CLI binária portátil (fase 1) `G`
- [ ] Reimplementar leitura/validação (doctor, sync-check, export-graph) na linguagem do ADR; PowerShell permanece para o resto durante transição
- [ ] Paridade de exit codes (0/1/2/3/4/10)

### H-08 · Daemon `arahd` (opcional) `G`
- [ ] Watch + batch de escrita (debounce/flush) + stream WS nativo
- [ ] CLI degrada para escrita direta sem daemon; console usa o stream quando presente (substitui C-03 watch)

### H-09 · Pre-commit hooks + branch protection guide `P`
- [ ] `arah install` oferece hooks; docs de enforcement (mitiga "cooperação apenas")

### H-10 · Fonte única de status de capacidades `P`
- [ ] `capabilities.yaml` no repo (available/experimental/planned) consumido pelo site (W-07) — status nunca diverge da realidade

### H-11 · Modo mínimo de adoção `M`
- [ ] `arah install --minimal`: só manifests + gates (sem organismo/bus) e upgrade path documentado

---

## Sequenciamento sugerido (6 sprints ilustrativos)

- **S1:** W-01→W-03 · H-01, H-02, H-06 · C-01
- **S2:** W-04→W-06 · C-02, C-03 · H-05
- **S3:** W-07→W-10 · C-04, C-05 · H-03
- **S4:** W-11, W-12 · C-06→C-08 · H-04
- **S5:** W-13→W-15 · C-09, C-10 · H-09, H-10
- **S6:** lançamento site+console MVP · H-07, H-11 · fase 2 (C-11, C-12, H-08)

**Definition of Done geral:** revisão humana em PR; gates verdes; docs atualizadas no mesmo PR; textos EN+PT quando tocar o site.

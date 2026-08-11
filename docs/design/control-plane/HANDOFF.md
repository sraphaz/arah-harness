# Handoff: ARAH Harness — Site de produto, portal de docs e Live Console

## Overview

Este pacote entrega três frentes para implementação:

1. **Site de produto + portal de documentação** (bilíngue EN/PT-BR) do ARAH Harness — posicionamento, arquitetura, funcionamento, TechOrganism, casos de uso e docs navegáveis com busca e CLI explorer.
2. **ARAH Live Console** — frontend de observabilidade read-only para repositórios governados pelo harness (feed de sinais, gates, territórios, fila de seleção humana, propostas de evolução), com arquitetura C4 já definida.
3. **Melhorias no harness** (backend/CLI) derivadas da análise técnica: separação estado quente × evidência fria, arquivo-por-evento, índice derivado, scrubbing de secrets etc.

O backlog executável completo está em `BACKLOG.md`. Este README traz o contexto de design e os contratos que o backlog referencia.

## About the Design Files

Os arquivos em `design-files/` são **referências de design criadas em HTML** (protótipos navegáveis com dados simulados), não código de produção para copiar diretamente. A tarefa é **recriar essas telas no ambiente do projeto real**. O brief do produto pede: **Next.js + TypeScript + Tailwind CSS, docs em MDX mantidas no próprio repositório** — use isso como alvo, a menos que o repositório real já tenha outra stack estabelecida.

Cada `*.dc.html` abre direto no navegador (com `support.js` na mesma pasta). A lógica interativa de cada página vive num bloco `<script data-dc-script>` no fim do arquivo — os dados mock (docs, comandos da CLI, eventos do feed) estão lá como arrays JS legíveis, prontos para extrair para JSON/MDX.

## Fidelity

**Alta fidelidade (hifi).** Cores, tipografia, espaçamentos, estados de hover e microinterações são finais e devem ser reproduzidos fielmente. O que é simulado: todos os dados do Live Console (3 repositórios fake), o número de versão `v1.5.0`, contagens (14 agentes, 22 skills), e os exemplos de manifests/outputs — calibrar tudo contra o repositório real `sraphaz/arah-harness`.

## Design Tokens

Cores (dark, todas as páginas exceto Análise Técnica):

- Fundo base: `#0A0C0F` · fundo alternado de seção: `#0B0E12`
- Superfície de card: `#0D1117` · superfície de código/terminal: `#080A0D` · painel profundo: `#0C0F14`
- Bordas: `#1E2630` (cards) · `#1A212B` (painéis) · `#141A22`/`#161C24` (hairlines de seção/nav) · `#232B36` (controles) · `#2A3340` (chips)
- Texto: primário `#E7ECF1` · secundário `#B7C0CB` · terciário `#9AA5B1`/`#8C97A5` · apagado `#5A6675` · mínimo `#4A5563`
- Acento ciano (marca/ação): `oklch(78% 0.09 200)` — hover `oklch(84% 0.09 200)`, texto sobre acento `#0A0C0F`, borda translúcida `oklch(75% 0.09 200 / .35–.5)`, fundo tinto `oklch(75% 0.09 200 / .05–.12)`, texto ciano claro `oklch(84% 0.07 200)`
- Acento âmbar (etapa humana, sempre): `oklch(78% 0.09 80)` / texto `oklch(84% 0.07 80)` / borda `oklch(75% 0.09 80 / .3–.5)`
- Verde (ok/available): `oklch(75% 0.1 160)` · Vermelho (falha/problema): `oklch(72% 0.11 30)`
- Documento impresso (Análise Técnica): texto `#1A2129`, secundário `#4A5560`/`#5B6570`, bordas `#D8DEE4`, links `#0E7490`

Tipografia:

- Display/corpo: **Schibsted Grotesk** (400–700). Hero: `clamp(34px,5.5vw,56px)/1.08/-0.02em`; H1 de página: `clamp(32px,5vw,48px)`; H2 de seção: `clamp(28px,4.2vw,38px)/1.15/-0.015em`; corpo 16.5px/1.65; cards 13.5–15.5px
- Técnico/labels: **IBM Plex Mono** (400–600). Eyebrows: 12–13px, `letter-spacing:.14em`, uppercase; labels de card: 10.5–11px `.12em`; código: 12.5–13px/1.7
- Semântica consistente: mono = comandos, paths, labels, números; grotesk = prosa e títulos

Espaçamento e forma:

- Container: `max-width:1200px` (site) / `1360px` (console e docs), padding lateral 32px
- Seções: padding vertical 90–100px, alternando fundo base/`#0B0E12`, separadas por `border-top:1px solid #141A22`
- Raios: 6–8px (botões/controles), 9–10px (cards), 12–14px (painéis) · sem sombras — profundidade por borda + fundo
- Grids: sempre `repeat(auto-fit,minmax(Npx,1fr))` com `gap` 8–16px
- Nav sticky: `min-height:64px`, fundo `rgba(10,12,15,.82)` + `backdrop-filter:blur(14px)`, com `flex-wrap`

Animações (discretas, sem ornamento):

- `flowDot`: ponto percorrendo a linha do pipeline do hero, 7s linear infinito
- `stagePulse`: borda dos estágios pulsando em cascata (delays 0–6.2s), sincronizada com o flowDot
- `cyclePulse`: opacidade .4→1 nos nós do ciclo TechOrganism, 7s, delays de 1s em cadeia
- `livePulse`: dot verde "ao vivo" do console, 2s
- `fadeUp`: entrada do hero, .6s ease, delay .15s na coluna direita

## Screens / Views

Todas as páginas compartilham: nav sticky (logo SVG de grafo + links + seletor EN|PT + "View on GitHub" outline + "Get Started" preenchido ciano), footer com tagline e links GitHub/Releases/Contributing. Páginas EN e PT são espelhos exatos — implementar como uma única base com i18n (o brief pede arquitetura preparada para i18n; os textos completos dos dois idiomas estão nos arquivos).

### 1. Home (`Home.dc.html` / `Home PT.dc.html`)

Seções, em ordem:

1. **Hero** — 2 colunas (`minmax(360px,1fr)`): título + subtítulo + 3 CTAs; à direita, painel "One governed change" com pipeline animado de 7 estágios (Human Intent → Repository Context → ARAH Harness → Agents+Skills+Policies → Gates+Evidence → Pull Request → Human Selection). Estágios 03 e 07 destacados (ciano/âmbar). Linha vertical com dot animado.
2. **The shift** — texto + blockquote com borda ciano à esquerda; grade 2×4 de chips (COORDINATE/GOVERN/PRESERVE + item).
3. **The problem** — 8 cards P·01–P·08 (label mono vermelho, título 15.5px, corpo 13.5px).
4. **The inversion** — 2 painéis: fluxo agent-first (4 chips mono) vs repository-first (8 chips, Agents/Skills em ciano, Pull request em âmbar; painel com borda ciano e gradiente tinto).
5. **Harness explorer** (interativo) — lista de 13 componentes à esquerda (botões com número, nome, estágio); painel sticky à direita com stage+arquivo, definição, grid FUNÇÃO/MATURIDADE, e bloco EXEMPLO em `<pre>`. Estado: índice selecionado.
6. **Layered architecture** (interativo) — 5 camadas clicáveis (C1–C5) + painel sticky de detalhe.
7. **Principles** — 8 cards estáticos (título ciano claro + corpo).
8. **Lifecycle strip** — 12 chips numerados (11 "Human Review" em âmbar) + CTA para How It Works.
9. **Capability map** (interativo) — 8 botões de área + painel com descrição e chips de capacidades.
10. **Positioning matrix** — tabela 14 dimensões × 6 categorias (● ◐ —), coluna ARAH destacada com fundo tinto; scroll horizontal em `min-width:960px`.
11. **What ARAH is not** — 9 chips com ✕ vermelho + blockquote âmbar.
12. **Capability status** — 3 painéis Available (verde) / Experimental (âmbar) / Planned (cinza) com chips.
13. **Quick start** — coluna esquerda com loop de 5 comandos; direita com 3 blocos de código copiáveis (header com label + botão Copy → "Copied ✓" por 1.5s via `navigator.clipboard`).
14. **CTA final** — "Agents propose. Humans select. The repository remembers." + 3 botões.

### 2. Architecture (`Architecture.dc.html` + PT)

Diagrama de camadas empilhadas ("autoridade desce, evidência sobe", ↓↑ entre camadas, ARAH destacado); duas árvores de diretórios lado a lado (produto `arah-harness/` e consumidor com comentários `#`); 6 cards de contratos/armazenamento (YAML+schema, bus JSONL, ledger, grafo derivado, IDE, CI); fluxo de atualização em 4 passos (último âmbar = PR humano).

### 3. How It Works (`How It Works.dc.html` + PT)

**Timeline interativa**: 12 chips de etapa → painel com descrição, ENTRADA/SAÍDA, COMANDO e ARTEFATOS GERADOS (dados completos no script do arquivo). **Walkthrough de 9 passos** do cenário `backend/payments/**`: dots 1–9 + botões Voltar/Avançar (com wrap), split view esquerda (fase, título, corpo) / direita (terminal `#080A0D` com pre). Conteúdo integral nos arrays `stepsData` e `demoData`.

### 4. TechOrganism (`TechOrganism.dc.html` + PT)

Ciclo de 7 nós com `cyclePulse` em cadeia (Select em âmbar) + legenda "Selection is the human step"; grade de 11 correspondências biologia→mecanismo (Selection com borda âmbar); 5 chips ✕ do que o modelo NÃO afirma; CTA.

### 5. Use Cases (`Use Cases.dc.html` + PT)

8 cards UC·01–UC·08 em grid `minmax(340px,1fr)` + CTA. Estático.

### 6. Docs (`Docs.dc.html` + PT)

Layout `264px minmax(0,1fr)`: sidebar sticky com busca (filtra títulos+seções, estado vazio "No pages match") e 8 seções de navegação (item ativo: fundo `#10151C` + borda esquerda ciano 2px); main com breadcrumb + "Edit this page on GitHub", título, intro, blocos (h2 / parágrafo / código com Copy / lista com marcador ciano), prev/next no rodapé. **CLI Reference** é página especial: lista de 14 comandos à esquerda + painel sticky (sintaxe em pre ciano, descrição, LÊ/ESCREVE, EXEMPLO·RESULTADO). Badge `v1.5.0` na nav. Todo o conteúdo (8 seções, ~30 páginas, 14 comandos) está nos arrays `docs` e `cli` do script — extrair para MDX/JSON.

### 7. ARAH Live Console (`ARAH Live Console.dc.html`)

Console dark de observabilidade: nav própria (badge EXPERIMENTAL); seletor de 3 repositórios (dot verde/âmbar por drift); chips kernel/drift/ao-vivo (livePulse); 5 KPIs; grid de painéis — Feed de sinais (filtros todos/consulta/gates/mudanças/evolução; linhas hora + tipo colorido por prefixo + mensagem + rota; `grid-row:span 2`), Gates última execução (✓/✗ + duração + sumário), Territórios (nome, path, saúde, agentes, sinais/24h, autonomia), Fila de seleção humana (borda âmbar; PRs com gates/autonomia/evidência/agente; estado vazio), Evolução+Autonomia (propostas com evidência; 3 barras de distribuição). Cores por tipo de evento: consultation=ciano, gates.passed=verde, gates.failed=vermelho, evolution=`#B49BE0`, demais=âmbar.

Abaixo, seção **Arquitetura C4** (referência para o time — pode virar docs): Contexto (3 pessoas, sistema em foco, 3 sistemas externos), Containers (Console Web SPA, Live Service, Índice SQLite derivado + 3 externos incl. futuro `arahd`), Componentes (8 do service: FS Watcher, Bus Reader, Ledger Indexer, Graph Builder, Gate Monitor, GitHub Adapter, WS Broadcaster, AuthZ; 6 do web: Signal Feed, Gate Panel, Territory Map, Selection Queue, Proposal Review, Graph Explorer) e 4 decisões (read-only por padrão; artefatos como contrato de dados; índice descartável; 3 formas de implantação: extensão IDE / app local / serviço org).

### 8. Análise Técnica (`Analise Tecnica.dc.html`)

Documento impresso (Letter, claro) — fonte para issues do backlog de melhorias do harness, não uma tela do site. Se quiser publicá-lo, renderizar como página de docs.

## Interactions & Behavior

- Toda interatividade é client-side por estado local (índice selecionado, filtro, página atual). Sem fetch no site; o console real consome REST+WS do Live Service.
- Botões Copy: `navigator.clipboard.writeText`, label → "Copied ✓"/"Copiado ✓" por 1500ms (um timer por página; guardar id do bloco copiado).
- Hovers: bordas `#1E2630→#3A4553`, links `#9AA5B1→#E7ECF1`, botão primário clareia o oklch. Transições padrão do navegador (sem transition custom).
- Walkthrough: Avançar/Voltar com wrap-around módulo 9.
- Busca das docs: filtro case-insensitive em título+seção; qualquer query não-vazia troca a sidebar por resultados; clicar num resultado navega e limpa a query.
- Responsivo por `auto-fit/minmax` + `clamp()` em títulos + nav com wrap. Sem breakpoints JS. Pendência conhecida: sidebar das docs não colapsa em hambúrguer no mobile (item do backlog).

## State Management (site real)

- Docs: rota por página (`/docs/<section>/<slug>`) com prev/next derivados da ordem do sitemap; busca local (FlexSearch/fuse) client-side.
- Console: estado servidor via WebSocket (eventos tipados) + REST para snapshots; seleção de repo/filtros na URL.
- i18n: `/en` e `/pt` com dicionários por página; comandos, paths e blocos de código nunca são traduzidos.

## Assets

Sem imagens externas. Logo = SVG inline de grafo (3 linhas + 4 círculos, nó central ciano) — está em todas as navs. Fontes via Google Fonts (Schibsted Grotesk 400–700, IBM Plex Mono 400–600).

## Files

- `design-files/Home.dc.html` + `Home PT.dc.html` — homepage EN/PT
- `design-files/Architecture.dc.html` + PT — arquitetura
- `design-files/How It Works.dc.html` + PT — timeline + walkthrough
- `design-files/TechOrganism.dc.html` + PT — modelo operacional
- `design-files/Use Cases.dc.html` + PT — casos de uso
- `design-files/Docs.dc.html` + PT — portal de docs + CLI explorer (conteúdo nos arrays do script)
- `design-files/ARAH Live Console.dc.html` — console + arquitetura C4
- `design-files/Analise Tecnica.dc.html` — análise técnica (imprimível)
- `design-files/support.js`, `design-files/doc-page.js` — runtime dos protótipos (apenas para abrir os .dc.html localmente; não portar)
- `BACKLOG.md` — backlog executável (épicos → issues com critérios de aceite)

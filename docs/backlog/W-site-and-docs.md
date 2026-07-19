# Épico W — Site de produto + portal de docs

**Status:** em implementação (`website/`) — MVP neste PR  
**Meta:** publicar o site bilíngue com docs em MDX mantidas no repositório.  
**Stack alvo:** Next.js (App Router) + TypeScript + Tailwind + MDX.  
**Referências:** [`docs/design/control-plane/design-files/*.dc.html`](../design/control-plane/design-files/) · [HANDOFF.md](../design/control-plane/HANDOFF.md)

---

### W-01 · Scaffold do site `M`
Next.js + TS + Tailwind + MDX (contentlayer ou next-mdx-remote). ESLint/Prettier. Deploy preview (Vercel ou Pages).
- [x] `pnpm dev` roda; página exemplo com fontes Schibsted Grotesk + IBM Plex Mono via `next/font`
- [x] Tokens do handoff (cores/raios/espaços) em `tailwind.config.ts` (`arah.bg`, `arah.surface`, `arah.line`, `arah.accent`…)
- [x] CI: lint + typecheck + build em PR

### W-02 · Layout global: Nav, Footer, seletor de idioma `M`
- [x] Nav sticky com blur, wrap em telas estreitas, logo SVG do grafo, CTAs GitHub/Get Started
- [x] Seletor EN|PT liga rotas `/en/*` ↔ `/pt/*`
- [x] Footer com tagline + GitHub/Releases/Changelog/Contributing/License
- [ ] Acessível por teclado; foco visível; contraste AA nos tons `#9AA5B1`+

### W-03 · i18n `M`
- [x] `/en` e `/pt` estáticos (SSG), `hreflang` correto
- [x] Código/comandos/paths nunca traduzidos
- [x] 404 localizada

### W-04 · Home §1–4: hero animado, shift, problema, inversão `G`
- [ ] Pipeline de 7 estágios com `flowDot` + `stagePulse` (CSS puro; `prefers-reduced-motion`)
- [ ] Grades responsivas idênticas ao design
- [ ] Lighthouse: sem CLS das animações

### W-05 · Home §5: Harness Explorer (13 componentes) `M`
- [ ] Lista + painel sticky com FUNÇÃO/MATURIDADE/EXEMPLO
- [ ] Navegável por teclado; deep-link `?part=slug`

### W-06 · Home §6–9: camadas, princípios, lifecycle, capability map `M`
- [ ] Camadas e capability map interativos com estado em URL-hash
- [ ] Chips do lifecycle linkam para How It Works

### W-07 · Home §10–14: matriz, not-list, status, quick start, CTA `M`
- [ ] Matriz com scroll horizontal; coluna ARAH destacada
- [ ] `CopyButton` reutilizável
- [ ] Painéis Available/Experimental/Planned de `capabilities.yaml` (H-10)

### W-08 · Página Architecture `M`
- [ ] Diagrama de camadas + `FileTree` com comentários
- [ ] 6 cards de contratos + fluxo de update em 4 passos

### W-09 · Página How It Works `G`
- [ ] Timeline de 12 etapas; walkthrough de 9 passos com teclado
- [ ] Terminal pane estilo `#080A0D`

### W-10 · Páginas TechOrganism e Use Cases `M`
- [ ] Ciclo com `cyclePulse`; grade de 11 correspondências; 8 use cases

### W-11 · Portal de docs (MDX) `G`
- [ ] Sidebar do filesystem; breadcrumb; TOC; shiki + CopyButton
- [ ] URLs `/docs/<section>/<slug>`

### W-12 · Busca das docs `M`
- [ ] Índice local (FlexSearch) no build; atalho `/`

### W-13 · CLI Reference interativa `M`
- [ ] Dados de `content/cli.{en,pt}.json` validados contra `cli/arah.ps1` (H-01)
- Depende de: H-01

### W-14 · SEO, OG e qualidade `M`
- [ ] Metadata, sitemap, Lighthouse ≥90, axe sem violações críticas

### W-15 · Conteúdo: calibrar mocks com o repo real `M`
- [ ] Substituir versão/contagens/exemplos pelos reais de `sraphaz/arah-harness`

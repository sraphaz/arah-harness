# ARAH Harness — Website

Site de produto + portal de docs + Live Console (UI mock), bilíngue EN/PT.

## Stack

Next.js 14 (App Router) · TypeScript · Tailwind CSS · conteúdo JSON extraído dos protótipos em `docs/design/control-plane/`.

## Desenvolvimento

```bash
cd website
pnpm install
pnpm dev
```

Abra `http://localhost:3000` (redireciona para `/pt`).

```bash
pnpm lint
pnpm typecheck
pnpm build
```

## Rotas

| Path | Página |
|------|--------|
| `/[locale]` | Home |
| `/[locale]/architecture` | Arquitetura |
| `/[locale]/how-it-works` | Como funciona |
| `/[locale]/techorganism` | TechOrganism |
| `/[locale]/use-cases` | Casos de uso |
| `/[locale]/docs/...` | Portal de docs + CLI explorer |
| `/[locale]/console` | Live Console (mock read-only) |

## Conteúdo

- `content/**` — JSON gerado a partir dos design-files (ver `content/README.md`)
- `capabilities.yaml` — espelho da raiz do harness (status Available/Experimental/Planned)

Regenerar conteúdo dos protótipos:

```bash
node scripts/extract-design-content.mjs
```

## CI

Workflow `.github/workflows/website.yml` — lint + typecheck + build em PRs que tocam `website/`.

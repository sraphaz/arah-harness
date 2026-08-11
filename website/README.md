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

Abra `http://localhost:3000` (redireciona para `/pt`). Localmente **sem** `basePath`.

```bash
pnpm lint
pnpm typecheck
pnpm build          # export estático → out/ (sem basePath)
pnpm build:pages    # igual ao CI (BASE_PATH=/arah-harness)
```

## Deploy — GitHub Pages (automático)

URL do projeto: **https://sraphaz.github.io/arah-harness/**

Workflow: [`.github/workflows/website.yml`](../.github/workflows/website.yml)

- Em todo PR que toca `website/`, `capabilities.yaml`, `VERSION` ou `docs/design/control-plane/`: lint + typecheck + build estático
- Em push para `main` nos mesmos paths: build + **deploy automático** via GitHub Actions
- Também pode rodar manualmente (*Actions → website → Run workflow*)

Source do Pages já está em **GitHub Actions**. Sem custom domain por enquanto — quando houver, apontar o DNS em Pages → Custom domain (sem alterar o `BASE_PATH` do project site, ou migrar para user site).

Conteúdo visual vem dos protótipos Control Plane em `docs/design/control-plane/` (pacote de handoff). Após atualizar os `.dc.html`, regenere com `node scripts/extract-design-content.mjs`.

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

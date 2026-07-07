# Specs — spec-before-work

Este repositório usa **spec-before-work**: toda mudança de código deve estar coberta por uma spec em `docs/specs/`.

## Convenção

- Formato: Markdown com frontmatter YAML (`---` … `---`).
- Campos obrigatórios: `id`, `title`, `status`, `owner`, `covers`, `updated_at`.
- Status: `draft` | `active` | `deprecated`.
- `covers`: globs de paths que a spec governa (ex.: `"app/**"`, `"components/**"`).

## Fluxo

1. Criar ou atualizar spec **antes** de implementar.
2. Referenciar `Spec-Id:` no PR.
3. CI (`validate-specs`) bloqueia mudanças fora de spec sem label `spec-reviewed`.

## Template

Use `_template.spec.yaml` ou copie uma spec existente como base.

## Registry

Specs ativas devem constar em `docs/specs/REGISTRY.md` (profile consulting).

# Specs — spec-before-work

Este repositório usa **spec-before-work**: mudanças cobertas por spec em `docs/specs/`, com `Spec-Id:` no PR.

## Convenção

| Campo | Obrigatório | Notas |
|-------|-------------|-------|
| `id` | sim | Identificador estável |
| `title` | sim | Título humano |
| `status` | sim | `draft` \| `active` \| `deprecated` |
| `owner` | sim | Agente ou papel |
| `covers` | sim | Globs governados |
| `updated_at` | sim | ISO date |
| `acceptance` | recomendado | EARS + `covered_by` |

Formatos aceitos: YAML (`.spec.yaml`) ou Markdown com frontmatter.

## Fluxo

1. Criar ou atualizar a spec **antes** de implementar  
2. Referenciar `Spec-Id:` no corpo do PR  
3. CI (`validate-specs`) bloqueia mudanças fora de spec sem label `spec-reviewed`  

## Specs ativas

Ver [REGISTRY.md](REGISTRY.md).

| Spec-Id | Tema |
|---------|------|
| `arah-biocomponent` | Discovery, organismo, sinais, evolve, regenerate |

## Template

[`_template.spec.yaml`](_template.spec.yaml)

# Backlog — ARAH Control Plane

Fonte: handoff **ARAH Control Plane** (jul/2026). Três épicos paralelizáveis.

| Épico | Escopo | Status neste PR |
|-------|--------|-----------------|
| **[W](./W-site-and-docs.md)** | Site de produto + portal de docs (UI) | **Backlog** — não implementar agora |
| **[C](./C-live-console.md)** | ARAH Live Console (UI + serviço) | **Backlog** — não implementar agora |
| **[H](./H-harness-improvements.md)** | Melhorias do harness (CLI, estado, segurança) | **Em andamento** — curto prazo neste PR |

Design tokens, fidelidade e protótipos HTML: [`docs/design/control-plane/`](../design/control-plane/README.md).

## Sequenciamento sugerido

- **S1:** W-01→W-03 · H-01, H-02, H-06 · C-01  
- **S2:** W-04→W-06 · C-02, C-03 · H-05  
- **S3:** W-07→W-10 · C-04, C-05 · H-03  
- **S4:** W-11, W-12 · C-06→C-08 · H-04  
- **S5:** W-13→W-15 · C-09, C-10 · H-09, H-10  
- **S6:** lançamento site+console MVP · H-07, H-11 · fase 2 (C-11, C-12, H-08)

**DoD geral:** revisão humana em PR; gates verdes; docs no mesmo PR; EN+PT quando tocar o site.

Legenda de estimativa nos épicos: P (≤½ dia) · M (1–2 dias) · G (3–5 dias).

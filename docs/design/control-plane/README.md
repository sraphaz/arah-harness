# Design — ARAH Control Plane (arquivo histórico)

Handoff de design para o site de produto, portal de docs e Live Console.

> **Arquivado como referência histórica ([ADR-003](../../adr/003-kernel-workspace-boundary.md)):**
> os protótipos que assumem agregação multi-repo/org-level estão fora do escopo do
> kernel. Se alguma dessas features avançar, avança no Surya Labs Workspace.
> Site (W) e console single-repo (C) permanecem válidos como referência de design.

| Artefato | Uso |
|----------|-----|
| [HANDOFF.md](./HANDOFF.md) | Tokens, fidelidade, telas, interações |
| [BACKLOG.md](./BACKLOG.md) | Backlog executável original (espelho) |
| [design-files/](./design-files/) | Protótipos HTML navegáveis (não são código de produção) |
| [`docs/backlog/`](../../backlog/) | Backlog operacional no repo (W/C em espera; H em progresso) |

## Como abrir os protótipos

Abra qualquer `design-files/*.dc.html` no navegador (requer `support.js` / `doc-page.js` na mesma pasta).

## Política deste ciclo

- **Interface visual (épicos W e C):** permanece no backlog — não portar os `.dc.html` para Next.js/console neste PR.
- **Melhorias do harness (épico H):** implementar no kernel/CLI conforme [H-harness-improvements.md](../../backlog/H-harness-improvements.md).

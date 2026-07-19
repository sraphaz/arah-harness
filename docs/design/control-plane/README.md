# Design — ARAH Control Plane

Handoff de design para o site de produto, portal de docs e Live Console.

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

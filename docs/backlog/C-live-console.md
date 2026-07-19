# Épico C — ARAH Live Console

**Status:** C-01→C-09 **done** (serviço Go + UI live/mock); C-10 empacotamento e fase 2 (C-11/C-12) backlog  
**Meta:** MVP do console read-only plugável.  
**Referência:** [`ARAH Live Console.dc.html`](../design/control-plane/design-files/ARAH%20Live%20Console.dc.html) (UI + C4).  
**Decisões travadas:** read-only; artefatos como contrato; índice SQLite descartável; 3 implantações (extensão IDE → app local → serviço org — MVP = app local).

Schemas: [`schemas/console/`](../../schemas/console/). Serviço: [`live/`](../../live/README.md). UI: `website/` → `/[locale]/console` (tenta `arah-live` em `127.0.0.1:8787`, senão mock).

---

### C-01 · Especificar contrato de dados dos artefatos `M` — **done**
- [x] JSON Schemas versionados em `schemas/console/`
- [x] Tipos de evento e prefixos (consultation.*, gates.*, change.*, evolution.*, session.*) com payloads mínimos
- Depende de: H-02.

### C-02 · Scaffold do Live Service `M` — **done**
- [x] Serviço Go `live/cmd/arah-live` com REST + WebSocket, `-repo <path>`
- [x] `GET /api/summary|feed|gates|domains|queue|proposals|graph`; `WS /events`
- [x] Sem endpoint de escrita — 405 + teste

### C-03 · FS Watcher + Bus Reader `G` — **done**
- [x] Watch com debounce em `.arah/local/` (e legado `.arah/bus`, `.arah/audit`); parse incremental
- [x] Eventos malformados → fila de erro no summary (não derruba o stream)
- [x] Testes com fixtures em `live/testdata`

### C-04 · Ledger Indexer (SQLite derivado) `G` — **done**
- [x] Projeção regenerável (`-reindex`); índice em `.arah/local/index/live-index.sqlite`

### C-05 · Graph Builder + Gate Monitor `M` — **done**
- [x] Serve `graph.json` + drift; agrega gates 24h no summary/painel

### C-06 · GitHub Adapter (fila de seleção) `M` — **done**
- [x] Poll de PRs abertos (`ARAH_GITHUB_REPO` + `GITHUB_TOKEN` opcional); ações = links

### C-07 · Console Web — shell + KPIs + seletor de repo `M` — **done**
- [x] Header, chips kernel/drift, indicador ao-vivo; 5 KPIs via REST+WS (fallback mock)

### C-08 · Console Web — Signal Feed + Gate Panel `M` — **done**
- [x] Feed com filtros; painel de gates ✓/✗

### C-09 · Console Web — Territory Map + Selection Queue + Proposals `M` — **done**
- [x] Territórios, fila, propostas com evidência; barras de autonomia

### C-10 · AuthZ + empacotamento app local `M` — backlog
- [ ] OAuth GitHub além de localhost; `npx arah-live` ou binário; README de operação

### C-11 · (Fase 2) Extensão de IDE `G` — backlog
- [ ] Webview VS Code/Cursor apontado ao workspace

### C-12 · (Fase 2) Graph Explorer `G` — backlog
- [ ] Visualização navegável do grafo (único painel ainda não prototipado)

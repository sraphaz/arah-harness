# Épico C — ARAH Live Console

**Status:** backlog (interface visual + serviço — não implementar UI neste ciclo)  
**Meta:** MVP do console read-only plugável.  
**Referência:** [`ARAH Live Console.dc.html`](../design/control-plane/design-files/ARAH%20Live%20Console.dc.html) (UI + C4).  
**Decisões travadas:** read-only; artefatos como contrato; índice SQLite descartável; 3 implantações (extensão IDE → app local → serviço org — MVP = app local).

Schemas de contrato (pré-requisito harness): ver `schemas/console/` quando C-01 for aberto. Depende de tipos de sinal estáveis ([H-02](./H-harness-improvements.md#h-02)).

---

### C-01 · Especificar contrato de dados dos artefatos `M`
- [ ] JSON Schemas versionados em `schemas/console/`
- [ ] Tipos de evento e prefixos (consultation.*, gates.*, change.*, evolution.*, session.*) com payloads mínimos
- Depende de: H-02. Bloqueia C-03+.

### C-02 · Scaffold do Live Service `M`
- [ ] Serviço local (Node/TS ou Go — ver ADR H-06) com REST + WebSocket, `--repo <path>`
- [ ] `GET /api/summary|feed|gates|domains|queue|proposals`; `WS /events`
- [ ] Sem endpoint de escrita — lint de rota

### C-03 · FS Watcher + Bus Reader `G`
- [ ] Watch com debounce em `.arah/local/` (e legado `.arah/bus`, `.arah/audit`); parse incremental
- [ ] Validação contra schemas C-01; malformados → fila de erro
- [ ] Testes com fixtures JSONL reais

### C-04 · Ledger Indexer (SQLite derivado) `G`
- [ ] Projeção regenerável (`--reindex`); KPIs &lt;50ms em 100k eventos

### C-05 · Graph Builder + Gate Monitor `M`
- [ ] Carrega `graph.json` + drift; agrega gates 24h/30d

### C-06 · GitHub Adapter (fila de seleção) `M`
- [ ] Poll/webhook de PRs com evidência; token read-only; ações = links

### C-07 · Console Web — shell + KPIs + seletor de repo `M`
- [ ] Header, chips kernel/drift, indicador ao-vivo; 5 KPIs via REST+WS

### C-08 · Console Web — Signal Feed + Gate Panel `M`
- [ ] Feed com filtros; painel de gates ✓/✗

### C-09 · Console Web — Territory Map + Selection Queue + Proposals `M`
- [ ] Territórios, fila, propostas com evidência; barras de autonomia

### C-10 · AuthZ + empacotamento app local `M`
- [ ] OAuth GitHub além de localhost; `npx arah-live` ou binário; README de operação

### C-11 · (Fase 2) Extensão de IDE `G`
- [ ] Webview VS Code/Cursor apontado ao workspace

### C-12 · (Fase 2) Graph Explorer `G`
- [ ] Visualização navegável do grafo (único painel ainda não prototipado)

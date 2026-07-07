# ARAH Live Session — extensão Cursor / VS Code

Painel lateral com agentes ARAH em tempo real: regras casadas, operacionais, domínios, specialists e subagentes Cursor.

## Rodar local (recomendado agora)

Um comando faz tudo — compila, instala VSIX no Cursor e simula eventos no IAutos:

```powershell
cd C:\Users\rapha\CursorRepos\arah-harness\extension\arah-live
./scripts/dev-local.ps1
```

Depois no Cursor:

1. **Ctrl+Shift+P** → `Developer: Reload Window`
2. Abra o workspace **IAutos**
3. Ícone **ARAH** na barra lateral → **Live Session**

### Modo debug (F5)

Abra a pasta `extension/arah-live` no Cursor → **Run and Debug** → **ARAH Live (Extension Development Host)**. Abre uma janela nova com IAutos e a extensão em modo dev.

### Simular eventos sem hooks

No projeto ARAH (ex. IAutos):

```powershell
./scripts/agents/demo-live-session.ps1
```

## Instalação rápida (VSIX manual)

```powershell
# No arah-harness (gerar ou usar o VSIX existente)
cd extension/arah-live
npm install
npm run compile
npm run package
```

No Cursor:

1. **Extensions** → **…** → **Install from VSIX…**
2. Selecione `extension/arah-live/arah-live-0.1.1.vsix`
3. Recarregue a janela

Alternativa (desenvolvimento): **Install from Folder…** → `extension/arah-live`

## Pré-requisitos no projeto

- Kernel ARAH ≥ 0.2.2 (`arah update -Force`)
- Hooks em `.cursor/hooks.json` (live session + domain review)
- Agent graph: `arah export-graph` → `docs/_meta/agent-graph.generated.json`

## Uso

1. Abra um repo ARAH (ex.: IAutos)
2. Ícone **ARAH** na barra de atividades → **Live Session**
3. **Status bar** (rodapé): `$(pulse) ARAH: frontend · core-cases` quando há atividade

Telemetria em `.cursor/arah-live/`:

| Arquivo | Conteúdo |
|---------|----------|
| `state.json` | Snapshot (agentes, regras, subagentes) |
| `events.jsonl` | Log de eventos |

## O painel mostra

- **Regras ativas** — chips da coreografia casada
- **Operacionais / Domínio / Specialists** — nós que pulsam quando ativos
- **Linhas SVG animadas** — conexões regra → agente (consulta em tracejado fino)
- **Subagentes Cursor** — explore, shell, bugbot…
- **Feed** — últimos eventos

## Comandos

- `ARAH: Abrir Live Session`
- `ARAH: Atualizar Live Session`

## Licença

MIT — [ARAH Harness](https://github.com/sraphaz/arah-harness)

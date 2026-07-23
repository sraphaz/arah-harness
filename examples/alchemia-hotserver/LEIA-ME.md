# Pacote ARAH — Alchemia HotServer (OpenTibia / Canary)

Pacote pronto para instalar o **ARAH Harness** no repositório do jogo em `D:\SERVIDOR NO D`.

## O que vem neste ZIP

| Item | Função |
|------|--------|
| `Install-AlchemiaArah.ps1` | Instalador: clona harness + `arah install` + aplica overlay Alchemia |
| `overlay/` | Config, domínios, skills OT, coreografia, docs, AGENTS mesclado |
| `tools/` | Helpers opcionais |
| `CHECKLIST.md` | Passos manuais pós-install |

## Pré-requisitos (máquina do amigo)

- Windows com PowerShell 5.1+
- Git
- (Recomendado) Visual Studio Build Tools / ambiente que ele já usa para o Canary/OTClient
- LuaJIT em `coisas do codex\tools\luajit_canary\luajit.exe` (já citado no AGENTS.md)

## Instalação em 3 comandos

1. Extrair este ZIP em qualquer pasta (ex.: `D:\Downloads\alchemia-arah-pack`).
2. Abrir PowerShell:
   ```powershell
   cd D:\Downloads\alchemia-arah-pack
   powershell -ExecutionPolicy Bypass -File .\Install-AlchemiaArah.ps1 -Target "D:\SERVIDOR NO D"
   ```
3. Seguir o checklist impresso no final (`CHECKLIST.md`).

O instalador:

1. Clona `arah-harness` em `D:\arah-harness` (se ainda não existir)
2. Roda `arah install` no servidor do jogo (brownfield — não apaga seu `AGENTS.md` antigo)
3. Copia o overlay Alchemia (config, domains, skills, choreography)
4. Faz merge seguro do `AGENTS.md` (backup em `coisas do codex\backups` se a pasta existir)
5. Roda `domain sync`, `validate-manifests` e `doctor` quando possível

## Depois de instalar — fluxo diário

```powershell
cd "D:\SERVIDOR NO D"
./scripts/agents/validate-manifests.ps1
./scripts/agents/invoke-skill.ps1 -Skill lua-validate
./scripts/agents/invoke-skill.ps1 -Skill add-spell   # checklist de nova magia
```

Abrir o projeto no Cursor: os agentes/skills aparecem via `.agents/` e `.skills/`.

## Referências

- Harness: https://github.com/sraphaz/arah-harness
- Canary upstream: https://github.com/opentibiabr/canary
- Skills OT de referência: https://github.com/jasondruid/Nostalrius-Otclient-AI-management

## Importante

- ARAH governa o **desenvolvimento** (PRs, domains, skills). Não é bot in-game.
- Builds C++ continuam pelas regras do Handbook / `build_otclient_release.bat`.
- Artefactos Codex continuam só em `coisas do codex\`.

# Alchemia × ARAH Harness

Guia curto para humanos e agentes no HotServer (Canary 3.4.1 + OTClient).

## Por que ARAH aqui

O datapack OpenTibia cresce em muitos eixos (magia, monstro, quest, cliente, C++).
Sem coreografia, agentes pisam uns nos outros e quebram higiene de build.

ARAH traz:

1. Domínios consultivos por área de jogo
2. Skills repetíveis (add-spell, balance-pass, lua-validate…)
3. Gates baratos (LuaJIT) antes de rebuild
4. Handbook/Manual como lei (ops-codex)

## O que ARAH **não** é

- Não é CaveBot / AttackBot / AI de personagem in-game
- Não substitui Remere's Map Editor ou Assets Editor
- Não abre merge sozinho — humano manda

## Layout esperado

```
SERVIDOR NO D/
├── canary-3.4.1/          # server
├── client_run/            # OTClient
├── coisas do codex/       # handbook, backups, tools
├── arah.config.yaml
├── AGENTS.md
├── .agents/
├── .skills/
└── scripts/agents/
```

Se as pastas tiverem outros nomes, edite `arah.config.yaml` e choreography.

## Referências externas

| Projeto | Uso |
|---------|-----|
| [opentibiabr/canary](https://github.com/opentibiabr/canary) | Upstream engine/datapack |
| [Nostalrius-Otclient-AI-management](https://github.com/jasondruid/Nostalrius-Otclient-AI-management) | Skills `/add-spell`, `/balance`, … |
| [sraphaz/arah-harness](https://github.com/sraphaz/arah-harness) | Kernel instalado |
| [sraphaz/iautos](https://github.com/sraphaz/iautos) | Consumidor ARAH de referência |

## Comandos

```powershell
$env:ARAH_HARNESS_PATH = "D:\arah-harness"
./scripts/agents/validate-manifests.ps1
./scripts/agents/invoke-skill.ps1 -Skill lua-validate
./scripts/agents/invoke-skill.ps1 -Skill add-spell
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 doctor -Target .
```

# AGENTS.md — Alchemia HotServer

Nosso foco e desenvolver um servidor de jogo 2d totalmente editado inspirado no Tibia e na magia.

**ARAH Harness 0.2.0** — camada de agentes para desenvolvimento coreografado (não é bot in-game).

## Documentos base obrigatorios

- Porta de entrada operacional do projeto: `coisas do codex\ALCHEMIA_ENGINEERING_HANDBOOK_V2_0.txt`.
- Manual historico append-only: `coisas do codex\MANUAL MESTRE CODEX ALCHEMIA.txt`.
- Antes de revisar sistema grande, balanceamento, bug antigo ou fluxo com risco de regressao, ler o Handbook primeiro; se envolver rollback, hash, backup, AppData, build especifico, decisao conflitante ou lacuna, consultar tambem o Manual Mestre, wiki e relatorios.
- O Handbook deve ser atualizado pelo mesmo metodo do Manual: sem apagar historico, com consolidacao nova, decisao mais recente/especifica vencendo e registro de sintoma, causa, arquivos, validacao, pendencias e comando de teste.
- Para balanceamento geral de itens, usar o capitulo 19 do Handbook como ponto de partida e gerar/consultar planilhas tecnicas antes de mexer em `items.xml`.

## Regras operacionais deste workspace

- Antes de qualquer build, verificar processos presos de compilacao/cliente com:
  `tasklist /FO CSV | findstr /I "cmake ninja cl.exe link.exe msbuild devenv otclient"`
- Nao rodar build de cliente/servidor longo no foreground. Se precisar compilar C++, iniciar em background com log, acompanhar por polling e avisar o usuario a cada 5 minutos.
- Para mudancas so em Lua/OTUI/XML/catalogos, nao compilar o cliente/servidor; basta informar que precisa reload/restart do processo correspondente.
- Para mudancas C++ do `client_run`, nunca chamar Ninja direto a partir de um PowerShell cru, porque o link pode falhar sem as libs do Windows SDK como `crypt32.lib`.
- O caminho correto e carregar o ambiente x64 do Visual Studio e entao chamar Ninja. Use o script:
  `client_run\build_otclient_release.bat`
- Se `build\windows-release` ou arquivos CMake tiverem sido arquivados em `coisas do codex`, restaurar/recriar a configuracao antes de chamar o script de build.
- Depois de build C++ bem-sucedido do cliente, copiar o binario gerado para o executavel que o usuario usa quando aplicavel, preservando backup.
- Se `cmake.exe`, `ninja.exe`, `cl.exe` ou `link.exe` ficarem vivos sem progresso, interromper o build, matar somente esses processos de build e relatar o estado.
- Regra de higiene permanente: nao criar nem deixar em `canary-3.4.1` ou `client_run` arquivos de trabalho do Codex, builds/cache, relatorios, backups, PDBs, logs, previews, sprites intermediarios ou pacotes auxiliares. O destino padrao para tudo isso e `coisas do codex`.
- Nas pastas do server/client so deve entrar o que faz parte do funcionamento real do servidor/cliente: codigo ativo, XML/OTUI/Lua carregado, assets finais, sprites/DAT/OTBM ativos, configs e executaveis/DLLs usados em runtime.
- Se for necessario preservar backup antes de uma edicao, criar em `coisas do codex\backups` ou em uma pasta de limpeza com manifesto. Se for necessario gerar build/cache, arquivar em `coisas do codex` quando terminar ou quando nao estiver em uso.

## Validacao rapida sem build

- Para mudancas Lua/OTUI, validar sintaxe com o LuaJIT local usando `loadfile`, sem executar o servidor:
  `coisas do codex\tools\luajit_canary\luajit.exe -e "local f,err=loadfile([[CAMINHO/arquivo.lua]]); if not f then error(err) end"`
- Nao usar `luajit -b` neste workspace: esta build local retorna erro de comando/modulo JIT.
- Se o arquivo depender de globais do Canary, nao executar o chunk; apenas `loadfile`.

## Sistemas custom atuais

- Party buffs finais ficam em `canary-3.4.1\data\scripts\spells\party\alchemia_party_buffs.lua`.
- As versoes antigas/atalhos por vocacao tambem existem em `support\aegis_mora.lua`, `vita_flora.lua`, `fera_celer.lua`, `umbra_pactum.lua` e `ignis_volt.lua`.
- Os bonus temporarios e o pulso visual/status bar dos party buffs ficam em `canary-3.4.1\data\scripts\custom_skills.lua`.
- Bursts mid game ficam em `canary-3.4.1\data\scripts\spells\attack\alchemia_mid_bursts.lua`; o padrao final e sequencia de 3 hits com efeito aplicado no ultimo hit.
- Magical Archive fica em `client_run\modules\game_cyclopedia\tab\magicalArchives\magicalArchives.lua` e deve receber override explicito para cada magia custom importante: elemento, area, tipo, descricao e preview.
- O arquivo OTUI do Magical Archive fica em `client_run\modules\game_cyclopedia\tab\magicalArchives\magicalArchives.otui`.

## Avatar / Power Beasts

- A selecao da Power Beast define instancia global do personagem: sprite do avatar, familiar e walking effect durante avatar.
- O avatar real e ativado por `Animagis`/`Fusio Bestialis`/efeito de legs, nao ao selecionar a instancia.
- O walking field atual e criado em `canary-3.4.1\data\scripts\creaturescripts\others\power_beasts.lua`.
- Hoje `USE_NATIVE_AVATAR_FIELD_ITEMS = false`; isso significa que o sistema aplica dano/efeito por tile temporario e envia efeitos visuais, mas nao deixa um item de field persistente no chao. Para field visual persistente sem dano duplicado, criar itens visuais custom sem dano nativo antes de ativar item no chao.

---

## ARAH (SDLC do repositório)

Camada de agentes para desenvolvimento autônomo coreografado — distinta de bots in-game (CaveBot, AttackBot, etc.).

| Camada | Onde | Papel |
|--------|------|--------|
| **ARAH SDLC** | `.agents/`, `.skills/`, `scripts/agents/` | PR, QA, specs, gates, domínio consultivo |
| **Game runtime** | `canary-3.4.1/`, `client_run/` | Servidor Canary + OTClient em execução |

```powershell
# Harness: D:\arah-harness (ou $env:ARAH_HARNESS_PATH)
./scripts/agents/validate-manifests.ps1
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 domain sync -Target .
powershell -File $env:ARAH_HARNESS_PATH\cli\arah.ps1 export-graph -Target .
./scripts/harness/validate-agent-graph.ps1
```

Config: [`arah.config.yaml`](arah.config.yaml)  
Coreografia: [`.agents/choreography.alchemia.yaml`](.agents/choreography.alchemia.yaml)  
Guia: [`docs/arah/ALCHEMIA_ARAH.md`](docs/arah/ALCHEMIA_ARAH.md)

**Princípios ARAH**: humano comanda merge · tudo via PR · escopo mínimo · handbook-first · pareceres de domínio passivos.

### Domínios consultivos

| Domínio | Paths principais |
|---------|------------------|
| `combat-magic` | spells, custom_skills |
| `power-beasts` | power_beasts / avatar |
| `items-economy` | items.xml (+ cap. 19 Handbook) |
| `monsters-spawns` | monster/, world/ |
| `quests-npcs` | npc/, actions/, quests |
| `client-ux` | client_run/modules, OTUI |
| `engine-cpp` | src/ Canary e OTClient |
| `ops-codex` | coisas do codex, AGENTS, config |

### Skills Alchemia (além do kernel)

```powershell
./scripts/agents/invoke-skill.ps1 -Skill lua-validate
./scripts/agents/invoke-skill.ps1 -Skill add-spell
./scripts/agents/invoke-skill.ps1 -Skill add-monster
./scripts/agents/invoke-skill.ps1 -Skill balance-pass
./scripts/agents/invoke-skill.ps1 -Skill power-beast-touch
./scripts/agents/invoke-skill.ps1 -Skill hot-reload-check
./scripts/agents/invoke-skill.ps1 -Skill build-client
./scripts/agents/invoke-skill.ps1 -Skill codex-append
```

### Fluxo

```
Intenção (humano) → Orquestrador → Agente + skills → PR → CI + domínio → ready-for-merge → Merge (humano)
```

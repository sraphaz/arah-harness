# Para quem vai instalar (HotServer / Alchemia)

Olá — este ZIP deixa o **ARAH Harness** pronto no teu servidor OpenTibia.

## Em 1 minuto

1. Extrai o ZIP
2. PowerShell:
   ```powershell
   cd caminho\onde\extraiu\alchemia-arah-pack
   powershell -ExecutionPolicy Bypass -File .\Install-AlchemiaArah.ps1 -Target "D:\SERVIDOR NO D"
   ```
3. Abre `CHECKLIST.md` e marca os itens
4. Abre `D:\SERVIDOR NO D` no **Cursor** e pede:
   > lê AGENTS.md e docs/arah/ALCHEMIA_ARAH.md — depois roda a skill add-spell para a próxima magia

## O que isso muda

- Agentes por domínio (magia, monstro, items, cliente, C++, handbook…)
- Skills: `lua-validate`, `add-spell`, `balance-pass`, `power-beast-touch`…
- Teu `AGENTS.md` antigo é preservado (backup) e mesclado com a seção ARAH
- Kernel ARAH fica em `D:\arah-harness` (clone automático)

## Não é

Bot de caça / CaveBot. É governança de **desenvolvimento** com agentes de IA.

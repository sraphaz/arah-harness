# Checklist pós-install — Alchemia + ARAH

Marque na ordem.

## A. Instalação

- [ ] ZIP extraído
- [ ] `Install-AlchemiaArah.ps1 -Target "D:\SERVIDOR NO D"` terminou sem erro fatal
- [ ] Existe `D:\arah-harness` (ou path customizado)
- [ ] Em `D:\SERVIDOR NO D` existem: `arah.config.yaml`, `.arah-version`, `.agents\`, `.skills\`, `scripts\agents\`
- [ ] `AGENTS.md` contém seção **ARAH** + regras Alchemia originais
- [ ] Backup do AGENTS antigo em `coisas do codex\backups\` (se a pasta existia)

## B. Validação

```powershell
cd "D:\SERVIDOR NO D"
powershell -File D:\arah-harness\cli\arah.ps1 doctor -Target .
powershell -File .\scripts\agents\validate-manifests.ps1
powershell -File D:\arah-harness\cli\arah.ps1 export-graph -Target .
```

- [ ] `doctor` OK
- [ ] manifests OK
- [ ] `docs\_meta\agent-graph.generated.json` gerado (ou path equivalente)

## C. Primeiro uso com agente

- [ ] Abrir pasta do servidor no Cursor
- [ ] Pedir: "leia AGENTS.md e docs/arah/ALCHEMIA_ARAH.md — depois proponha PR para documentar uma magia custom no Magical Archive"
- [ ] Rodar skill `lua-validate` em um arquivo Lua conhecido
- [ ] Rodar skill `add-spell` (gera checklist; não inventa balance)

## D. Git (quando o repo estiver versionado)

```powershell
git add arah.config.yaml .arah-version AGENTS.md .agents .skills scripts .cursor docs/arah .github/workflows/agents-validate.yml
git commit -m "chore: bootstrap ARAH Harness for Alchemia HotServer"
git push
```

- [ ] Commit feito
- [ ] Branch protection em `main` (opcional mas recomendado)
- [ ] Workflow `agents-validate.yml` ativo no GitHub

## E. Higiene (já no AGENTS)

- [ ] Nada de PDBs/logs/relatórios Codex dentro de `canary-3.4.1` ou `client_run`
- [ ] Builds longos só em background
- [ ] Handbook-first antes de sistemas grandes

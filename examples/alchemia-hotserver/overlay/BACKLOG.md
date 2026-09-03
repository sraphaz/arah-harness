# BACKLOG — Alchemia HotServer (ARAH)

Entrega guiada por fases. Ajuste IDs conforme o Handbook.

## Fase A — Bootstrap ARAH ✅ (este pack)

| ID | Entrega |
|----|---------|
| A-01 | `arah install` + overlay Alchemia |
| A-02 | Domínios + choreography |
| A-03 | Skills lua-validate / add-spell / balance-pass |

**Gate:** `doctor` + `validate-manifests` OK; AGENTS.md com seção ARAH.

## Fase B — Magia & Archive ← **sugestão atual**

| ID | Entrega | Domínio |
|----|---------|---------|
| B-01 | Inventário de magias custom vs overrides Magical Archive | combat-magic, client-ux |
| B-02 | Checklist add-spell no fluxo de toda magia nova | combat-magic |
| B-03 | Party buffs: documentar no Handbook estado final vs atalhos support/ | ops-codex |

**Gate:** nenhuma magia custom importante sem preview no Archive.

## Fase C — Power Beasts

| ID | Entrega |
|----|---------|
| C-01 | Documentar matriz beast → sprite/familiar/field |
| C-02 | Decisão field persistente vs temporário (itens sem dano nativo) |

## Fase D — Balance & Economia

| ID | Entrega |
|----|---------|
| D-01 | Planilha cap.19 vigente arquivada em `coisas do codex` |
| D-02 | Todo PR de items.xml passa skill `balance-pass` |

## Fase E — Conteúdo (monsters / quests)

| ID | Entrega |
|----|---------|
| E-01 | Pipeline add-monster |
| E-02 | Pipeline quests com storage map |

## Como operar

```powershell
./scripts/agents/invoke-skill.ps1 -Skill add-spell
./scripts/agents/invoke-skill.ps1 -Skill lua-validate
```

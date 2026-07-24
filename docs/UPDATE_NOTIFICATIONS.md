# Notificações de atualização do harness

**Problema:** consumidores do ARAH pinam a versão em `.arah-version`, mas sem um canal de notificação ficam para sempre na versão instalada.

## Padrão de mercado (o que adotamos)

Para tooling/template distribuído por **git clone + copy** (não npm/NuGet), o padrão consolidado é:

| Camada | Prática | Exemplos |
|--------|---------|----------|
| Fonte de verdade | **GitHub Releases** (`vX.Y.Z`) | Spec Kit, copier templates, actions |
| Drift de arquivos | `sync --check` / hash | harnessforge, ARAH `sync-check` |
| Notificar consumidor | **Cron no repo consumidor** → issue/PR | muitos “template sync” bots |
| Opcional PR automático de pin | **Renovate** regex manager | Renovate customManagers |

**Não** usamos Dependabot como updater principal do kernel — ele espera ecossistemas de pacote. Dependabot continua útil para Actions/npm do próprio repo.

Fluxo ARAH:

```text
Humano: PR com VERSION + CHANGELOG → merge em main
        ↓
Harness (Actions): cut-release → tag vX.Y.Z + GitHub Release  (autônomo)
        ↓
Consumidor (weekly workflow): lê .arah-version → compara com releases/latest
        ↓
outdated → abre/atualiza issue `arah-harness-update`
        ↓
Humano (consumidor): checkout tag → arah update -Force → PR
```

Alinhado a [MARKET_REFERENCE.md](MARKET_REFERENCE.md) (copier update, harnessforge sync --check) e à governança “humano mergeia o bump; a máquina publica o release”.

## O que foi implementado

1. **`arah update-check`** / `scripts/agents/check-harness-update.ps1`  
   Compara pin × latest release (API GitHub). Exit `2` se outdated.
2. **Workflow consumidor** `harness-update-check.yml` (instalado no `init`)  
   Cron semanal + `workflow_dispatch`; cria issue com label `arah-harness-update`.
3. **Workflow upstream** `.github/workflows/release.yml` + **`cut-release.ps1`**  
   Em todo push em `main` (e `workflow_dispatch`): se ainda não existir Release para `VERSION`, cria tag anotada + GitHub Release. Idempotente.
4. **CLI** `arah release cut` — mesmo corte, local/CI (requer `GITHUB_TOKEN`).
5. **Template Renovate** `templates/renovate-arah.json`  
   Opcional: PR só no pin `.arah-version`.

## Uso no consumidor

```powershell
# Manual
powershell -File ./scripts/agents/check-harness-update.ps1
# ou
powershell -File $env:ARAH_HARNESS_PATH/cli/arah.ps1 update-check

# Offline / teste
./scripts/agents/check-harness-update.ps1 -LatestVersion 0.9.0
```

Atualizar quando notificado:

```powershell
git -C $env:ARAH_HARNESS_PATH fetch --tags
git -C $env:ARAH_HARNESS_PATH checkout v0.4.1
powershell -File $env:ARAH_HARNESS_PATH/cli/arah.ps1 update -Force
# commit .arah-version + diff do kernel → PR
```

## Config (`arah.config.yaml`)

```yaml
update_check:
  enabled: true
  repository: sraphaz/arah-harness   # owner/name upstream
  notify:
    issue: true
    label: arah-harness-update
```

`init`/`regenerate` acrescentam o bloco se ausente (não sobrescrevem customizações).

## Renovate (opcional)

Copie [`templates/renovate-arah.json`](../templates/renovate-arah.json) para `renovate.json` no consumidor (ou faça merge com sua config). O Renovate abre PR alterando só o pin; o humano ainda precisa rodar `arah update -Force`.

## Publicar release no harness (autônomo)

1. Abra PR que altera `VERSION` + seção correspondente em `CHANGELOG.md` (+ pin `.arah-version` se dogfooding).
2. Merge em `main` (humano).
3. O workflow **Release** corre sozinho e publica `vX.Y.Z` + GitHub Release.

Manual / backfill:

```bash
# Disparar na UI: Actions → Release → Run workflow
# ou localmente (com token):
powershell -File ./cli/arah.ps1 release cut
# dry-run:
powershell -File ./scripts/agents/cut-release.ps1 -DryRun
```

Sem Release/tag, o `update-check` do consumidor cai para a tag semver mais recente; se não houver nenhuma, falha pedindo publicação.

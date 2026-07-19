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
Harness: tag vX.Y.Z → GitHub Release
        ↓
Consumidor (weekly workflow): lê .arah-version → compara com releases/latest
        ↓
outdated → abre/atualiza issue `arah-harness-update`
        ↓
Humano: checkout tag → arah update -Force → PR
```

Alinhado a [MARKET_REFERENCE.md](MARKET_REFERENCE.md) (copier update, harnessforge sync --check) e à governança “humano faz merge”.

## O que foi implementado

1. **`arah update-check`** / `scripts/agents/check-harness-update.ps1`  
   Compara pin × latest release (API GitHub). Exit `2` se outdated.
2. **Workflow consumidor** `harness-update-check.yml` (instalado no `init`)  
   Cron semanal + `workflow_dispatch`; cria issue com label `arah-harness-update`.
3. **Workflow upstream** `.github/workflows/release.yml`  
   Publica Release a partir de tags `v*.*.*`.
4. **Template Renovate** `templates/renovate-arah.json`  
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

## Publicar release no harness

```bash
# VERSION e CHANGELOG já bumpados no PR
git tag v0.4.1
git push origin v0.4.1
# Actions → Release workflow cria o GitHub Release
```

Sem Release/tag, o check cai para a tag semver mais recente; se não houver nenhuma, falha pedindo publicação.

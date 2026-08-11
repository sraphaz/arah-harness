# Playbook — README Beautify (fila one-by-one)

Skill: [oil-oil/beautify-github-readme](https://github.com/oil-oil/beautify-github-readme)  
Instalar (uma vez por máquina/repo): `npx skills add oil-oil/beautify-github-readme`

## Padrão por repositório

1. **Move** o agent root para o repo (`move_agent_to_root`).
2. **Audit** read-only do README (clareza, prova, first-use).
3. **Asset-only**: hero SVG + headers (tema nativo; sem inventar features).
4. **Preview** local (`assets/readme/preview.html`).
5. **Review** humana → aprovar assets.
6. **Embed PR** (só após OK): inserir snippets, branch, PR — sem push na default.
7. **Next** na fila só depois do merge/aprovação do passo atual.

Não redesenhar vários repos em paralelo. Não commit/push sem pedido explícito.

## Fila

| # | Repo | Path | Status |
|---|------|------|--------|
| 1 | arah-harness | `CursorRepos/arah-harness` | **Piloto** — assets gerados; aguarda approve/embed |
| 2 | sky-forge | `CursorRepos/sky-forge` | Próximo |
| 3 | Hawk | `CursorRepos/Hawk` | — |
| 4 | iautos | `CursorRepos/iautos` | — |
| 5 | arah | `CursorRepos/arah` | — |
| 6 | surya-labs-workspace | `CursorRepos/surya-labs-workspace` | — |
| 7 | Plantae_BR | `CursorRepos/Plantae_BR` | — |
| 8 | raphael-silva-site | `CursorRepos/raphael-silva-site` | — |
| 9 | SacredChants | `CursorRepos/SacredChants` | — |

## Direção visual (ARAH Harness)

```text
Palette: #0A0C0F / #E7ECF1 / #3EE0C8 / #0A7A5A / #9AA5B1
Motif: células conectadas + ciclo discover → evolve
Composition: technical split (título + painel do ciclo)
Mode: pure SVG (estático)
```

Assets do piloto: `assets/readme/`

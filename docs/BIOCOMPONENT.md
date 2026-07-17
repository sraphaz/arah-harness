# Biocomponente ARAH

**Versão**: 0.3.0 · **Spec**: `arah-biocomponent`

O ARAH Harness passa a operar como **biocomponente tecnológico** instalado no repositório: observa domínio e stack, propõe células (agentes), define comunicação tipada entre elas e evolui por seleção via Pull Request.

## Metáfora operacional

| Natureza | ARAH |
|----------|------|
| Organismo | Harness instalado no repo |
| Célula | Agente (operacional / domínio / specialist) |
| Tecido | Grupo de agentes com tópico compartilhado |
| Sinal químico | Mensagem tipada em `.arah/bus/signals.jsonl` |
| Homeostase | `arah regenerate` |
| Evolução | `arah evolve` → propostas → PR → merge humano |
| Ontogenia | `arah organism bootstrap` (primeiro momento) |

## Princípio de governança

**Agentes propõem; humanos aplicam.**

- Não há criação silenciosa de agentes.
- Discovery e evolve escrevem artefatos em `docs/_meta/`.
- `-Apply` só mescla candidaturas revisáveis em `arah.config.yaml`.
- Kernel nunca é reescrito em silêncio — consumidores recebem via `update` / `regenerate -UpdateKernel`.

Isso preserva auditabilidade e reinterpretá o anti-padrão “agentes que criam agentes” como **loop de proposta + gate**.

## Ciclo de vida

```text
install/update
    → discover          # percebe stack + domínio
    → organism bootstrap  # define células, tecidos, vias
    → (trabalho diário + sinais)
    → evolve            # aprende com audit/sinais/live
    → regenerate        # homeostase completa
    → PR humano         # seleção natural
```

## CLI

```powershell
arah discover [-Apply] [-DryRun]
arah organism bootstrap [-Force]
arah organism status
arah organism signal -From orchestrator -Type attract -To backend -Topic delivery
arah evolve [-Apply]
arah regenerate [-UpdateKernel] [-Force] [-ApplyDiscovery]
```

## Artefatos

| Artefato | Função |
|----------|--------|
| `docs/_meta/discovery.proposed.yaml` | Observação + propostas de domínio/stack |
| `docs/_meta/organism.manifest.yaml` | Mapa vivo de células e tecidos |
| `docs/_meta/evolution.proposed.yaml` | Propostas de self-learning |
| `.arah/bus/signals.jsonl` | Barramento append-only |
| `.arah/organism/state.json` | Estado ontogênico |

## Tipos de sinal

`attract` · `consult` · `propose` · `acknowledge` · `coalesce` · `evolve` · `status`

Default: modo **passive** (arquivo + CI, economia de tokens). Sinais ativos são propostas, não chat multi-turno obrigatório.

## Atualizar consumidores

Quem já usa o harness:

```powershell
# No clone do arah-harness (após pull da release)
powershell -File cli/arah.ps1 regenerate -Target C:\meu-projeto -UpdateKernel -Force
```

Isso reaplica o kernel, regenera domínios, rediscovery, rebootstrap do organismo, evolve e doctor — e deixa sugestões em `docs/_meta/` para o repo evoluir.

## Schemas

- `schemas/arah-harness/discovery.schema.yaml`
- `schemas/arah-harness/organism.schema.yaml`
- `schemas/arah-harness/signal.schema.yaml`
- `schemas/arah-harness/evolution.schema.yaml`

## Referências

- Spec: [docs/specs/arah-biocomponent.spec.yaml](specs/arah-biocomponent.spec.yaml)
- Método: [METHOD.md](METHOD.md)
- Governança: [GOVERNANCE.md](GOVERNANCE.md)
- Mercado: [MARKET_REFERENCE.md](MARKET_REFERENCE.md)

# Compatibilidade de sinais

**Schema:** `arah-harness/signal` · **versão do schema:** 0.2.0 · **wire `v`:** 1

## Política

Alterações no contrato de sinal são **somente aditivas**:

1. Novos campos opcionais podem ser adicionados.
2. Novos valores de enum exigem bump de `v` **e** documento de migração.
3. Remover ou renomear campos / tipos existentes é **proibido** sem major do harness.
4. Leitores devem ignorar campos desconhecidos (forward-compatible).

## Tipos congelados (`type`)

| Tipo | Semântica |
|------|-----------|
| `attract` | Solicita atenção / co-ativação |
| `consult` | Parecer consultivo |
| `propose` | Proposta — requer Apply humano |
| `acknowledge` | Confirma recebimento |
| `coalesce` | Forma/reforça tecido |
| `evolve` | Gatilho de ciclo de evolução |
| `status` | Heartbeat / homeostase |

## Campo `v`

Todo evento emitido por `signal-bus.ps1` inclui `v: 1`. Consumidores (evolve, Live Console) devem:

- Aceitar eventos sem `v` como legado (`v=0` implícito).
- Tratar `v >` suportado como erro suave (log + skip), não crash.

## Storage

| Camada | Path |
|--------|------|
| Quente (pending) | `.arah/local/bus/pending/<ULID>.json` |
| Quente (archive) | `.arah/local/bus/archive/YYYY-MM.jsonl` |
| Legado | `.arah/bus/signals.jsonl` (lido; migrar com `arah migrate-state`) |

Ver [STATE_MODEL.md](STATE_MODEL.md).

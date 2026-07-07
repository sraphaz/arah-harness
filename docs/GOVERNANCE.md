# GOVERNANCE — ARAH Harness

**Versão**: 1.1 · **Data**: 2026-07-06

O ARAH Harness governa repositórios mantidos por agentes. **Domain agents, governança, observabilidade e auditoria fazem parte do harness-model** — ver [MODEL.md](MODEL.md).

## Princípios

1. **Spec-before-work** — mudanças estruturais exigem spec ou ADR.
2. **Domínio consultivo** — agentes de domínio nunca ultrapassam `consult`.
3. **Human-in-the-loop** — merge, release e ações destrutivas são humanas.
4. **Auditabilidade** — toda ação relevante gera evento em `.arah/audit/events.jsonl`.

## Níveis de autonomia (0–6)

| Nível | ID | Capacidade |
|-------|-----|------------|
| 0 | `observe` | Ler specs, coreografia, código |
| 1 | `consult` | Parecer consultivo (domínio) |
| 2 | `route` | Handoff e opções (orquestrador) |
| 3 | `activate` | Escrever docs, specs, manifests |
| 4 | `invoke_skill` | Executar skills/scripts |
| 5 | `side_effect` | Release, deploy, registry externo |
| 6 | `public` | Release público — gate humano obrigatório |

Fonte: [`.agents/autonomy.yaml`](../.agents/autonomy.yaml)

## Agentes de domínio

| ID | Função |
|----|--------|
| `clean-craft-advisor` | Craftsmanship (SOLID, boundaries, smells) |
| `test-architect` | Estratégia de testes TEA, risk-based, gates CI |
| `architecture-documenter` | C4, ADRs, jornadas, consistência documental |

Agentes operacionais (backend, frontend, qa, pr-steward, etc.) ativam checklists e skills dentro do escopo declarado.

## Gates humanos

| Gate | Bloqueia |
|------|----------|
| `spec_before_work` | Skills e side effects sem spec |
| `release_approval` | Corte de release / deploy |
| `destructive` | Ações irreversíveis |

Registrar aprovações em `.arah/approvals.yaml`:

```yaml
spec_before_work: approved
release_approval: pending
```

## Verificar antes de agir

```powershell
./scripts/agents/check-autonomy.ps1 -AgentId release -Action release.cut
./scripts/harness/validate-agent-graph.ps1
./scripts/agents/validate-manifests.ps1
```

## Profiles

Ver [HARNESS_PROFILES.md](HARNESS_PROFILES.md) — `minimal`, `consulting`, `product`, `enterprise`, `open-source`.

## Referências

- [MODEL.md](MODEL.md) — harness-model (first-class)
- [AUDIT.md](AUDIT.md) — trilha de eventos
- [OBSERVABILITY.md](OBSERVABILITY.md) — telemetria e diagnósticos
- [METHOD.md](METHOD.md) — método ARAH

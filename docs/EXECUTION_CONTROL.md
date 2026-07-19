# Execution Control Protocol

**Versão:** 1.0 · **Spec:** [`docs/specs/arah-execution-control.spec.yaml`](specs/arah-execution-control.spec.yaml)

## Problema resolvido

O ARAH já tinha agentes, coreografia, autonomia e gates, mas a sessão podia entrar em ciclos de análise → consulta → handoff → reanálise sem alteração concreta. O orquestrador roteia corretamente e não altera produto; a coreografia, porém, podia co-ativar vários agentes sem um **executor primário** canônico.

## Princípio central

> Uma tarefa pode possuir vários participantes, mas deve possuir **exatamente um** executor primário.

Consultores e revisores participam; só o executor conduz a entrega.

## Ciclo de vida

```text
intenção
  → roteamento (orquestrador)
  → contrato (routed)
  → executing (orquestrador encerra o comando da sessão)
  → consultas limitadas (parecer → executor)
  → alteração concreta
  → verifying
  → done | blocked
```

Estados: `intake` → `routed` → `executing` → `verifying` → `done` | `blocked`.

Após `executing`, **não** há retorno a `routed`.

## Contrato

Schema: [`schemas/arah-harness/execution-contract.schema.yaml`](../schemas/arah-harness/execution-contract.schema.yaml)

Ledger (quente, gitignored):

```text
.arah/local/execution/
  active/<task-id>.yaml
  completed/<task-id>.yaml
  blocked/<task-id>.yaml
  <task-id>/BRIEFING.md
  <task-id>/consultations/*.yaml
```

## Papéis

| Papel | Quem | Pode |
|-------|------|------|
| Router | orchestrator | classificar, resolver coreografia, criar contrato |
| Executor | um agente `can_execute` | alterar arquivos, evidenciar, concluir/bloquear |
| Consultant | domain / advisors | parecer estruturado ao executor |
| Reviewer | qa, pr-steward, security | revisar; não redefinir executor |

Parecer: [`consultation-result.schema.yaml`](../schemas/arah-harness/consultation-result.schema.yaml). Consultor **não** chama outro consultor.

## Classes de trabalho

| Classe | Consultas | Handoffs | Spec |
|--------|-----------|----------|------|
| trivial | 0 | 0 | não |
| standard | 1 | 1 | leve |
| architectural | 2 | 2 | completa |
| release | 2 | 1 | + human gate |

## Evidência e bloqueio

- `intent_type: execution` + `done` ⇒ evidência concreta (arquivo/comando/teste/artefato). Análise sozinha **não** fecha.
- `blocked` ⇒ `blocking_reason` específico (uma dependência ou decisão humana).

## Limites

Configuráveis em `arah.config.yaml` → `execution_control.limits`. Ao atingir o limite: executor prossegue com o melhor disponível **ou** bloqueia com pergunta humana — nunca “mais uma análise genérica”.

## CLI

```powershell
arah task create -Objective "…" -Area backend -Class standard
arah task status -TaskId task-…
arah task validate -TaskId task-…
arah task complete -TaskId task-… -Evidence "src/x.ts updated; tests passed"
arah task block -TaskId task-… -Reason "Credencial X ausente"
```

Runtime: `./scripts/agents/execute-task.ps1` · Validator: `./scripts/harness/validate-execution-contract.ps1`

## Cursor / Claude Code

Rule: [`.cursor/rules/arah-execution-control.mdc`](../.cursor/rules/arah-execution-control.mdc)

O runtime **não** simula LLM: emite contrato + briefing; o agente executor aplica a mudança e chama `task complete|block`.

## Compatibilidade

- Instalações novas: `execution_control.enabled: true`
- Consumidores sem bloco: `regenerate` / `init` migram defaults sem sobrescrever overlays
- `enabled: false` desliga o contrato formal (modo compat)

## Troubleshooting

| Sintoma | Causa provável |
|---------|----------------|
| `exactly_one_primary_executor_required` | regra sem executor elegível |
| `consultant_to_consultant_handoff_forbidden` | consultor tentou acionar outro consultor |
| `completion_evidence_required` | tentou `done` só com análise |
| `invalid_state_transition` | ex.: executing → routed |
| `consultation_limit_exceeded` | prosseguir ou `task block` |

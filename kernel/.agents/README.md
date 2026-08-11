# Agent manifests — kernel ARAH (genérico)

Manifests YAML definem **quem** pode fazer **o quê** no repo. Consumidos por:

- Agentes Cursor (via `AGENTS.md` + `.cursor/rules/arah-execution-control.mdc`)
- `scripts/agents/execute-task.ps1` + `task-control.ps1` (Execution Control Protocol)
- `scripts/agents/choreograph-agents.ps1` (resolução de coreografia)
- CI (`validate-manifests.ps1`)

Organização orientada a **contratos**, não a uma rede livre de conversação:

```text
intenção → roteamento → contrato → executor → consultas limitadas → alteração → verificação → done|blocked
```

Cada manifest operacional/consultivo declara `execution_role` (`can_route` / `can_execute` / `can_consult` / `can_review`).

## Operacionais — fluxo (abrem PR ou operam gates)

| Arquivo | Agente | Papel ECP |
|---------|--------|-----------|
| orchestrator.agent.yaml | Orquestrador | Router — não executa produto |
| planner.agent.yaml | Planner | Executor de planejamento |
| docs-steward.agent.yaml | Docs Steward | Executor de docs |
| backend.agent.yaml | Backend | Executor de API / serviços |
| frontend.agent.yaml | Frontend | Executor de UI |
| qa.agent.yaml | QA / Review | Executor em testes; reviewer em PR |
| pr-steward.agent.yaml | PR Steward | Executor de stewarding de PR |
| spec-steward.agent.yaml | Spec Steward | Executor de specs SDD |
| solutions-architect.agent.yaml | Solutions Architect | Executor de ADRs; consultor em craft |
| release.agent.yaml | Release / DevOps | Executor de CI/CD |
| security.agent.yaml | Security | Reviewer / consultor |

## Consultivos — domínio (parecer estruturado ao executor)

Gerados por `arah domain sync` a partir de `arah.config.yaml` → `.agents/domain/`.  
`can_execute: false` — nunca assumem a tarefa.

## Consultivos — especialistas (profundidade por stack)

Adicionados conforme stack do projeto em `arah.config.yaml` → `.agents/specialists/`.

## Coreografia

`choreography.yaml` mapeia paths → agentes com `execution.primary_executor` e `role: executor|consultant|reviewer|router`.  
Regras genéricas cobrem specs, PR e áreas de código; regras de domínio são mescladas no init.

Ver [`docs/EXECUTION_CONTROL.md`](../docs/EXECUTION_CONTROL.md).

Checklists em `checklists/`; templates de PR/gates em `templates/`.

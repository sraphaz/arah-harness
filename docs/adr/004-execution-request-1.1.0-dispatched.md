# ADR-004 — ExecutionRequest 1.1.0: status `dispatched` e registro do sinal de transporte

- **Status:** Proposed — **decisão arquitetural; requer aprovação humana explícita (merge é o ato de aceite)**
- **Date:** 2026-09-03
- **Deciders:** maintainers (proposta agente; merge humano)
- **Supersedes / relates:** ADR-003 (fronteira Kernel × Control Plane) — este ADR é a
  evolução de contrato prevista lá: "a integração vira contrato explícito".
  Workspace ADR-0011 / spec WS-16 (origem do caso de uso).

## Context

O Surya Labs Workspace implementou o dispatch de ExecutionRequest
(fase 1 do transporte: issue estruturada no repo alvo, criada pela tool MCP
`ws_dispatch_execution_request` após gate humano assinado —
[PR sraphaz/surya-labs-workspace#22](https://github.com/sraphaz/surya-labs-workspace/pull/22),
spec WS-16, ADR-0011). O contrato canônico **ExecutionRequest 1.0.0**
(`schemas/contracts/execution-request-1.0.0.schema.yaml`, vendorizado pelo
Workspace) não tem `dispatched` no enum de `status`; o enum cobre apenas o
ciclo de vida do executor (`pending, accepted, in_progress, completed, failed,
cancelled`).

Corretamente, o Workspace **não editou o schema vendorizado** (a origem
canônica dos contratos é este repo — doc I §6) e registrou a transição
`pending → dispatched` como estado workspace-local (sidecar em
`data/execution/dispatches/<request_id>.yaml` + evento assinado
`execution_request.dispatched` na hash chain), deixando explícito o follow-up:
propor a evolução aqui. Este ADR é esse follow-up.

## Options

| Opção | Coesão da federação | Risco | Valor |
|-------|--------------------|-------|-------|
| **A — 1.1.0 aditiva: `dispatched` no enum + objeto opcional `dispatch`** (arquivo novo ao lado do 1.0.0, imutável) | Alta — o estado de transporte vira canônico, visível a todos os consumidores | Baixo — minor SemVer; 1.0.0 intacto; tolerant reader ignora o campo novo | Alto |
| B — Manter o estado só no sidecar do Workspace | Média — funciona, mas o dado fica invisível ao executor e a qualquer outro consumidor do contrato | Baixo | Baixo — perpetua estado paralelo para uma transição que é do fluxo federado |
| C — Contrato separado "DispatchRecord" | Média | Médio — quinto contrato para um único objeto pequeno; mais superfície de versionamento | Baixo |

## Decision

Adotar a **opção A** — publicar **ExecutionRequest 1.1.0**, evolução ADITIVA:

1. **`dispatched` entra no enum de `status`**, entre `pending` e `accepted`.
   Semântica: o sinal de transporte foi publicado no repo alvo. A transição
   **`pending → dispatched` é efetuada exclusivamente pelo Control Plane
   (Surya Labs Workspace), após gate humano assinado**
   (`execution_request_approved` — Workspace ADR-0011). As transições
   seguintes (`accepted → in_progress → …`) permanecem do executor
   (ARAH Harness no repo alvo).
2. **Objeto opcional `dispatch`** com o essencial que o registro do Workspace
   já grava: `issue_url` (obrigatório dentro do objeto), `issue_number`
   (nullable), `dispatched_by` (actor_id) e `dispatched_at` (obrigatório).
   Nada especulativo: campos de fase 2 do transporte
   (`repository_dispatch`/MCP) só quando a fase 2 existir.
3. **Imutabilidade por versão** (regra existente do diretório): arquivo novo
   `execution-request-1.1.0.schema.yaml` ao lado do 1.0.0, com `$id` e
   `schema_version` próprios; exemplo novo em `examples/`, validado pelo job
   `contracts-validate`.
4. **CanonicalEvent 1.0.0 não muda**: o evento `execution_request.dispatched`
   já cabe no envelope existente (`type` casa com o padrão
   `^(kernel|workspace|shaping|product)\.[a-z0-9_.]+$`, ex.
   `workspace.execution_request.dispatched`; payload é tolerant reader).

## Consequences

### Positivas

- O estado de transporte deixa de ser paralelo: quando o Workspace adotar a
  1.1.0, o sidecar pode ser aposentado e `status: dispatched` viaja no próprio
  documento canônico, visível ao executor e a qualquer consumidor.
- Retrocompatibilidade total: documentos 1.0.0 continuam validando contra o
  schema 1.0.0 (imutável); consumidores tolerant-reader não quebram com o
  objeto `dispatch`.
- O gate humano continua obrigatório e documentado no próprio contrato (quem
  pode efetuar a transição está na descrição do enum).

### Negativas / trade-offs

- Convivência de duas versões até os emissores migrarem (custo esperado da
  regra de SemVer por schema).
- O objeto `dispatch` assume o transporte fase 1 (issue); a fase 2 exigirá
  nova minor.

### Follow-up esperado (fora deste repo)

- Workspace: vendorizar a 1.1.0, emitir requests `schema_version: "1.1.0"` e
  migrar do sidecar `data/execution/dispatches/` para o status canônico
  (registrado como follow-up no ADR-0011 de lá).

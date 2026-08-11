# `schemas/contracts/` — contratos da federação Surya

Os quatro contratos que ligam os planos do ecossistema (Shaping → Control Plane →
Engineering Kernel) **sem banco compartilhado mutável e sem cross-repository imports**.
Formato: **JSON Schema draft 2020-12 escrito em YAML** (doc I do programa de evolução).

| Contrato | Arquivo | Versão | Emitido por | Consumido por |
|---|---|---|---|---|
| ProjectManifest | `project-manifest-2.0.0.schema.yaml` | 2.0.0 | Surya Labs Workspace | todos |
| ExecutionRequest | `execution-request-1.0.0.schema.yaml` | 1.0.0 | Workspace | ARAH Harness (repo alvo) |
| ExecutionEvidence | `execution-evidence-1.0.0.schema.yaml` | 1.0.0 | ARAH Harness (repo alvo) | Workspace |
| CanonicalEvent | `canonical-event-1.0.0.schema.yaml` | 1.0.0 | qualquer plano | qualquer plano |

Exemplos válidos (validados em CI pelo job `contracts-validate`): [`examples/`](./examples/).

## Por que vivem aqui

O kernel já é o ativo versionado, portátil e instalado nos repositórios do
portfólio — os contratos viajam com ele. (Decisão em `P-surya-platform-decision.md`
do programa; alternativa de repo dedicado de schemas rejeitada por ora.)

## Regras de versionamento

- **SemVer por schema**; a versão faz parte do nome do arquivo e do `$id`.
- **Minor** = apenas campos novos opcionais. **Major** = remoção/renomeio, com
  migração documentada e período de convivência.
- **Tolerant reader**: consumidores DEVEM ignorar campos desconhecidos.
- IDs e regras de propagação (`correlation_id`, `project_id`, `demand_id`,
  `spec_id`, `task_id`, `run_id`, `actor_id`): doc I §1 do programa.
  Reconciliação com o envelope do kernel: `trace_id` = ID técnico por operação;
  `correlation_id` = fio de negócio (nullable em uso standalone).

## Changelog de contratos

| Data | Contrato | Versão | Mudança |
|---|---|---|---|
| 2026-08-11 | todos | inicial | Publicação inicial (ProjectManifest 2.0.0; ExecutionRequest, ExecutionEvidence, CanonicalEvent 1.0.0) |

## Validação

Local (requer Python):

```bash
pip install check-jsonschema
check-jsonschema --schemafile schemas/contracts/project-manifest-2.0.0.schema.yaml schemas/contracts/examples/project-manifest-2.0.0.example.yaml
```

Em CI: job `contracts-validate` em `.github/workflows/ci.yml` valida todos os
exemplos contra os schemas correspondentes.

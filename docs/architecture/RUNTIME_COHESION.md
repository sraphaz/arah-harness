# Arah Harness 0.5 — Runtime Cohesion

**Status:** proposed  
**ADR:** [ADR-002](../adr/002-runtime-cohesion-0.5.md)  
**Spec-Id:** `arah-runtime-cohesion`  
**Harness atual:** 0.4.4 · **Alvo:** 0.5.0

---

## 1. Problema

O Arah Harness já tem método e superfície ricos. O maior risco não é
capacidade insuficiente — é a **superfície crescer mais rápido que o runtime**.

Hoje:

| Sintoma | Efeito |
|---------|--------|
| Regras em docs + YAML + scripts PS | Difícil provar as mesmas invariantes em todas as superfícies |
| CLI Go parcial (H-07) | Duas implementações; writes ainda só em PowerShell |
| `.agents/` e `kernel/.agents/` (e análogos) | Drift e testes de equivalência |
| Estado quente só em arquivos | Concorrência frágil com CLI + MCP + extensão + futuro `arahd` |
| Sem envelope de erro estável | Agentes parseiam texto; CI/MCP/CLI divergem |
| Agent Graph ∪ Knowledge Graph | Falta o **Evidence Graph** (prova da entrega) |

**Norte:** transformar artefatos competentes em um **runtime de governança tipado,
transacional, explicável e acessível por contratos estáveis**.

---

## 2. O que absorver de [rafaelnicolett/kern](https://github.com/rafaelnicolett/kern)

**Não** copiar RAG genérico nem ontologia aprendida por LLM (produto kern).

Absorver decisões de engenharia (ADRs kern 0001/0004/0007):

| Princípio (kern) | Aplicação no Arah Harness |
|------------------|---------------------------|
| Hexagonal | Domínio / casos de uso / portas / adapters (`internal/core`) |
| Binário local coeso | CLI Go como runtime canônico de Execution Control |
| MCP como contrato | `arah mcp serve` — interface estável para agentes |
| Estado isolado por projeto | `.arah/local/` por repositório |
| Resposta com evidência | Envelope JSON + códigos `BOUNDED.CODE` (mais rígido que kern) |
| ADRs honestos | Contratos alinhados ao código real |

---

## 3. Arquitetura alvo

```text
                 ┌─────────────────────────────────────────┐
  CLI Go ───────►│                                         │
  MCP serve ────►│   adapters (in)                         │
  (CI / ext) ───►│                                         │
                 └──────────────────┬──────────────────────┘
                                    ▼
                 ┌─────────────────────────────────────────┐
                 │              arah-core                    │
                 │  Task · Run · Agent · Skill · Policy      │
                 │  Gate · Approval · Evidence · Event       │
                 │  ExecutionContract                        │
                 │  plan → validate → apply                  │
                 └──────────────────┬──────────────────────┘
                                    ▼
                 ┌─────────────────────────────────────────┐
                 │  ports: StateStore · EventStore · VCS     │
                 │  PR · PolicyEngine · KnowledgeProvider    │
                 │  TelemetrySink · ModelProvider            │
                 └──────────────────┬──────────────────────┘
                                    ▼
        filesystem │ SQLite │ Git │ GitHub │ Graphify │ OTel │ …
```

### 3.1 `arah-core`

O core **não** conhece PowerShell, GitHub, Cursor, Codex ou Graphify.
Contém modelo + invariantes tipadas:

- exatamente um `primary_executor` por tarefa executável;
- após `executing`, não retorna a `routed`;
- `done` exige evidência; `blocked` exige causa concreta;
- consultor não executa alteração;
- ações acima da autonomia exigem policy;
- merge / publicação / destrutivo exigem gate humano.

O [Execution Control Protocol](../EXECUTION_CONTROL.md) já descreve o domínio;
0.5 o torna **código central** usado por todas as superfícies.

### 3.2 Portas (mínimo)

```text
StateStore · EventStore · RepositoryScanner · VersionControl
PullRequestProvider · PolicyEngine · TelemetrySink
KnowledgeProvider · ModelProvider
```

Graphify deixa de ser exceção em scripts e passa a ser um `KnowledgeProvider`.

### 3.3 StateStore

Modelo hot/cold permanece ([STATE_MODEL.md](../STATE_MODEL.md)):

| Camada | 0.5 |
|--------|-----|
| Config / manifests / specs / approvals | Git (inalterado) |
| Estado quente | SQLite local WAL — `.arah/local/runtime.db` |
| Eventos | Append-only lógico via `EventStore` |
| Evidência fria | `docs/_meta/runs/` |

Acesso exclusivo por portas; adapter filesystem permanece na migração.

### 3.4 MCP (primeira classe)

Ferramentas de leitura (exemplos): `arah_get_task`, `arah_explain_route`,
`arah_get_evidence`, `arah_get_pending_gates`, `arah_get_capabilities`.

Mutação governada: `arah_create_task`, `arah_complete_task`, `arah_block_task`,
`arah_request_approval`, `arah_submit_consultation`.

MCP e CLI chamam os **mesmos** casos de uso. Sem shell genérico remoto.

### 3.5 Contratos de saída

Todo comando aceita `--json` com envelope estável:

```json
{
  "ok": false,
  "code": "EXECUTION.COMPLETION_EVIDENCE_REQUIRED",
  "message": "A tarefa não pode ser concluída sem evidência.",
  "trace_id": "01K...",
  "details": { "task_id": "task-123" },
  "remediation": ["Informe arquivos alterados", "Registre os testes executados"]
}
```

### 3.6 Evidence Graph

Grafo **determinístico** derivado só de schemas Arah (sem LLM como fonte):

**Entidades:** spec, task, run, agent, skill, gate, approval, PR, arquivo,
capability, domínio.

**Relações:** `depends_on`, `covers`, `assigned_to`, `consulted`, `invokes`,
`validated_by`, `blocked_by`, `approved_by`, `produced`, `evidenced_by`,
`implements`, `supersedes`.

| Grafo | Pergunta |
|-------|----------|
| Agent Graph | Quem colabora com quem? |
| Knowledge Graph | O que o repositório contém? |
| **Evidence Graph** | Como a entrega foi provada? |

Busca textual/vetorial futura = projeção, nunca fonte de verdade.

### 3.7 Observabilidade por Run

Identidade comum: `task_id`, `run_id`, `correlation_id`, `agent_id`,
`session_id`, `trace_id`. Timeline consultável do create → done/blocked.
OpenTelemetry como adapter opcional; default 100% local.

### 3.8 Kernel gerado

```text
fonte canônica → validação → geração do pacote → manifest de hashes → install
```

Preferência cumprida na fase 2: `go:embed` de `internal/kernel/payload/kernel.zip`.
Hoje: fonte canônica na raiz → `arah kernel sync` → `kernel/` + manifest SHA-256 +
payload zip → `arah kernel verify` / CI; `arah kernel install` extrai o embed.
`kernel/` no release **não** é editado à mão.



**Princípio:** se dois arquivos precisam permanecer iguais, um deles não
deveria ser fonte.

---

## 4. Migração PowerShell → Go (estrangulamento)

1. `arah-core` em Go.
2. CLI Go executa casos de uso.
3. Scripts PS tornam-se wrappers de compatibilidade.
4. Fluxos de escrita migram progressivamente.
5. Após janela de compatibilidade, PS deixa de ser implementação canônica
   (pode permanecer adapter Windows corporativo).

Prioridade de comandos no core/CLI:

`install` · `update` · `doctor` · `sync-check` · `resolve` ·
`task create|status|complete|block` · `policy check` · `graph export` ·
`evidence explain` · `mcp serve`

---

## 5. Conformance suite

Fixtures: repo vazio, web app, API, monorepo, harness antigo, drift,
config inválida, tarefa bloqueada, gate pendente.

Provas: install idempotente; update não destrutivo; migração de versões;
mesma decisão CLI≡MCP; transições válidas; policy determinística; geração
reproduzível; error codes estáveis; recuperação após interrupção.

Dogfooding mínimo: `arah-harness`, `sraphaz/arah`, consumidores aplicáveis.

---

## 6. Escopo por prioridade

### P0 — fundação

- `arah-core` + modelo tipado execução/policy
- Pipeline único `plan → validate → apply`
- Remoção de fontes duplicadas do kernel (geração)
- StateStore + migração versionada
- Contratos JSON + catálogo de erros

### P1 — interface agentic

- MCP sobre os mesmos casos de uso
- Evidence Graph determinístico
- Timeline unificada task/run
- Suite de conformidade + fixtures

### P2 — depois de 0.5 estável

- Daemon `arahd` (H-08)
- Knowledge providers plugáveis
- Full-text local; semântica opcional
- Control plane multi-repositório

### Explicitamente fora de 0.5

Novos tipos de agentes · ontologia LLM · RAG próprio · dashboards antes de
contratos estáveis · swarm irrestrito · expandir Graphify antes da porta
`KnowledgeProvider` · novas metáforas TechOrganism sem representação no core.

---

## 7. Definition of Done (versão 0.5.0)

1. CLI e MCP produzem a mesma decisão para o mesmo input.
2. Toda mutação possui dry-run, diff e idempotência.
3. Nenhuma cópia distribuída é editada manualmente.
4. Todas as transições do Execution Control têm testes no core.
5. Toda decisão de policy é explicável.
6. Toda conclusão possui evidência verificável.
7. Upgrade e rollback de schema do StateStore são testados
   (`internal/adapters/sqlitestore` — `SchemaVersion` / `RollbackTo`).
8. Repo sem daemon ou provider semântico continua funcionando.
9. Nenhum campo governado é interpretado por LLM.
10. Os próprios repositórios Arah passam na conformance suite.

---

## 8. Leitura final

O Arah Harness já tem visão de produto mais forte que um kernel genérico de
RAG/ontologia. O que falta não é incorporar outro produto — é alcançar no
**runtime** a mesma clareza que o método já possui.

0.5 transforma o harness de conjunto de artefatos e automações em runtime de
governança. Isso vale mais do que qualquer feature agentic adicional.

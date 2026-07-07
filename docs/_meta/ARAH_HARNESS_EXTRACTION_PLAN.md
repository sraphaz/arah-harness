# Ara Harness — Plano de extração (do Arah)

> Status: rascunho para revisão humana · Versão 0.1 · 2026-07-06
> Repo destino: `arah-harness` (novo repo; extraído de `arah`)
> Objetivo: extrair do Arah a camada de governança de repositório como produto reutilizável — instalável em qualquer repo da Surya Labs ou de terceiros — sem carregar o domínio territorial do Arah.

---

## 1. Tese

O Arah acumulou uma prática de repositório governado: specs antes de trabalho, agentes com escopo declarado, PR Steward, CI de validação, ADRs e Definition of Done. Isso é valor **genérico**. A extração separa:

- **Genérico → `arah-harness`:** mecânica de governança (templates, schemas, profiles, scripts, CI).
- **Específico → permanece no `arah`:** domínio territorial (território, comunidades, araponga.app), conteúdo das specs do Arah, agentes de domínio.

**Regra:** o Ara Harness não deve conter nenhum conceito territorial. O Arah passa a ser o **primeiro consumidor** do harness (dogfooding), via migração.

## 2. Estrutura proposta do repo

```
arah-harness/
  templates/                        # copiados/instanciados no repo alvo
    AGENTS.md                       # contrato de operação de agentes no repo
    docs/specs/README.md            # como specs funcionam (spec-before-work)
    docs/governance/DEFINITION_OF_DONE.md
    docs/ops/AGENT_OPERATION.md     # como rodar/limitar agentes
    .agents/                        # definições de agentes (agent.yaml por agente)
    .cursor/rules/                  # regras para Cursor
    .cursor/skills/                 # skills reutilizáveis
    .github/workflows/              # validate-specs, agent-graph-check, pr-checks
  schemas/
    spec.schema.yaml
    agent.schema.yaml
    agent-graph.schema.yaml
    harness-profile.schema.yaml
    pr-check.schema.yaml
  profiles/
    minimal.yaml
    consulting.yaml
    product.yaml
    enterprise.yaml
    open-source.yaml
  scripts/                          # PowerShell primeiro (paridade com Arah); bash espelhado depois
    install-harness.ps1             # instala/atualiza um profile num repo alvo
    validate-specs.ps1
    validate-agent-graph.ps1
    run-harness.ps1                 # roda validações localmente (modo doctor+check)
    export-agent-graph.ps1          # exporta grafo (mermaid/json) p/ docs
    doctor-harness.ps1              # diagnóstico: o que falta, o que divergiu
  docs/
    HARNESS_OVERVIEW.md
    INSTALLATION.md
    PROFILES.md
    SURYA_LABS_USAGE.md             # como a Surya usa (exemplo de consumidor)
    MIGRATION_FROM_ARAH.md
```

## 3. Triagem: genérico × específico × template × profile

| Item hoje no Arah | Destino | Forma |
|---|---|---|
| AGENTS.md (estrutura, guardrails) | harness | template com placeholders (`{{project_name}}`, paths) |
| specs de domínio territorial | arah | permanecem; formato vira `spec.schema.yaml` |
| convenção spec-before-work | harness | template `docs/specs/README.md` + check de CI |
| agentes de domínio (ex.: curadoria territorial) | arah | `.agents/*.yaml` locais, validados pelo schema do harness |
| PR Steward (mecânica de revisão) | harness | agente template + workflow |
| Definition of Done | harness | template por profile (cada profile tem DoD base) |
| workflows de CI | harness | templates parametrizados |
| ADR (formato e pasta) | harness | template `docs/adr/` |
| Agent Graph (quem chama quem, limites) | harness | schema + validador + export |
| regras .cursor específicas do Arah | arah | mantêm; base genérica vira template |

**Critério de decisão:** se remover a palavra "território/Arah" do item e ele continuar fazendo sentido para um repo qualquer → harness. Se não → arah.

## 4. Instalação em um repo novo

```
1. escolher profile (minimal | consulting | product | enterprise | open-source)
2. ./scripts/install-harness.ps1 -Target ../meu-repo -Profile consulting
   → copia templates instanciados (placeholders resolvidos)
   → escreve harness-profile.yaml na raiz do alvo (profile, versão, opções)
   → instala workflows de CI
3. ./scripts/doctor-harness.ps1 -Target ../meu-repo
   → confere integridade; lista o que falta (ex.: specs vazias)
4. commit inicial "chore: install ara-harness (consulting)"
```

Propriedades exigidas:
- **Idempotente:** reinstalar não duplica nem destrói customizações locais (merge por marcadores de bloco gerenciado).
- **Atualizável:** `install -Update` traz nova versão dos blocos gerenciados; diffs fora deles são preservados.
- **Removível:** `install -Uninstall` remove blocos gerenciados; specs e ADRs do projeto ficam.

## 5. Validação e CI

- `validate-specs`: toda spec ativa valida contra `spec.schema.yaml`; PR que muda código coberto por spec exige spec atualizada (spec-before-work).
- `validate-agent-graph`: agentes declarados formam grafo acíclico; todo agente tem escopo, limites e avaliação; nenhum agente sem humano responsável.
- `pr-checks`: conforme `pr-check.schema.yaml` do profile — ex.: bloquear commit direto em `main`, exigir referência a spec/ADR, scan de segredos.
- CI roda os mesmos scripts do local (paridade local/CI).

## 6. Como agentes obedecem paths

`AGENTS.md` + `.agents/<agente>.yaml` declaram por agente: `allowed_paths`, `forbidden_paths`, `max_diff_lines`, ferramentas permitidas, orçamento. O PR Steward valida mecanicamente: PR de agente tocando path fora do escopo → check falha + label `needs-human`. Humanos não são limitados por paths — são responsabilizados pelo log.

## 7. PR Steward

Agente template do harness:
- roda em todo PR: checa escopo (paths), tamanho de diff, presença de spec/ADR quando exigido, DoD do profile;
- comenta um resumo estruturado (o que muda, riscos, checklist);
- **nunca aprova nem faz merge** — aprovação é humana;
- registra sua análise como comentário auditável.

## 8. Definition of Done

Cada profile carrega uma DoD base (template): código revisado por humano; specs sincronizadas; checks verdes; sem segredos; ADR para decisões estruturais; changelog quando aplicável. Profiles fortalecem a DoD (ver [`HARNESS_PROFILES.md`](./HARNESS_PROFILES.md)).

## 9. Migração do Arah

1. Congelar mudanças de governança no Arah durante a extração (janela curta).
2. Extrair templates/schemas/scripts → `arah-harness` v0.1.
3. Instalar `arah-harness` no próprio Arah com profile `product`; apagar duplicatas locais.
4. Divergências viram issues no harness (não patches locais no Arah).
5. Documentar em `MIGRATION_FROM_ARAH.md` o que ficou para trás e por quê.

## 10. Riscos e pendências

- **Risco:** generalização prematura de coisas que só o Arah usa → só extrair o que o segundo consumidor (Surya) precisa; resto espera.
- **Risco:** drift entre harness e repos instalados → blocos gerenciados + `doctor` acusando divergência.
- **Pendência:** scripts em PowerShell atendem o ambiente atual; espelhos bash para CI Linux — priorizar cedo.
- **Pendência:** licença do `arah-harness` (aberto desde o início? — **Decisão** a registrar em ADR; recomendação: sim, MIT/Apache-2.0, alinhado à soberania).
- **Validação externa necessária:** nenhuma neste documento (governança técnica).

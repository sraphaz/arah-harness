import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { isLocale, localePath, type Locale } from "@/lib/i18n";

const dict = {
  en: {
    metaTitle: "Architecture — ARAH Harness",
    eyebrow: "Architecture",
    heroTitle: "Two repositories, one contract: kernel and consumer",
    heroBody:
      "The harness is developed and versioned in its own repository (the product), and installed into consumer repositories as a maintained kernel. Everything the consumer runs is versioned, inspectable and regenerable.",
    layersLabel: "Layered view · authority flows down, evidence flows up",
    layers: [
      {
        name: "Human Intent",
        detail: "goals · decisions · approval · merge",
      },
      {
        name: "AI Environments",
        detail: "Claude Code · Cursor · Copilot · Codex · local models",
      },
      {
        name: "ARAH Harness",
        detail:
          "agents · skills · choreography · policies · gates · signals · graph · audit · evolution",
        accent: true,
      },
      {
        name: "Repository",
        detail: "code · architecture · domain · specs · docs · tests · infra · CI",
      },
      {
        name: "Delivery",
        detail: "pull request · review · merge · deployment",
      },
    ],
    treesHeadline: "Product repository and consumer repository",
    treesBody:
      "The kernel (agents, skills, scripts, schemas, templates) is developed in arah-harness and distributed into consumers by install and regenerate. The consumer keeps its own state: config, signals, audit, organism.",
    productLabel: "PRODUCT · KERNEL SOURCE",
    productHint: "distributed →",
    consumerLabel: "CONSUMER · YOUR REPOSITORY",
    consumerHint: "← installed kernel + local state",
    productTree: [
      ["arah-harness/", ""],
      ["├── cli/", "arah.ps1 entrypoint"],
      ["├── kernel/", ""],
      ["│   ├── .agents/", "agent manifests"],
      ["│   ├── .skills/", "executable skills"],
      ["│   ├── .cursor/", "IDE rules"],
      ["│   └── scripts/", "gates & operations"],
      ["├── schemas/", "contract validation"],
      ["├── templates/", "bootstrap templates"],
      ["├── extension/", "ARAH Live (IDE)"],
      ["├── harness/", "choreography core"],
      ["└── docs/", ""],
    ] as [string, string][],
    consumerTree: [
      ["consumer-repository/", ""],
      ["├── arah.config.yaml", "policies & autonomy"],
      ["├── AGENTS.md", "entry contract"],
      ["├── .agents/", "installed manifests"],
      ["├── .skills/", "installed skills"],
      ["├── .arah/", ""],
      ["│   ├── audit/", "append-only ledger"],
      ["│   ├── bus/", "typed signals"],
      ["│   └── organism/", "organism state"],
      ["├── docs/_meta/", "domains · graph · discovery"],
      ["├── scripts/", "gates (local + CI)"],
      ["└── .github/workflows/", "CI integration"],
    ] as [string, string][],
    contractsHeadline: "Contracts, storage and generation",
    contractsBody:
      "Every mechanism is a file with a schema — no hidden state, no external database.",
    contracts: [
      {
        eyebrow: "CONTRACTS",
        title: "Schema-validated YAML",
        body: "Agent manifests, skill manifests, choreography, domains and config validate against schemas/ on every gate run.",
      },
      {
        eyebrow: "SIGNAL STORAGE",
        title: "Append-only JSONL bus",
        body: "Typed events under .arah/bus/ — readable by any tool, diffable in PRs, portable across models.",
      },
      {
        eyebrow: "AUDIT STORAGE",
        title: "Session ledger",
        body: "Actions, results, autonomy decisions and session telemetry under .arah/audit/.",
      },
      {
        eyebrow: "GRAPH GENERATION",
        title: "Derived, never hand-written",
        body: "arah export-graph derives the agent graph from manifests; a gate fails if the graph drifts from its sources.",
      },
      {
        eyebrow: "IDE INTEGRATION",
        title: "Rules + ARAH Live",
        body: "Installed .cursor/ rules and AGENTS.md bind Cursor, Claude Code and Copilot to the contracts; ARAH Live (experimental) streams bus events into the IDE.",
      },
      {
        eyebrow: "CI INTEGRATION",
        title: "Gates in workflows",
        body: "The same gates run in .github/workflows/: a PR that fails validation cannot be merged.",
      },
    ],
    updateHeadline: "Update lifecycle",
    updateBody:
      "The consumer stays consistent with the kernel through regeneration — never by manual copying.",
    updateSteps: [
      {
        code: "kernel update",
        body: "new agents, skills, gates land in arah-harness",
      },
      {
        code: "arah regenerate",
        body: "kernel synced, discovery re-run, domains synced",
      },
      {
        code: "arah doctor · sync-check",
        body: "installation validated, drift detected",
      },
      {
        code: "pull request",
        body: "the update itself is reviewed and merged by humans",
      },
    ],
    ctaQuote:
      "Every agent needs a role, a scope, a capability and an accountable path to production.",
    ctaFlow: "See how a change flows →",
    ctaDocs: "Read the docs",
    flowHref: "/how-it-works",
    docsHref: "/docs",
  },
  pt: {
    metaTitle: "Arquitetura — ARAH Harness",
    eyebrow: "Arquitetura",
    heroTitle: "Dois repositórios, um contrato: kernel e consumidor",
    heroBody:
      "O harness é desenvolvido e versionado no seu próprio repositório (o produto) e instalado em repositórios consumidores como um kernel mantido. Tudo o que o consumidor executa é versionado, inspecionável e regenerável.",
    layersLabel: "Visão em camadas · autoridade desce, evidência sobe",
    layers: [
      {
        name: "Intenção Humana",
        detail: "objetivos · decisões · aprovação · merge",
      },
      {
        name: "Ambientes de IA",
        detail: "Claude Code · Cursor · Copilot · Codex · modelos locais",
      },
      {
        name: "ARAH Harness",
        detail:
          "agentes · skills · coreografia · políticas · gates · sinais · grafo · auditoria · evolução",
        accent: true,
      },
      {
        name: "Repositório",
        detail: "código · arquitetura · domínio · specs · docs · testes · infra · CI",
      },
      {
        name: "Entrega",
        detail: "pull request · revisão · merge · deployment",
      },
    ],
    treesHeadline: "Repositório do produto e repositório consumidor",
    treesBody:
      "O kernel (agentes, skills, scripts, schemas, templates) é desenvolvido em arah-harness e distribuído aos consumidores por install e regenerate. O consumidor mantém seu próprio estado: config, sinais, auditoria, organismo.",
    productLabel: "PRODUTO · FONTE DO KERNEL",
    productHint: "distribuído →",
    consumerLabel: "CONSUMIDOR · SEU REPOSITÓRIO",
    consumerHint: "← kernel instalado + estado local",
    productTree: [
      ["arah-harness/", ""],
      ["├── cli/", "entrypoint arah.ps1"],
      ["├── kernel/", ""],
      ["│   ├── .agents/", "manifests de agentes"],
      ["│   ├── .skills/", "skills executáveis"],
      ["│   ├── .cursor/", "regras de IDE"],
      ["│   └── scripts/", "gates & operações"],
      ["├── schemas/", "validação de contratos"],
      ["├── templates/", "templates de bootstrap"],
      ["├── extension/", "ARAH Live (IDE)"],
      ["├── harness/", "núcleo da coreografia"],
      ["└── docs/", ""],
    ] as [string, string][],
    consumerTree: [
      ["consumer-repository/", ""],
      ["├── arah.config.yaml", "políticas & autonomia"],
      ["├── AGENTS.md", "contrato de entrada"],
      ["├── .agents/", "manifests instalados"],
      ["├── .skills/", "skills instaladas"],
      ["├── .arah/", ""],
      ["│   ├── audit/", "ledger append-only"],
      ["│   ├── bus/", "sinais tipados"],
      ["│   └── organism/", "estado do organismo"],
      ["├── docs/_meta/", "domínios · grafo · discovery"],
      ["├── scripts/", "gates (local + CI)"],
      ["└── .github/workflows/", "integração com CI"],
    ] as [string, string][],
    contractsHeadline: "Contratos, armazenamento e geração",
    contractsBody:
      "Cada mecanismo é um arquivo com schema — sem estado oculto, sem banco de dados externo.",
    contracts: [
      {
        eyebrow: "CONTRATOS",
        title: "YAML validado por schema",
        body: "Manifests de agentes e skills, coreografia, domínios e config validam contra schemas/ em toda execução de gates.",
      },
      {
        eyebrow: "ARMAZENAMENTO DE SINAIS",
        title: "Bus JSONL append-only",
        body: "Eventos tipados em .arah/bus/ — legíveis por qualquer ferramenta, diffáveis em PRs, portáveis entre modelos.",
      },
      {
        eyebrow: "ARMAZENAMENTO DE AUDITORIA",
        title: "Ledger de sessões",
        body: "Ações, resultados, decisões de autonomia e telemetria de sessão em .arah/audit/.",
      },
      {
        eyebrow: "GERAÇÃO DO GRAFO",
        title: "Derivado, nunca manual",
        body: "arah export-graph deriva o grafo dos manifests; um gate falha se o grafo divergir das fontes.",
      },
      {
        eyebrow: "INTEGRAÇÃO COM IDE",
        title: "Regras + ARAH Live",
        body: "Regras .cursor/ instaladas e AGENTS.md vinculam Cursor, Claude Code e Copilot aos contratos; o ARAH Live (experimental) transmite eventos do bus para o IDE.",
      },
      {
        eyebrow: "INTEGRAÇÃO COM CI",
        title: "Gates em workflows",
        body: "Os mesmos gates rodam em .github/workflows/: um PR que falha na validação não pode ser mesclado.",
      },
    ],
    updateHeadline: "Ciclo de atualização",
    updateBody:
      "O consumidor permanece consistente com o kernel por regeneração — nunca por cópia manual.",
    updateSteps: [
      {
        code: "atualização do kernel",
        body: "novos agentes, skills e gates chegam ao arah-harness",
      },
      {
        code: "arah regenerate",
        body: "kernel sincronizado, discovery reexecutado, domínios sincronizados",
      },
      {
        code: "arah doctor · sync-check",
        body: "instalação validada, drift detectado",
      },
      {
        code: "pull request",
        body: "a própria atualização é revisada e mesclada por humanos",
      },
    ],
    ctaQuote:
      "Todo agente precisa de um papel, um escopo, uma capacidade e um caminho responsável até produção.",
    ctaFlow: "Veja como uma mudança flui →",
    ctaDocs: "Ler as docs",
    flowHref: "/how-it-works",
    docsHref: "/docs",
  },
} as const;

export function generateStaticParams() {
  return [{ locale: "en" }, { locale: "pt" }];
}

export function generateMetadata({
  params,
}: {
  params: { locale: string };
}): Metadata {
  if (!isLocale(params.locale)) return {};
  return { title: dict[params.locale as Locale].metaTitle };
}

function FileTree({
  rows,
  accent = false,
  label,
  hint,
}: {
  rows: [string, string][];
  accent?: boolean;
  label: string;
  hint: string;
}) {
  return (
    <div
      className={`overflow-hidden rounded-panel border ${
        accent
          ? "border-accent bg-gradient-to-b from-[oklch(75%_0.09_200_/_0.05)] to-arah-surface"
          : "border-arah-line bg-arah-surface"
      }`}
    >
      <div className="flex items-center justify-between gap-3 border-b border-arah-hair px-6 py-3.5">
        <span
          className={`font-mono text-[12px] tracking-[0.1em] ${accent ? "text-accent" : "text-arah-faint"}`}
        >
          {label}
        </span>
        <span className="font-mono text-[11px] text-arah-fade">{hint}</span>
      </div>
      <pre className="m-0 overflow-auto px-6 py-5 font-mono text-[13.5px] leading-[1.85] text-arah-muted">
        {rows.map(([path, comment], i) => (
          <div key={i}>
            <span>{path}</span>
            {comment ? (
              <>
                <span>{"    "}</span>
                <span className="text-arah-fade">{`# ${comment}`}</span>
              </>
            ) : null}
          </div>
        ))}
      </pre>
    </div>
  );
}

export default function ArchitecturePage({
  params,
}: {
  params: { locale: string };
}) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const t = dict[locale];

  return (
    <>
      <header className="mx-auto max-w-site px-8 pb-16 pt-24">
        <p className="eyebrow mb-5 text-accent">{t.eyebrow}</p>
        <h1 className="mb-5 max-w-[820px] font-display text-[clamp(32px,5vw,48px)] font-bold leading-[1.1] tracking-[-0.02em] text-balance">
          {t.heroTitle}
        </h1>
        <p className="max-w-[720px] text-[18px] leading-relaxed text-arah-dim text-pretty">
          {t.heroBody}
        </p>
      </header>

      {/* Layer stack */}
      <section className="mx-auto max-w-site px-8 pb-20 pt-6">
        <div className="rounded-panel border border-arah-hair bg-arah-deep p-9">
          <p className="eyebrow mb-6 text-arah-fade">{t.layersLabel}</p>
          <div className="mx-auto flex max-w-[860px] flex-col gap-2">
            {(t.layers as ReadonlyArray<{ name: string; detail: string; accent?: boolean }>).map((layer, i) => (
              <div key={layer.name} className="contents">
                <div
                  className={`flex flex-wrap items-baseline justify-between gap-4 rounded-card border px-6 py-4 ${
                    layer.accent
                      ? "border-accent bg-[oklch(75%_0.09_200_/_0.06)]"
                      : "border-arah-chip"
                  }`}
                >
                  <span
                    className={`text-[16px] font-semibold ${layer.accent ? "text-accent" : ""}`}
                  >
                    {layer.name}
                  </span>
                  <span className="font-mono text-[12px] text-arah-faint">
                    {layer.detail}
                  </span>
                </div>
                {i < t.layers.length - 1 ? (
                  <div className="text-center font-mono text-[13px] text-[#3A4553]">
                    ↓ ↑
                  </div>
                ) : null}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Trees */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[760px]">
            <h2 className="mb-4 font-display text-[clamp(27px,4vw,34px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {t.treesHeadline}
            </h2>
            <p className="text-[16px] leading-relaxed text-arah-dim text-pretty">
              {t.treesBody}
            </p>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <FileTree
              rows={t.productTree}
              label={t.productLabel}
              hint={t.productHint}
            />
            <FileTree
              rows={t.consumerTree}
              accent
              label={t.consumerLabel}
              hint={t.consumerHint}
            />
          </div>
        </div>
      </section>

      {/* Contracts */}
      <section className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[760px]">
            <h2 className="mb-4 font-display text-[clamp(27px,4vw,34px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {t.contractsHeadline}
            </h2>
            <p className="text-[16px] leading-relaxed text-arah-dim">
              {t.contractsBody}
            </p>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {t.contracts.map((c) => (
              <div
                key={c.eyebrow}
                className="rounded-card border border-arah-line bg-arah-surface p-6"
              >
                <div className="mb-2.5 font-mono text-[11px] text-arah-fade">
                  {c.eyebrow}
                </div>
                <div className="mb-2 text-[15.5px] font-semibold">
                  {c.title}
                </div>
                <div className="text-[13.5px] leading-relaxed text-arah-faint">
                  {c.body}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Update flow */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-10 max-w-[760px]">
            <h2 className="mb-4 font-display text-[clamp(27px,4vw,34px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {t.updateHeadline}
            </h2>
            <p className="text-[16px] leading-relaxed text-arah-dim">
              {t.updateBody}
            </p>
          </div>
          <div className="flex flex-wrap items-stretch gap-3 font-mono text-[13px]">
            {t.updateSteps.map((step, i) => (
              <div key={step.code} className="contents">
                <div
                  className={`flex-1 min-w-[170px] rounded-card border px-5 py-4 ${
                    i === t.updateSteps.length - 1
                      ? "border-amber"
                      : "border-arah-line"
                  }`}
                >
                  <div className="mb-2 text-[10.5px] text-arah-fade">
                    {i + 1}
                  </div>
                  <div
                    className={`mb-1.5 ${
                      i === t.updateSteps.length - 1
                        ? "text-amber"
                        : "text-arah-text"
                    }`}
                  >
                    {step.code}
                  </div>
                  <div className="font-display text-[11.5px] leading-relaxed text-arah-faint">
                    {step.body}
                  </div>
                </div>
                {i < t.updateSteps.length - 1 ? (
                  <span className="self-center text-[#3A4553]">→</span>
                ) : null}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-arah-hair px-8 py-20 text-center">
        <div className="mx-auto max-w-[680px]">
          <p className="mb-7 text-[20px] font-medium leading-snug text-[#DDE4EB] text-balance">
            {t.ctaQuote}
          </p>
          <div className="flex flex-wrap justify-center gap-3">
            <Link
              href={localePath(locale, t.flowHref)}
              className="rounded-control bg-accent px-6 py-3 text-[15px] font-semibold"
            >
              {t.ctaFlow}
            </Link>
            <Link
              href={localePath(locale, t.docsHref)}
              className="rounded-control border border-arah-chip px-6 py-3 text-[15px] font-medium hover:border-[#4A5563]"
            >
              {t.ctaDocs}
            </Link>
          </div>
        </div>
      </section>
    </>
  );
}

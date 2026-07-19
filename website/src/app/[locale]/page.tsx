import { Suspense } from "react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { CapabilityMap } from "@/components/home/CapabilityMap";
import { HarnessExplorer } from "@/components/home/HarnessExplorer";
import { HeroPipeline } from "@/components/home/HeroPipeline";
import { LayerPicker } from "@/components/home/LayerPicker";
import { PositioningMatrix, type Matrix } from "@/components/home/PositioningMatrix";
import { QuickStart } from "@/components/home/QuickStart";
import { StatusPanels } from "@/components/home/StatusPanels";
import { loadCapabilities } from "@/lib/capabilities";
import { getHome } from "@/lib/content";
import { isLocale, localePath, type Locale } from "@/lib/i18n";

type Section = {
  eyebrow?: string;
  headline?: string;
  body?: string;
  quote?: string;
};

type HomeCta = { label: string; href: string };

type HomeData = {
  meta: { title: string; version?: string; agents?: number; skills?: number };
  hero: Section & {
    pipeline: Array<{ num: string; name: string; hint?: string }>;
    ctas: HomeCta[];
  };
  shift: Section & {
    tiles: Array<{ kind: string; label: string }>;
  };
  problem: Section & {
    items: Array<{ id: string; title: string; description: string }>;
  };
  inversion: Section;
  whatIs: Section;
  layersSection: Section;
  principles: Section & {
    items: Array<{ slug: string; name: string; description: string }>;
  };
  lifecycle: Section & {
    steps: Array<{ num: string; name: string; slug: string }>;
  };
  capabilityMap: Section;
  positioning: Section;
  notList: Section & { items: string[] };
  status: Section;
  quickstart: Section & {
    commands: string[];
    steps: Array<{ label: string; code: string }>;
  };
  cta: Section;
  parts: Array<{
    slug?: string;
    name: string;
    file: string;
    stage: string;
    definition: string;
    role: string;
    maturity: string;
    example: string;
  }>;
  layers: Array<{
    slug?: string;
    name: string;
    items: string;
    detail: string;
    itemsLong: string;
    order?: number;
  }>;
  areas: Array<{
    slug?: string;
    name: string;
    description: string;
    capabilities: string[];
    order?: number;
  }>;
  matrix: Matrix;
};

const ui = {
  en: {
    pipelineTitle: "One governed change",
    function: "FUNCTION",
    maturity: "MATURITY",
    example: "EXAMPLE",
    dimension: "DIMENSION",
    available: "AVAILABLE",
    experimental: "EXPERIMENTAL",
    planned: "PLANNED",
    copy: "Copy",
    copied: "Copied ✓",
    walkthrough: "Open the interactive walkthrough →",
    architectureLink: "Architecture page",
    docsLink: "documentation",
    agentFirstTitle: "Agent-first",
    agentFirstNote:
      "Authority flows from the conversation. Context lives in the session. Nothing persists.",
    repoFirstTitle: "Repository-first · ARAH",
    repoFirstNote:
      "Authority, scope and validation are versioned contracts in the repository. Everything persists.",
    agentFlow: ["Prompt", "Agent", "Tools", "Code change"],
    repoFlow: [
      "Human intent",
      "Repo context",
      "Domain",
      "Agents",
      "Skills",
      "Gates",
      "Evidence",
      "Pull request",
    ],
    ctaRepo: "View Repository",
    ctaDocs: "Read Documentation",
    ctaIssue: "Open an Issue",
  },
  pt: {
    pipelineTitle: "Uma mudança governada",
    function: "FUNÇÃO",
    maturity: "MATURIDADE",
    example: "EXEMPLO",
    dimension: "DIMENSÃO",
    available: "DISPONÍVEL",
    experimental: "EXPERIMENTAL",
    planned: "PLANEJADO",
    copy: "Copiar",
    copied: "Copiado ✓",
    walkthrough: "Abrir o walkthrough interativo →",
    architectureLink: "página de Arquitetura",
    docsLink: "documentação",
    agentFirstTitle: "Agent-first",
    agentFirstNote:
      "A autoridade flui da conversa. O contexto vive na sessão. Nada persiste.",
    repoFirstTitle: "Repository-first · ARAH",
    repoFirstNote:
      "Autoridade, escopo e validação são contratos versionados no repositório. Tudo persiste.",
    agentFlow: ["Prompt", "Agente", "Ferramentas", "Mudança de código"],
    repoFlow: [
      "Intenção humana",
      "Contexto do repo",
      "Domínio",
      "Agentes",
      "Skills",
      "Gates",
      "Evidência",
      "Pull request",
    ],
    ctaRepo: "Ver repositório",
    ctaDocs: "Ler documentação",
    ctaIssue: "Abrir uma issue",
  },
} as const;

function resolveHref(locale: Locale, href: string): string {
  if (href.startsWith("/docs")) return `/${locale}${href}`;
  if (href.startsWith("/") && !href.startsWith("//")) {
    return localePath(locale, href);
  }
  return href;
}

function FlowChips({
  items,
  accentIndexes = [],
  amberIndexes = [],
}: {
  items: readonly string[];
  accentIndexes?: number[];
  amberIndexes?: number[];
}) {
  return (
    <div className="flex flex-wrap items-center gap-2 font-mono text-[13.5px]">
      {items.map((item, i) => {
        const accent = accentIndexes.includes(i);
        const amber = amberIndexes.includes(i);
        return (
          <span key={`${item}-${i}`} className="contents">
            {i > 0 ? <span className="text-arah-fade">→</span> : null}
            <span
              className={`rounded-control border px-3 py-2 ${
                accent
                  ? "border-accent text-accent"
                  : amber
                    ? "border-amber text-amber"
                    : "border-arah-chip text-arah-text"
              }`}
            >
              {item}
            </span>
          </span>
        );
      })}
    </div>
  );
}

export function generateStaticParams() {
  return [{ locale: "en" }, { locale: "pt" }];
}

export function generateMetadata({
  params,
}: {
  params: { locale: string };
}): Metadata {
  if (!isLocale(params.locale)) return {};
  const home = getHome(params.locale) as HomeData;
  return { title: home.meta.title };
}

export default function HomePage({ params }: { params: { locale: string } }) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const home = getHome(locale) as HomeData;
  const caps = loadCapabilities();
  const t = ui[locale];

  const statusAvailable =
    caps.available.length > 0
      ? caps.available
      : ((home.status as { buckets?: { available?: string[] } }).buckets
          ?.available ?? []);
  const statusExperimental =
    caps.experimental.length > 0
      ? caps.experimental
      : ((home.status as { buckets?: { experimental?: string[] } }).buckets
          ?.experimental ?? []);
  const statusPlanned =
    caps.planned.length > 0
      ? caps.planned
      : ((home.status as { buckets?: { planned?: string[] } }).buckets
          ?.planned ?? []);

  return (
    <>
      {/* Hero */}
      <header className="mx-auto grid max-w-site items-center gap-16 px-8 pb-20 pt-24 md:grid-cols-2">
        <div className="animate-fadeUp">
          <p className="eyebrow mb-[22px] text-accent">{home.hero.eyebrow}</p>
          <h1 className="mb-6 font-display text-[clamp(34px,5.5vw,56px)] font-bold leading-[1.08] tracking-[-0.02em] text-balance">
            {home.hero.headline}
          </h1>
          <p className="mb-9 max-w-[560px] text-[19px] leading-relaxed text-arah-dim text-pretty">
            {home.hero.body}
          </p>
          <div className="flex flex-wrap gap-3">
            {home.hero.ctas.map((cta, i) => {
              const href = resolveHref(locale, cta.href);
              const external = href.startsWith("http");
              const className =
                i === 0
                  ? "rounded-control bg-accent px-6 py-3 text-[15px] font-semibold"
                  : i === 1
                    ? "rounded-control border border-arah-chip px-6 py-3 text-[15px] font-medium text-arah-text hover:border-[#4A5563]"
                    : "px-3 py-3 font-mono text-[15px] text-arah-dim hover:text-arah-text";
              if (external) {
                return (
                  <a
                    key={cta.label}
                    href={href}
                    className={className}
                    target="_blank"
                    rel="noreferrer"
                  >
                    {cta.label}
                  </a>
                );
              }
              return (
                <Link key={cta.label} href={href} className={className}>
                  {cta.label}
                </Link>
              );
            })}
          </div>
        </div>
        <HeroPipeline
          stages={home.hero.pipeline}
          title={t.pipelineTitle}
        />
      </header>

      {/* Shift */}
      <section className="section-alt section-pad">
        <div className="mx-auto grid max-w-site items-start gap-14 md:grid-cols-2">
          <div>
            <p className="eyebrow mb-4">{home.shift.eyebrow}</p>
            <h2 className="mb-5 font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em] text-balance">
              {home.shift.headline}
            </h2>
            <p className="mb-7 text-[16.5px] leading-relaxed text-arah-dim text-pretty">
              {home.shift.body}
            </p>
            <blockquote className="rounded-r-card border-l-2 border-[var(--arah-accent)] bg-[#0E1319] px-6 py-5 text-[19px] font-medium leading-snug text-[#DDE4EB]">
              {home.shift.quote}
            </blockquote>
          </div>
          <div className="grid grid-cols-2 gap-2.5 pt-12">
            {home.shift.tiles.map((tile) => (
              <div
                key={`${tile.kind}-${tile.label}`}
                className="rounded-card border border-arah-line px-[18px] py-4 text-[14.5px]"
              >
                <span className="mb-1.5 block font-mono text-[11px] text-arah-fade">
                  {tile.kind}
                </span>
                <span>{tile.label}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Problem */}
      <section className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-14 max-w-[720px]">
            <p className="eyebrow mb-4">{home.problem.eyebrow}</p>
            <h2 className="mb-[18px] font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em] text-balance">
              {home.problem.headline}
            </h2>
            <p className="text-lg leading-relaxed text-arah-dim text-pretty">
              {home.problem.body}
            </p>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {home.problem.items.map((item, i) => (
              <div
                key={item.id}
                className="rounded-card border border-arah-line bg-arah-surface p-[22px]"
              >
                <div className="mb-3 font-mono text-[11px] text-bad">
                  P·{String(i + 1).padStart(2, "0")}
                </div>
                <div className="mb-2 text-[15.5px] font-semibold">
                  {item.title}
                </div>
                <div className="text-[13.5px] leading-relaxed text-arah-faint">
                  {item.description}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Inversion */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-14 max-w-[720px]">
            <p className="eyebrow mb-4">{home.inversion.eyebrow}</p>
            <h2 className="mb-[18px] font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.inversion.headline}
            </h2>
            <p className="text-[16.5px] leading-relaxed text-arah-dim text-pretty">
              {home.inversion.body}
            </p>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-panel border border-arah-line bg-arah-surface p-7">
              <p className="mb-5 font-mono text-xs uppercase tracking-[0.12em] text-arah-faint">
                {t.agentFirstTitle}
              </p>
              <FlowChips items={t.agentFlow} />
              <p className="mt-[22px] text-[13.5px] leading-relaxed text-[#7B8694]">
                {t.agentFirstNote}
              </p>
            </div>
            <div className="rounded-panel border border-accent bg-gradient-to-b from-[oklch(75%_0.09_200_/_0.06)] to-transparent p-7">
              <p className="mb-5 font-mono text-xs uppercase tracking-[0.12em] text-accent">
                {t.repoFirstTitle}
              </p>
              <FlowChips
                items={t.repoFlow}
                accentIndexes={[3, 4]}
                amberIndexes={[7]}
              />
              <p className="mt-[22px] text-[13.5px] leading-relaxed text-arah-dim">
                {t.repoFirstNote}
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Harness Explorer */}
      <section className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[760px]">
            <p className="eyebrow mb-4">{home.whatIs.eyebrow}</p>
            <h2 className="mb-[18px] font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.whatIs.headline}
            </h2>
            <p className="text-[16.5px] leading-relaxed text-arah-dim text-pretty">
              {home.whatIs.body}
            </p>
          </div>
          <Suspense
            fallback={
              <div className="h-[420px] rounded-panel border border-arah-line bg-arah-surface" />
            }
          >
            <HarnessExplorer
              parts={home.parts}
              labels={{
                function: t.function,
                maturity: t.maturity,
                example: t.example,
              }}
            />
          </Suspense>
        </div>
      </section>

      {/* Layers */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[720px]">
            <p className="eyebrow mb-4">{home.layersSection.eyebrow}</p>
            <h2 className="mb-[18px] font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.layersSection.headline}
            </h2>
            <p className="text-[16.5px] leading-relaxed text-arah-dim">
              {home.layersSection.body}{" "}
              <Link
                href={localePath(locale, "/architecture")}
                className="text-accent underline decoration-[oklch(80%_0.09_200_/_0.4)] underline-offset-2"
              >
                {t.architectureLink}
              </Link>
            </p>
          </div>
          <LayerPicker layers={home.layers} />
        </div>
      </section>

      {/* Principles */}
      <section className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[720px]">
            <p className="eyebrow mb-4">{home.principles.eyebrow}</p>
            <h2 className="font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.principles.headline}
            </h2>
          </div>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {home.principles.items.map((item) => (
              <div
                key={item.slug}
                className="rounded-card border border-arah-line bg-arah-surface p-6"
              >
                <div className="mb-2.5 text-base font-semibold text-accent">
                  {item.name}
                </div>
                <div className="text-[13.5px] leading-relaxed text-arah-faint">
                  {item.description}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Lifecycle */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 flex flex-wrap items-end justify-between gap-6">
            <div className="max-w-[680px]">
              <p className="eyebrow mb-4">{home.lifecycle.eyebrow}</p>
              <h2 className="font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
                {home.lifecycle.headline}
              </h2>
            </div>
            <Link
              href={localePath(locale, "/how-it-works")}
              className="whitespace-nowrap rounded-control border border-accent px-5 py-2.5 text-[14.5px] font-semibold text-accent hover:bg-accent-tint"
            >
              {t.walkthrough}
            </Link>
          </div>
          <div className="grid grid-cols-2 gap-2.5 font-mono sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
            {home.lifecycle.steps.map((step, i) => {
              const human = i === 10;
              return (
                <Link
                  key={step.slug}
                  href={localePath(locale, `/how-it-works#${step.slug}`)}
                  className={`rounded-card border p-3.5 text-[13px] transition-colors hover:border-[#3A4553] ${
                    human
                      ? "border-amber text-amber"
                      : "border-arah-line text-arah-text"
                  }`}
                >
                  <span className="mb-1.5 block text-[10.5px] text-arah-fade">
                    {step.num}
                  </span>
                  <span>{step.name}</span>
                </Link>
              );
            })}
          </div>
        </div>
      </section>

      {/* Capability map */}
      <section className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[720px]">
            <p className="eyebrow mb-4">{home.capabilityMap.eyebrow}</p>
            <h2 className="font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.capabilityMap.headline}
            </h2>
          </div>
          <CapabilityMap areas={home.areas} />
        </div>
      </section>

      {/* Positioning */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[760px]">
            <p className="eyebrow mb-4">{home.positioning.eyebrow}</p>
            <h2 className="mb-[18px] font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.positioning.headline}
            </h2>
            <p className="text-[16.5px] leading-relaxed text-arah-dim text-pretty">
              {home.positioning.body}
            </p>
          </div>
          <PositioningMatrix
            matrix={home.matrix}
            dimensionLabel={t.dimension}
          />
        </div>
      </section>

      {/* Not list */}
      <section className="section-base section-pad">
        <div className="mx-auto grid max-w-site items-start gap-14 md:grid-cols-2">
          <div>
            <p className="eyebrow mb-4">{home.notList.eyebrow}</p>
            <h2 className="mb-6 font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.notList.headline}
            </h2>
            <p className="mb-6 text-[16.5px] leading-relaxed text-arah-dim">
              {home.notList.body}
            </p>
            <blockquote className="rounded-r-card border-l-2 border-[var(--arah-amber)] bg-[#0E1319] px-6 py-5 text-[17px] font-medium leading-snug text-[#DDE4EB]">
              {home.notList.quote}
            </blockquote>
          </div>
          <div className="grid grid-cols-1 gap-2 pt-14 sm:grid-cols-2">
            {home.notList.items.map((item, i) => (
              <div
                key={item}
                className={`rounded-card border border-[#241C1C] px-4 py-3.5 text-sm text-arah-muted ${
                  i === home.notList.items.length - 1 ? "sm:col-span-2" : ""
                }`}
              >
                <span className="mr-2.5 font-mono text-bad">✕</span>
                <span>{item}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Status */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-12 max-w-[720px]">
            <p className="eyebrow mb-4">{home.status.eyebrow}</p>
            <h2 className="mb-[18px] font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {home.status.headline}
            </h2>
            <p className="text-[16.5px] leading-relaxed text-arah-dim">
              {home.status.body}
            </p>
          </div>
          <StatusPanels
            available={statusAvailable}
            experimental={statusExperimental}
            planned={statusPlanned}
            labels={{
              available: t.available,
              experimental: t.experimental,
              planned: t.planned,
            }}
          />
        </div>
      </section>

      {/* Quick start */}
      <section id="quickstart" className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <QuickStart
            eyebrow={home.quickstart.eyebrow ?? ""}
            headline={home.quickstart.headline ?? ""}
            body={
              <>
                {home.quickstart.body}{" "}
                <Link
                  href={localePath(locale, "/docs")}
                  className="text-accent underline decoration-[oklch(80%_0.09_200_/_0.4)] underline-offset-2"
                >
                  {t.docsLink}
                </Link>
              </>
            }
            commands={home.quickstart.commands}
            steps={home.quickstart.steps}
            copyLabel={t.copy}
            copiedLabel={t.copied}
          />
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-arah-hair bg-gradient-to-b from-arah-bgAlt to-[#080A0D] px-8 py-[110px] text-center">
        <div className="mx-auto max-w-[760px]">
          <h2 className="mb-[18px] font-display text-[clamp(30px,4.5vw,42px)] font-bold leading-[1.15] tracking-[-0.015em] text-balance">
            {home.cta.headline}
          </h2>
          <p className="mb-9 text-[17px] leading-relaxed text-arah-dim text-pretty">
            {home.cta.body}
          </p>
          <div className="flex flex-wrap justify-center gap-3">
            <a
              href="https://github.com/sraphaz/arah-harness"
              className="rounded-control bg-accent px-6 py-3 text-[15px] font-semibold"
              target="_blank"
              rel="noreferrer"
            >
              {t.ctaRepo}
            </a>
            <Link
              href={resolveHref(locale, "/docs")}
              className="rounded-control border border-arah-chip px-6 py-3 text-[15px] font-medium hover:border-[#4A5563]"
            >
              {t.ctaDocs}
            </Link>
            <a
              href="https://github.com/sraphaz/arah-harness/issues"
              className="rounded-control border border-arah-chip px-6 py-3 text-[15px] font-medium hover:border-[#4A5563]"
              target="_blank"
              rel="noreferrer"
            >
              {t.ctaIssue}
            </a>
          </div>
        </div>
      </section>
    </>
  );
}

import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTechOrganism } from "@/lib/content";
import { isLocale, type Locale } from "@/lib/i18n";

const ui = {
  en: {
    cycleLabel: "The cycle · selection is the human step",
    metaphor: "METAPHOR",
    mechanism: "MECHANISM",
    mappingHeadline: "Metaphor and mechanism",
    mappingBody:
      "Each biological metaphor maps to an explicit, versioned mechanism.",
    notHeadline: "What the model does not claim",
    ctaBody:
      "The system proposes. Humans select. Evolution happens in pull requests.",
  },
  pt: {
    cycleLabel: "O ciclo · seleção é o passo humano",
    metaphor: "METÁFORA",
    mechanism: "MECANISMO",
    mappingHeadline: "Metáfora e mecanismo",
    mappingBody:
      "Cada metáfora biológica corresponde a um mecanismo explícito e versionado.",
    notHeadline: "O que o modelo não afirma",
    ctaBody:
      "O sistema propõe. Humanos selecionam. A evolução acontece em pull requests.",
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
  const data = getTechOrganism(params.locale);
  return { title: data.meta.title };
}

const DELAYS = [
  "0s",
  "0.6s",
  "1.2s",
  "1.8s",
  "2.4s",
  "3s",
  "3.6s",
  "4.2s",
];

export default function TechOrganismPage({
  params,
}: {
  params: { locale: string };
}) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const data = getTechOrganism(locale);
  const t = ui[locale];
  const selectionSlug = locale === "pt" ? "selecionar" : "select";

  return (
    <>
      <header className="mx-auto max-w-site px-8 pb-14 pt-24">
        <p className="eyebrow mb-5 text-accent">{data.hero.eyebrow}</p>
        <h1 className="mb-5 max-w-[820px] font-display text-[clamp(32px,5vw,48px)] font-bold leading-[1.1] tracking-[-0.02em] text-balance">
          {data.hero.headline}
        </h1>
        <p className="max-w-[720px] text-[18px] leading-relaxed text-arah-dim text-pretty">
          {data.hero.body}
        </p>
      </header>

      {/* Cycle nodes */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <p className="eyebrow mb-8 text-arah-fade">{t.cycleLabel}</p>
          <div className="flex flex-wrap items-center gap-3 font-mono text-[14px]">
            {data.cycle.nodes.map((node, i) => {
              const isSelection =
                node.slug === selectionSlug || node.slug === "selecionar";
              return (
                <div key={node.slug} className="contents">
                  <span
                    className={`animate-cyclePulse rounded-control border px-5 py-3 ${
                      isSelection
                        ? "border-amber text-amber"
                        : "border-accent text-accent"
                    }`}
                    style={{ animationDelay: DELAYS[i % DELAYS.length] }}
                  >
                    {node.name}
                  </span>
                  {i < data.cycle.nodes.length - 1 ? (
                    <span className="text-arah-fade">→</span>
                  ) : null}
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Mapping */}
      <section className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-10 max-w-[720px]">
            <h2 className="mb-4 font-display text-[clamp(27px,4vw,34px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {t.mappingHeadline}
            </h2>
            <p className="text-[16px] leading-relaxed text-arah-dim">
              {t.mappingBody}
            </p>
          </div>
          <div
            className="grid gap-3"
            style={{
              gridTemplateColumns: "repeat(auto-fit, minmax(240px, 1fr))",
            }}
          >
            {data.mapping.map((item) => {
              const isSelection =
                item.slug === "revisao-humana" || item.slug === "human-review";
              return (
                <div
                  key={item.slug}
                  className={`rounded-card border px-5 py-4 ${
                    isSelection
                      ? "border-amber bg-[oklch(75%_0.09_80_/_0.05)]"
                      : "border-arah-line bg-arah-surface"
                  }`}
                >
                  <div className="mb-1 font-mono text-[11px] text-arah-fade">
                    {t.metaphor}
                  </div>
                  <div className="mb-3 text-[15px] font-semibold">
                    {item.metaphor}
                  </div>
                  <div className="mb-1 font-mono text-[11px] text-arah-fade">
                    {t.mechanism}
                  </div>
                  <div
                    className={`font-mono text-[13.5px] ${isSelection ? "text-amber" : "text-accent"}`}
                  >
                    {item.mechanism}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </section>

      {/* Not-claims */}
      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-8 max-w-[760px]">
            <h2 className="mb-4 font-display text-[clamp(27px,4vw,34px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {t.notHeadline}
            </h2>
            <p className="text-[16px] leading-relaxed text-arah-dim">
              {data.boundaries.body}
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            {data.boundaries.items.map((item) => (
              <div
                key={item}
                className="rounded-card border border-[#241C1C] px-4 py-3 text-sm text-arah-muted"
              >
                <span className="mr-2.5 font-mono text-bad">✕</span>
                <span>{item}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="border-t border-arah-hair px-8 py-20 text-center">
        <div className="mx-auto max-w-[720px]">
          <p className="font-display text-[clamp(20px,3vw,26px)] font-semibold leading-snug text-balance text-[#DDE4EB]">
            {data.cta.headline}
          </p>
        </div>
      </section>
    </>
  );
}

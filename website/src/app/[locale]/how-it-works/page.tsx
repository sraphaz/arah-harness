import { Suspense } from "react";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { LifecycleTimeline } from "@/components/how-it-works/LifecycleTimeline";
import { Walkthrough } from "@/components/how-it-works/Walkthrough";
import { getHowItWorks } from "@/lib/content";
import { isLocale, type Locale } from "@/lib/i18n";

const ui = {
  en: {
    command: "Command",
    input: "Input",
    output: "Output",
    artifacts: "Artifacts",
    prev: "Prev",
    next: "Next",
    step: "Step",
    walkthroughEyebrow: "Walkthrough",
    walkthroughHeadline: "One change through the harness",
    walkthroughBody:
      "Scenario: a change touches backend/payments/**. Step through what the harness does.",
  },
  pt: {
    command: "Comando",
    input: "Entrada",
    output: "Saída",
    artifacts: "Artefatos",
    prev: "Voltar",
    next: "Avançar",
    step: "Passo",
    walkthroughEyebrow: "Walkthrough",
    walkthroughHeadline: "Uma mudança pelo harness",
    walkthroughBody:
      "Cenário: uma mudança toca backend/payments/**. Percorra o que o harness faz.",
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
  const data = getHowItWorks(params.locale);
  return { title: data.meta.title };
}

export default function HowItWorksPage({
  params,
}: {
  params: { locale: string };
}) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const data = getHowItWorks(locale);
  const t = ui[locale];
  const walkthrough = {
    eyebrow: data.walkthrough?.eyebrow ?? t.walkthroughEyebrow,
    headline: data.walkthrough?.headline ?? t.walkthroughHeadline,
    body: data.walkthrough?.body ?? t.walkthroughBody,
  };

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

      <section className="section-alt section-pad">
        <div className="mx-auto max-w-site">
          <Suspense
            fallback={
              <div className="h-[420px] rounded-panel border border-arah-line bg-arah-surface" />
            }
          >
            <LifecycleTimeline
              steps={data.steps}
              labels={{
                command: t.command,
                input: t.input,
                output: t.output,
                artifacts: t.artifacts,
              }}
            />
          </Suspense>
        </div>
      </section>

      <section className="section-base section-pad">
        <div className="mx-auto max-w-site">
          <div className="mb-10 max-w-[760px]">
            <p className="eyebrow mb-4">{walkthrough.eyebrow}</p>
            <h2 className="mb-4 font-display text-[clamp(27px,4vw,34px)] font-bold leading-[1.15] tracking-[-0.015em]">
              {walkthrough.headline}
            </h2>
            <p className="text-[16px] leading-relaxed text-arah-dim text-pretty">
              {walkthrough.body}
            </p>
          </div>
          <Suspense
            fallback={
              <div className="h-[360px] rounded-panel border border-arah-line bg-arah-surface" />
            }
          >
            <Walkthrough
              steps={data.demo}
              labels={{ prev: t.prev, next: t.next, step: t.step }}
            />
          </Suspense>
        </div>
      </section>
    </>
  );
}

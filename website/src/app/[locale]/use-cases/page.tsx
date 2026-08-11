import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { getUseCases } from "@/lib/content";
import { isLocale, localePath, type Locale } from "@/lib/i18n";

const ui = {
  en: {
    ctaHeadline: "Not sure which fits?",
    ctaBody: "Start with the docs — the harness is one install away.",
    ctaLink: "Read the docs →",
  },
  pt: {
    ctaHeadline: "Sem certeza de qual se aplica?",
    ctaBody: "Comece pelas docs — o harness fica a uma instalação de distância.",
    ctaLink: "Ler as docs →",
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
  const data = getUseCases(params.locale);
  return { title: data.meta.title };
}

export default function UseCasesPage({
  params,
}: {
  params: { locale: string };
}) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const data = getUseCases(locale);
  const t = ui[locale];

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
          <div
            className="grid gap-4"
            style={{
              gridTemplateColumns: "repeat(auto-fit, minmax(340px, 1fr))",
            }}
          >
            {data.useCases.map((uc, i) => (
              <article
                key={uc.id}
                className="rounded-panel border border-arah-line bg-arah-surface p-7"
              >
                <div className="mb-4 font-mono text-[11px] text-accent">
                  UC·{String(i + 1).padStart(2, "0")}
                </div>
                <h3 className="mb-3 font-display text-[19px] font-semibold tracking-[-0.01em]">
                  {uc.title}
                </h3>
                <p className="text-[14.5px] leading-relaxed text-arah-faint text-pretty">
                  {uc.description}
                </p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-arah-hair px-8 py-20 text-center">
        <div className="mx-auto max-w-[640px]">
          <h2 className="mb-4 font-display text-[26px] font-bold tracking-[-0.015em]">
            {t.ctaHeadline}
          </h2>
          <p className="mb-6 text-[16px] leading-relaxed text-arah-dim">
            {t.ctaBody}
          </p>
          <Link
            href={localePath(locale, "/docs")}
            className="rounded-control bg-accent px-6 py-3 text-[15px] font-semibold"
          >
            {t.ctaLink}
          </Link>
        </div>
      </section>
    </>
  );
}

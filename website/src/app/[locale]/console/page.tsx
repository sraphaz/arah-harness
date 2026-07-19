import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { LiveConsole } from "@/components/console/LiveConsole";
import { getConsoleMock } from "@/lib/content";
import { isLocale, type Locale } from "@/lib/i18n";

const ui = {
  en: {
    metaTitle: "Live Console — ARAH Harness",
    eyebrow: "Live Console",
    headline: "Observability for a governed repository",
    body: "A read-only observability surface: signal feed, gate results, territories, human selection queue and evolution proposals. This page renders a static mock — no data leaves your machine.",
    experimental: "Experimental",
    kernel: "kernel",
    drift: "drift",
    live: "live",
    readonly: "Read-only mock · demo data from content/console/mock.json",
    signalFeed: "Signal feed",
    gatesPanel: "Gates",
    gatesLastRun: "Last run",
    territories: "Territories",
    queue: "Human selection queue",
    queueEmpty: "Queue empty",
    proposals: "Evolution proposals",
    autonomyMix: "Autonomy mix",
    kpisCells: "Active cells",
    kpisSignals: "Signals · 24h",
    kpisGateRate: "Gate pass rate",
    kpisAwaiting: "Awaiting selection",
    kpisProposals: "Open proposals",
    path: "path",
    health: "health",
    agents: "agents",
    signals: "signals",
    autonomy: "autonomy",
    gates: "gates",
    evidence: "evidence",
    agent: "agent",
  },
  pt: {
    metaTitle: "Live Console — ARAH Harness",
    eyebrow: "Live Console",
    headline: "Observabilidade para um repositório governado",
    body: "Uma superfície de observabilidade somente-leitura: feed de sinais, resultados de gates, territórios, fila de seleção humana e propostas de evolução. Esta página renderiza um mock estático — nenhum dado sai da sua máquina.",
    experimental: "Experimental",
    kernel: "kernel",
    drift: "drift",
    live: "ao vivo",
    readonly: "Mock somente-leitura · dados de demonstração de content/console/mock.json",
    signalFeed: "Feed de sinais",
    gatesPanel: "Gates",
    gatesLastRun: "Última execução",
    territories: "Territórios",
    queue: "Fila de seleção humana",
    queueEmpty: "Fila vazia",
    proposals: "Propostas de evolução",
    autonomyMix: "Distribuição de autonomia",
    kpisCells: "Células ativas",
    kpisSignals: "Sinais · 24h",
    kpisGateRate: "Taxa de aprovação",
    kpisAwaiting: "Aguardando seleção",
    kpisProposals: "Propostas abertas",
    path: "path",
    health: "saúde",
    agents: "agentes",
    signals: "sinais",
    autonomy: "autonomia",
    gates: "gates",
    evidence: "evidência",
    agent: "agente",
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
  return { title: ui[params.locale as Locale].metaTitle };
}

export default function ConsolePage({
  params,
}: {
  params: { locale: string };
}) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const t = ui[locale];
  const mock = getConsoleMock();

  return (
    <>
      <header className="mx-auto max-w-wide px-6 pb-8 pt-16">
        <p className="eyebrow mb-4 text-accent">{t.eyebrow}</p>
        <h1 className="mb-3 max-w-[820px] font-display text-[clamp(28px,4vw,38px)] font-bold leading-[1.1] tracking-[-0.02em] text-balance">
          {t.headline}
        </h1>
        <p className="max-w-[760px] text-[16.5px] leading-relaxed text-arah-dim text-pretty">
          {t.body}
        </p>
      </header>
      <LiveConsole mock={mock} labels={t} />
    </>
  );
}

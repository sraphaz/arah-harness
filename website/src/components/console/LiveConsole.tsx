"use client";

import { useMemo, useState } from "react";
import type {
  ConsoleFeedItem,
  ConsoleFilter,
  ConsoleMock,
  ConsoleRepo,
} from "@/lib/content";

type Labels = {
  experimental: string;
  kernel: string;
  drift: string;
  live: string;
  readonly: string;
  signalFeed: string;
  gatesPanel: string;
  gatesLastRun: string;
  territories: string;
  queue: string;
  queueEmpty: string;
  proposals: string;
  autonomyMix: string;
  kpisCells: string;
  kpisSignals: string;
  kpisGateRate: string;
  kpisAwaiting: string;
  kpisProposals: string;
  path: string;
  health: string;
  agents: string;
  signals: string;
  autonomy: string;
  gates: string;
  evidence: string;
  agent: string;
};

function feedColor(type: string): string {
  if (type.startsWith("consultation")) return "text-accent";
  if (type === "gates.passed") return "text-ok";
  if (type === "gates.failed") return "text-bad";
  if (type.startsWith("evolution")) return "text-[#B49BE0]";
  return "text-amber";
}

export function LiveConsole({
  mock,
  labels,
}: {
  mock: ConsoleMock;
  labels: Labels;
}) {
  const [repoSlug, setRepoSlug] = useState(mock.repos[0]?.slug);
  const [filterKey, setFilterKey] = useState<string>("");

  const repo = useMemo(
    () => mock.repos.find((r) => r.slug === repoSlug) ?? mock.repos[0],
    [mock.repos, repoSlug],
  );

  if (!repo) return null;

  const filteredFeed: ConsoleFeedItem[] = filterKey
    ? repo.feed.filter((f) => f.type.startsWith(filterKey))
    : repo.feed;

  return (
    <div className="mx-auto max-w-wide px-6 py-8">
      {/* Header */}
      <div className="mb-6 flex flex-wrap items-center gap-3">
        <span className="rounded-control border border-amber bg-[oklch(75%_0.09_80_/_0.08)] px-2.5 py-1 font-mono text-[10.5px] uppercase tracking-[0.14em] text-amber">
          {labels.experimental}
        </span>
        <RepoPicker
          repos={mock.repos}
          activeSlug={repo.slug}
          onSelect={setRepoSlug}
        />
        <div className="ml-auto flex flex-wrap items-center gap-2 font-mono text-[11.5px]">
          <Chip label={`${labels.kernel} ${repo.kernel}`} tone="neutral" />
          <Chip
            label={`${labels.drift}: ${repo.drift}`}
            tone={repo.driftOk ? "ok" : "warn"}
          />
          <span className="flex items-center gap-1.5 rounded-control border border-arah-control px-2.5 py-1 text-arah-dim">
            <span className="inline-block h-2 w-2 animate-livePulse rounded-full bg-[oklch(75%_0.1_160)]" />
            {labels.live} · {repo.sync}
          </span>
        </div>
      </div>

      <p className="mb-6 font-mono text-[11.5px] text-arah-fade">
        {labels.readonly}
      </p>

      {/* KPIs */}
      <div className="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
        <Kpi label={labels.kpisCells} value={repo.kpis.cells} />
        <Kpi label={labels.kpisSignals} value={repo.kpis.signals24h} />
        <Kpi
          label={labels.kpisGateRate}
          value={repo.kpis.gateRate}
          tone={repo.kpis.gateOk ? "ok" : "warn"}
        />
        <Kpi
          label={labels.kpisAwaiting}
          value={repo.kpis.awaiting}
          tone={repo.kpis.awaiting > 0 ? "amber" : "neutral"}
        />
        <Kpi label={labels.kpisProposals} value={repo.kpis.proposals} />
      </div>

      {/* Main grid */}
      <div className="grid gap-4 lg:grid-cols-3">
        {/* Signal Feed */}
        <Panel
          title={labels.signalFeed}
          className="lg:col-span-2 lg:row-span-2"
        >
          <FeedFilters
            filters={mock.filters}
            activeKey={filterKey}
            onSelect={setFilterKey}
          />
          <ul className="mt-3 flex max-h-[540px] flex-col overflow-y-auto font-mono text-[12.5px]">
            {filteredFeed.map((f, i) => (
              <li
                key={i}
                className="flex flex-wrap gap-3 border-b border-arah-hair px-1 py-2 last:border-b-0"
              >
                <span className="w-[68px] shrink-0 text-arah-fade">{f.time}</span>
                <span className={`w-[170px] shrink-0 ${feedColor(f.type)}`}>
                  {f.type}
                </span>
                <span className="min-w-[200px] flex-1 text-arah-muted">
                  {f.message}
                </span>
                <span className="text-arah-fade">{f.route}</span>
              </li>
            ))}
          </ul>
        </Panel>

        {/* Gates panel */}
        <Panel title={labels.gatesPanel}>
          <p className="mb-3 font-mono text-[11px] text-arah-fade">
            {labels.gatesLastRun}
          </p>
          <ul className="flex flex-col gap-1.5 font-mono text-[12.5px]">
            {repo.gates.map((g) => {
              const ok = g.status === "ok";
              return (
                <li
                  key={g.slug}
                  className="flex items-center justify-between rounded-control border border-arah-hair px-3 py-1.5"
                >
                  <span className="flex items-center gap-2">
                    <span className={ok ? "text-ok" : "text-bad"}>
                      {ok ? "✓" : "✗"}
                    </span>
                    <span className="text-arah-text">{g.name}</span>
                  </span>
                  <span className="text-arah-fade">{g.duration}</span>
                </li>
              );
            })}
          </ul>
          <p className="mt-3 font-mono text-[11.5px] text-arah-dim">
            {repo.gateSummary}
          </p>
        </Panel>

        {/* Territories */}
        <Panel title={labels.territories}>
          <ul className="flex flex-col gap-2 font-mono text-[12px]">
            {repo.territories.map((tr) => (
              <li
                key={tr.slug}
                className="rounded-control border border-arah-hair px-3 py-2"
              >
                <div className="flex items-center justify-between">
                  <span className="text-[13px] text-arah-text">{tr.name}</span>
                  <span
                    className={`text-[11px] ${tr.health === "saudável" || tr.health === "healthy" ? "text-ok" : "text-amber"}`}
                  >
                    {tr.health}
                  </span>
                </div>
                <div className="mt-1 text-[11px] text-arah-fade">{tr.path}</div>
                <div className="mt-1.5 flex flex-wrap gap-3 text-[11px] text-arah-dim">
                  <span>
                    {labels.agents}: {tr.agents}
                  </span>
                  <span>
                    {labels.signals}: {tr.signals}
                  </span>
                  <span className="text-accent">{tr.autonomy}</span>
                </div>
              </li>
            ))}
          </ul>
        </Panel>

        {/* Selection Queue - amber border */}
        <Panel title={labels.queue} accent="amber">
          {repo.queue.length === 0 ? (
            <p className="font-mono text-[12px] text-arah-fade">
              {labels.queueEmpty}
            </p>
          ) : (
            <ul className="flex flex-col gap-2">
              {repo.queue.map((q) => (
                <li
                  key={q.id}
                  className="rounded-control border border-amber bg-[oklch(75%_0.09_80_/_0.04)] px-3 py-2 font-mono text-[12px]"
                >
                  <div className="flex items-center justify-between">
                    <span className="text-[13px] text-arah-text">{q.title}</span>
                    <span className="text-amber">{q.id}</span>
                  </div>
                  <div className="mt-1 flex flex-wrap gap-3 text-[11px] text-arah-dim">
                    <span>
                      {labels.gates}: {q.gates}
                    </span>
                    <span>
                      {labels.autonomy}: {q.autonomy}
                    </span>
                    <span>
                      {labels.evidence}: {q.evidence}
                    </span>
                    <span className="text-accent">
                      {labels.agent}: {q.agent}
                    </span>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </Panel>

        {/* Proposals + autonomy mix */}
        <Panel title={labels.proposals}>
          {repo.proposals.length === 0 ? (
            <p className="font-mono text-[12px] text-arah-fade">—</p>
          ) : (
            <ul className="mb-5 flex flex-col gap-2">
              {repo.proposals.map((p, i) => (
                <li
                  key={i}
                  className="rounded-control border border-arah-hair px-3 py-2 text-[12.5px]"
                >
                  <div className="mb-1 text-arah-text">{p.title}</div>
                  <div className="font-mono text-[11px] text-arah-fade">
                    {p.evidence}
                  </div>
                </li>
              ))}
            </ul>
          )}
          <div className="mb-2 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
            {labels.autonomyMix}
          </div>
          <div className="flex flex-col gap-2">
            {repo.autonomyMix.map((a) => (
              <div key={a.slug}>
                <div className="mb-1 flex justify-between font-mono text-[11px]">
                  <span className="text-arah-dim">{a.name}</span>
                  <span className="text-arah-muted">{a.percent}%</span>
                </div>
                <div className="h-1.5 overflow-hidden rounded-full bg-arah-hair">
                  <div
                    className="h-full bg-accent"
                    style={{ width: `${a.percent}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </Panel>
      </div>
    </div>
  );
}

function Panel({
  title,
  children,
  className = "",
  accent,
}: {
  title: string;
  children: React.ReactNode;
  className?: string;
  accent?: "amber";
}) {
  return (
    <section
      className={`rounded-panel border ${accent === "amber" ? "border-amber" : "border-arah-line"} bg-arah-surface p-5 ${className}`}
    >
      <h2 className="mb-3 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
        {title}
      </h2>
      {children}
    </section>
  );
}

function Kpi({
  label,
  value,
  tone = "neutral",
}: {
  label: string;
  value: string | number;
  tone?: "neutral" | "ok" | "warn" | "amber";
}) {
  const toneCls =
    tone === "ok"
      ? "text-ok"
      : tone === "warn"
        ? "text-bad"
        : tone === "amber"
          ? "text-amber"
          : "text-arah-text";
  return (
    <div className="rounded-card border border-arah-line bg-arah-surface p-4">
      <div className="mb-1 font-mono text-[10.5px] uppercase tracking-[0.12em] text-arah-fade">
        {label}
      </div>
      <div
        className={`font-display text-[22px] font-semibold tracking-[-0.01em] ${toneCls}`}
      >
        {value}
      </div>
    </div>
  );
}

function Chip({
  label,
  tone,
}: {
  label: string;
  tone: "ok" | "warn" | "neutral";
}) {
  const toneCls =
    tone === "ok"
      ? "border-arah-control text-ok"
      : tone === "warn"
        ? "border-arah-control text-amber"
        : "border-arah-control text-arah-dim";
  return (
    <span className={`rounded-control border px-2.5 py-1 ${toneCls}`}>
      {label}
    </span>
  );
}

function FeedFilters({
  filters,
  activeKey,
  onSelect,
}: {
  filters: ConsoleFilter[];
  activeKey: string;
  onSelect: (k: string) => void;
}) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {filters.map((f) => {
        const active = activeKey === f.key;
        return (
          <button
            key={f.slug}
            type="button"
            onClick={() => onSelect(f.key)}
            className={`rounded-control border px-2.5 py-1 font-mono text-[11px] ${
              active
                ? "border-accent bg-accent-tint text-accent"
                : "border-arah-control text-arah-dim hover:border-[#3A4553] hover:text-arah-text"
            }`}
          >
            {f.name}
          </button>
        );
      })}
    </div>
  );
}

function RepoPicker({
  repos,
  activeSlug,
  onSelect,
}: {
  repos: ConsoleRepo[];
  activeSlug: string;
  onSelect: (slug: string) => void;
}) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {repos.map((r) => {
        const active = r.slug === activeSlug;
        return (
          <button
            key={r.slug}
            type="button"
            onClick={() => onSelect(r.slug)}
            className={`flex items-center gap-2 rounded-control border px-3 py-1.5 font-mono text-[12px] ${
              active
                ? "border-accent bg-accent-tint text-accent"
                : "border-arah-control text-arah-dim hover:border-[#3A4553] hover:text-arah-text"
            }`}
          >
            <span
              className={`inline-block h-2 w-2 rounded-full ${r.driftOk ? "bg-[oklch(75%_0.1_160)]" : "bg-[oklch(75%_0.09_80)]"}`}
            />
            {r.name}
          </button>
        );
      })}
    </div>
  );
}

"use client";

import { useEffect, useMemo, useState } from "react";
import { LiveConsole } from "@/components/console/LiveConsole";
import type { ConsoleMock, ConsoleRepo } from "@/lib/content";
import { fetchLiveBundle, liveApiBase, tryLiveSummary } from "@/lib/live-api";

type Labels = Parameters<typeof LiveConsole>[0]["labels"] & {
  connecting?: string;
  liveMode?: string;
  mockMode?: string;
};

function mapLiveToRepo(bundle: Awaited<ReturnType<typeof fetchLiveBundle>>): ConsoleRepo {
  const s = bundle.summary as {
    repo: string;
    kernel: string;
    drift: string;
    drift_ok: boolean;
    synced_at: string;
    kpis: {
      cells: number;
      signals_24h: number;
      gate_rate: string;
      gate_ok: boolean;
      awaiting: number;
      proposals: number;
    };
  };
  const feed = (bundle.feed.events || []).map(
    (e: {
      ts: string;
      type: string;
      message: string;
      route?: string;
      agent_id?: string;
    }) => ({
      time: typeof e.ts === "string" ? e.ts.slice(11, 19) : "",
      type: e.type,
      message: e.message,
      route: e.route || e.agent_id || "",
    }),
  );
  const gates = (bundle.gates.gates || []).map(
    (g: {
      slug: string;
      name: string;
      status: string;
      duration?: string;
    }) => ({
      slug: g.slug,
      name: g.name,
      status: g.status === "ok" ? "ok" : "fail",
      duration: g.duration || "",
    }),
  );
  const gateSummary =
    typeof bundle.gates.summary === "string" && bundle.gates.summary
      ? bundle.gates.summary
      : `${s.kpis.gate_rate} · live`;

  const territories = (bundle.domains.domains || []).map(
    (d: {
      id?: string;
      name: string;
      path: string;
      health: string;
      agents: number;
      signals_24h?: number;
      signals?: number;
      autonomy: string;
    }) => ({
      slug: d.id || d.name,
      name: d.name,
      path: d.path,
      health: d.health,
      agents: d.agents,
      signals: d.signals_24h ?? d.signals ?? 0,
      autonomy: d.autonomy,
    }),
  );

  const autonomyCounts = new Map<string, number>();
  for (const t of territories) {
    autonomyCounts.set(t.autonomy, (autonomyCounts.get(t.autonomy) || 0) + 1);
  }
  const autonomyTotal = territories.length || 1;
  const autonomyMix =
    autonomyCounts.size > 0
      ? Array.from(autonomyCounts.entries()).map(([slug, n]) => ({
          slug,
          name: slug,
          percent: Math.round((n / autonomyTotal) * 100),
        }))
      : [{ slug: "propose-only", name: "propose-only", percent: 100 }];

  const queue = (bundle.queue.items || []).map(
    (p: {
      number: number;
      title: string;
      url: string;
      author: string;
      draft?: boolean;
    }) => ({
      id: `#${p.number}`,
      title: p.title,
      agent: p.author,
      gates: p.draft ? "draft" : "open",
      autonomy: "propose-only",
      evidence: p.url,
    }),
  );
  const proposals = (bundle.proposals.proposals || []).map(
    (p: { kind?: string; change?: string; rationale?: string }) => ({
      title: p.change || p.kind || "proposal",
      evidence: p.rationale || "",
    }),
  );

  return {
    slug: "local",
    name: s.repo,
    kernel: s.kernel,
    drift: s.drift,
    driftOk: !!s.drift_ok,
    sync: s.synced_at,
    kpis: {
      cells: s.kpis.cells,
      signals24h: s.kpis.signals_24h,
      gateRate: s.kpis.gate_rate,
      gateOk: !!s.kpis.gate_ok,
      awaiting: s.kpis.awaiting,
      proposals: s.kpis.proposals,
    },
    gates,
    gateSummary,
    territories,
    queue,
    proposals,
    autonomyMix,
    feed,
  };
}

export function LiveConsoleApp({
  mock,
  labels,
}: {
  mock: ConsoleMock;
  labels: Labels;
}) {
  const [mode, setMode] = useState<"loading" | "live" | "mock">("loading");
  const [liveMock, setLiveMock] = useState<ConsoleMock | null>(null);
  const base = useMemo(() => liveApiBase(), []);

  useEffect(() => {
    let cancelled = false;
    let timer: number | undefined;
    let ws: WebSocket | undefined;

    async function connect() {
      const summary = await tryLiveSummary(base);
      if (cancelled) return;
      if (!summary) {
        setMode("mock");
        return;
      }
      try {
        const bundle = await fetchLiveBundle(base);
        if (cancelled) return;
        const repo = mapLiveToRepo(bundle);
        setLiveMock({
          ...mock,
          repos: [repo, ...mock.repos.filter((r) => r.slug !== "local")],
        });
        setMode("live");

        const wsUrl = base.replace(/^http/, "ws") + "/events";
        try {
          ws = new WebSocket(wsUrl);
          ws.onmessage = async () => {
            try {
              const b = await fetchLiveBundle(base);
              const r = mapLiveToRepo(b);
              setLiveMock((prev) =>
                prev
                  ? { ...prev, repos: [r, ...prev.repos.filter((x) => x.slug !== "local")] }
                  : prev,
              );
            } catch {
              /* keep last */
            }
          };
        } catch {
          timer = window.setInterval(async () => {
            try {
              const b = await fetchLiveBundle(base);
              const r = mapLiveToRepo(b);
              setLiveMock((prev) =>
                prev
                  ? { ...prev, repos: [r, ...prev.repos.filter((x) => x.slug !== "local")] }
                  : prev,
              );
            } catch {
              /* ignore */
            }
          }, 5000);
        }
      } catch {
        if (!cancelled) setMode("mock");
      }
    }

    void connect();
    return () => {
      cancelled = true;
      if (timer) window.clearInterval(timer);
      ws?.close();
    };
  }, [base, mock]);

  if (mode === "loading") {
    return (
      <p className="mx-auto max-w-wide px-6 pb-8 font-mono text-sm text-arah-dim">
        {labels.connecting || "Connecting to arah-live…"}
        <span className="ml-2 text-arah-fade">({base})</span>
      </p>
    );
  }

  const data = mode === "live" && liveMock ? liveMock : mock;
  const banner =
    mode === "live"
      ? labels.liveMode || `Live · ${base}`
      : labels.mockMode || labels.readonly;

  return (
    <>
      <p className="mx-auto max-w-wide px-6 pb-2 font-mono text-[12px] text-arah-fade">
        {banner}
      </p>
      <LiveConsole
        mock={data}
        labels={{ ...labels, readonly: banner }}
      />
    </>
  );
}

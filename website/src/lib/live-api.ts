export type LiveSummary = {
  repo: string;
  kernel: string;
  drift: string;
  drift_ok: boolean;
  live: boolean;
  synced_at: string;
  kpis: {
    cells: number;
    signals_24h: number;
    gate_rate: string;
    gate_ok: boolean;
    awaiting: number;
    proposals: number;
  };
  errors?: string[];
};

export type LiveEvent = {
  id?: number;
  ts: string;
  type: string;
  message: string;
  route?: string;
  agent_id?: string;
  outcome?: string;
  source?: string;
};

const DEFAULT_API = "http://127.0.0.1:8787";

export function liveApiBase(): string {
  if (typeof window !== "undefined") {
    const q = new URLSearchParams(window.location.search).get("api");
    if (q) return q.replace(/\/$/, "");
  }
  return (process.env.NEXT_PUBLIC_LIVE_API || DEFAULT_API).replace(/\/$/, "");
}

export async function tryLiveSummary(
  base = liveApiBase(),
  ms = 900,
): Promise<LiveSummary | null> {
  const ctrl = new AbortController();
  const t = window.setTimeout(() => ctrl.abort(), ms);
  try {
    const res = await fetch(`${base}/api/summary`, { signal: ctrl.signal });
    if (!res.ok) return null;
    return (await res.json()) as LiveSummary;
  } catch {
    return null;
  } finally {
    window.clearTimeout(t);
  }
}

export async function fetchLiveBundle(base = liveApiBase()) {
  const [summary, feed, gates, domains, queue, proposals] = await Promise.all([
    fetch(`${base}/api/summary`).then((r) => r.json()),
    fetch(`${base}/api/feed`).then((r) => r.json()),
    fetch(`${base}/api/gates`).then((r) => r.json()),
    fetch(`${base}/api/domains`).then((r) => r.json()),
    fetch(`${base}/api/queue`).then((r) => r.json()),
    fetch(`${base}/api/proposals`).then((r) => r.json()),
  ]);
  return { summary, feed, gates, domains, queue, proposals, base } as const;
}

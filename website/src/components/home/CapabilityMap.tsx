"use client";

import { useCallback, useEffect, useState } from "react";

export type CapabilityArea = {
  name: string;
  description: string;
  capabilities: string[];
  slug?: string;
  order?: number;
};

function readHashIndex(count: number): number {
  if (typeof window === "undefined") return 0;
  const match = window.location.hash.match(/^#area-(\d+)$/);
  if (!match) return 0;
  const n = Number.parseInt(match[1], 10);
  if (Number.isNaN(n) || n < 1 || n > count) return 0;
  return n - 1;
}

export function CapabilityMap({ areas }: { areas: CapabilityArea[] }) {
  const [selected, setSelected] = useState(0);

  useEffect(() => {
    setSelected(readHashIndex(areas.length));
    const onHash = () => setSelected(readHashIndex(areas.length));
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, [areas.length]);

  const select = useCallback((index: number) => {
    setSelected(index);
    const hash = `#area-${index + 1}`;
    if (window.location.hash !== hash) {
      window.history.replaceState(null, "", hash);
    }
  }, []);

  const current = areas[selected] ?? areas[0];
  if (!current) return null;

  return (
    <div>
      <div className="mb-4 grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-8">
        {areas.map((area, index) => {
          const active = index === selected;
          return (
            <button
              key={area.slug ?? area.name}
              type="button"
              id={`area-${index + 1}`}
              onClick={() => select(index)}
              className={`rounded-card border px-2 py-3.5 text-[14px] font-semibold transition-colors hover:border-[#3A4553] ${
                active
                  ? "border-accent bg-accent-tint text-accent"
                  : "border-arah-line bg-arah-surface text-arah-muted"
              }`}
            >
              {area.name}
            </button>
          );
        })}
      </div>

      <div className="grid items-start gap-8 rounded-panel border border-arah-line bg-arah-surface p-8 md:grid-cols-2">
        <div>
          <h3 className="mb-2.5 font-display text-2xl font-bold text-accent">
            {current.name}
          </h3>
          <p className="text-[14.5px] leading-relaxed text-arah-dim">
            {current.description}
          </p>
        </div>
        <div className="flex flex-wrap gap-2.5">
          {current.capabilities.map((cap) => (
            <span
              key={cap}
              className="rounded-control border border-arah-chip px-4 py-2.5 font-mono text-[13.5px] text-arah-muted"
            >
              {cap}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

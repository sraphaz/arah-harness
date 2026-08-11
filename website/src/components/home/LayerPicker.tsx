"use client";

import { useCallback, useEffect, useState } from "react";

export type Layer = {
  name: string;
  items: string;
  detail: string;
  itemsLong: string;
  slug?: string;
  order?: number;
};

function readHashIndex(count: number): number {
  if (typeof window === "undefined") return 0;
  const match = window.location.hash.match(/^#layer-(\d+)$/);
  if (!match) return 0;
  const n = Number.parseInt(match[1], 10);
  if (Number.isNaN(n) || n < 1 || n > count) return 0;
  return n - 1;
}

export function LayerPicker({ layers }: { layers: Layer[] }) {
  const [selected, setSelected] = useState(0);

  useEffect(() => {
    setSelected(readHashIndex(layers.length));
    const onHash = () => setSelected(readHashIndex(layers.length));
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, [layers.length]);

  const select = useCallback((index: number) => {
    setSelected(index);
    const hash = `#layer-${index + 1}`;
    if (window.location.hash !== hash) {
      window.history.replaceState(null, "", hash);
    }
  }, []);

  const current = layers[selected] ?? layers[0];
  if (!current) return null;

  return (
    <div className="grid items-start gap-5 md:grid-cols-2">
      <div className="flex flex-col gap-2">
        {layers.map((layer, index) => {
          const active = index === selected;
          return (
            <button
              key={layer.slug ?? layer.name}
              type="button"
              id={`layer-${index + 1}`}
              onClick={() => select(index)}
              className={`flex items-baseline gap-4 rounded-card border px-[22px] py-[18px] text-left transition-colors hover:border-[#3A4553] ${
                active
                  ? "border-accent bg-accent-tint"
                  : "border-arah-line bg-arah-surface"
              }`}
            >
              <span className="w-[26px] shrink-0 font-mono text-[11px] text-arah-fade">
                L{index + 1}
              </span>
              <span
                className={`min-w-[170px] text-[17px] font-semibold ${
                  active ? "text-accent" : "text-arah-text"
                }`}
              >
                {layer.name}
              </span>
              <span className="font-mono text-[13px] text-[#7B8694]">
                {layer.items}
              </span>
            </button>
          );
        })}
      </div>

      <div className="sticky top-[84px] min-h-[300px] rounded-panel border border-arah-line bg-arah-surface p-7">
        <p className="mb-2.5 font-mono text-[11px] uppercase tracking-[0.12em] text-accent">
          Layer {selected + 1}
        </p>
        <h3 className="mb-3 font-display text-[22px] font-bold">{current.name}</h3>
        <p className="mb-[18px] text-[14.5px] leading-relaxed text-arah-dim">
          {current.detail}
        </p>
        <div className="font-mono text-[12.5px] leading-[2] text-[#A9C7D2]">
          {current.itemsLong}
        </div>
      </div>
    </div>
  );
}

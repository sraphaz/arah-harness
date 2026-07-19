"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useCallback, useEffect, useMemo } from "react";
import type { HowItWorksDemoStep } from "@/lib/content";

type Labels = {
  prev: string;
  next: string;
  step: string;
};

export function Walkthrough({
  steps,
  labels,
}: {
  steps: HowItWorksDemoStep[];
  labels: Labels;
}) {
  const params = useSearchParams();
  const router = useRouter();
  const activeSlug = params?.get("demo") ?? steps[0]?.slug;
  const index = useMemo(() => {
    const i = steps.findIndex((s) => s.slug === activeSlug);
    return i < 0 ? 0 : i;
  }, [activeSlug, steps]);
  const active = steps[index];

  const goto = useCallback(
    (i: number) => {
      const wrapped = ((i % steps.length) + steps.length) % steps.length;
      const next = new URLSearchParams(params?.toString() ?? "");
      next.set("demo", steps[wrapped].slug);
      router.replace(`?${next.toString()}`, { scroll: false });
    },
    [params, router, steps],
  );

  useEffect(() => {
    function handler(e: KeyboardEvent) {
      if (e.key === "ArrowLeft") goto(index - 1);
      else if (e.key === "ArrowRight") goto(index + 1);
    }
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [goto, index]);

  if (!active) return null;

  return (
    <div className="rounded-panel border border-arah-line bg-arah-surface p-8">
      <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap gap-2">
          {steps.map((s, i) => (
            <button
              key={s.slug}
              type="button"
              aria-label={`${labels.step} ${i + 1}`}
              onClick={() => goto(i)}
              className={`h-8 w-8 rounded-full border font-mono text-[12px] ${
                i === index
                  ? "border-accent bg-accent-tint text-accent"
                  : "border-arah-line text-arah-dim hover:border-[#3A4553]"
              }`}
            >
              {i + 1}
            </button>
          ))}
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => goto(index - 1)}
            className="rounded-control border border-arah-control px-3.5 py-1.5 font-mono text-[12px] text-arah-dim hover:border-[#3A4553] hover:text-arah-text"
          >
            ← {labels.prev}
          </button>
          <button
            type="button"
            onClick={() => goto(index + 1)}
            className="rounded-control border border-arah-control px-3.5 py-1.5 font-mono text-[12px] text-arah-dim hover:border-[#3A4553] hover:text-arah-text"
          >
            {labels.next} →
          </button>
        </div>
      </div>
      <div className="grid gap-6 md:grid-cols-2">
        <div>
          <p className="eyebrow mb-4 text-accent">{active.phase}</p>
          <h3 className="mb-4 font-display text-[22px] font-bold tracking-[-0.015em]">
            {active.title}
          </h3>
          <p className="text-[15px] leading-relaxed text-arah-dim text-pretty">
            {active.body}
          </p>
        </div>
        <pre className="m-0 overflow-auto rounded-card border border-arah-hair bg-arah-code px-5 py-4 font-mono text-[13px] leading-[1.7] text-arah-muted whitespace-pre-wrap">
          {active.terminal}
        </pre>
      </div>
    </div>
  );
}

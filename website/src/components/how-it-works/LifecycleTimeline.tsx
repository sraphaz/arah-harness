"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useMemo } from "react";
import type { HowItWorksStep } from "@/lib/content";

type Labels = {
  command: string;
  input: string;
  output: string;
  artifacts: string;
};

export function LifecycleTimeline({
  steps,
  labels,
}: {
  steps: HowItWorksStep[];
  labels: Labels;
}) {
  const params = useSearchParams();
  const router = useRouter();
  const activeSlug = params?.get("step") ?? steps[0]?.slug;
  const active = useMemo(
    () => steps.find((s) => s.slug === activeSlug) ?? steps[0],
    [activeSlug, steps],
  );

  function selectStep(slug: string) {
    const next = new URLSearchParams(params?.toString() ?? "");
    next.set("step", slug);
    router.replace(`?${next.toString()}`, { scroll: false });
  }

  if (!active) return null;

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)]">
      <div className="grid grid-cols-2 gap-2 self-start font-mono sm:grid-cols-3 lg:grid-cols-2">
        {steps.map((step) => {
          const isActive = step.slug === active.slug;
          return (
            <button
              key={step.slug}
              type="button"
              onClick={() => selectStep(step.slug)}
              className={`rounded-card border px-3.5 py-3 text-left text-[13px] transition-colors ${
                isActive
                  ? "border-accent bg-accent-tint text-accent"
                  : "border-arah-line text-arah-text hover:border-[#3A4553]"
              }`}
            >
              <span className="mb-1.5 block text-[10.5px] text-arah-fade">
                {step.num}
              </span>
              <span>{step.name}</span>
            </button>
          );
        })}
      </div>
      <div className="rounded-panel border border-arah-line bg-arah-surface p-8">
        <div className="mb-3 font-mono text-[12px] text-arah-fade">
          {active.num}
        </div>
        <h3 className="mb-4 font-display text-[24px] font-bold tracking-[-0.015em]">
          {active.name}
        </h3>
        <p className="mb-7 text-[15.5px] leading-relaxed text-arah-dim">
          {active.description}
        </p>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label={labels.input} text={active.input} />
          <Field label={labels.output} text={active.output} />
        </div>
        {active.command ? (
          <div className="mt-5">
            <div className="mb-2 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
              {labels.command}
            </div>
            <pre className="m-0 whitespace-pre-wrap rounded-card border border-arah-hair bg-arah-code px-4 py-3 font-mono text-[13px] text-arah-muted">
              {active.command}
            </pre>
          </div>
        ) : null}
        {active.artifacts ? (
          <div className="mt-5">
            <div className="mb-2 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
              {labels.artifacts}
            </div>
            <pre className="m-0 whitespace-pre-wrap rounded-card border border-arah-hair bg-arah-code px-4 py-3 font-mono text-[13px] text-arah-muted">
              {active.artifacts}
            </pre>
          </div>
        ) : null}
      </div>
    </div>
  );
}

function Field({ label, text }: { label: string; text?: string }) {
  if (!text) return null;
  return (
    <div>
      <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
        {label}
      </div>
      <p className="text-[14px] leading-relaxed text-arah-muted">{text}</p>
    </div>
  );
}

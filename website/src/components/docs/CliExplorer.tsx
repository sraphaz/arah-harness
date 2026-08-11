"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useMemo } from "react";
import { CopyButton } from "@/components/CopyButton";
import type { CliCommand } from "@/lib/content";

type Labels = {
  syntax: string;
  reads: string;
  writes: string;
  example: string;
  output: string;
  copy: string;
  copied: string;
};

export function CliExplorer({
  commands,
  labels,
}: {
  commands: CliCommand[];
  labels: Labels;
}) {
  const params = useSearchParams();
  const router = useRouter();
  const activeSlug = params?.get("cmd") ?? commands[0]?.slug;
  const active = useMemo(
    () => commands.find((c) => c.slug === activeSlug) ?? commands[0],
    [activeSlug, commands],
  );

  function selectCmd(slug: string) {
    const next = new URLSearchParams(params?.toString() ?? "");
    next.set("cmd", slug);
    router.replace(`?${next.toString()}`, { scroll: false });
  }

  if (!active) return null;

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,240px)_minmax(0,1fr)]">
      <ul className="flex max-h-[520px] flex-col gap-1 overflow-y-auto self-start rounded-panel border border-arah-line bg-arah-surface p-2">
        {commands.map((c) => {
          const isActive = c.slug === active.slug;
          return (
            <li key={c.slug}>
              <button
                type="button"
                onClick={() => selectCmd(c.slug)}
                className={`w-full rounded-control px-3 py-2 text-left font-mono text-[12.5px] ${
                  isActive
                    ? "bg-accent-tint text-accent"
                    : "text-arah-dim hover:bg-arah-hair hover:text-arah-text"
                }`}
              >
                {c.name}
              </button>
            </li>
          );
        })}
      </ul>
      <div className="rounded-panel border border-arah-line bg-arah-surface p-7">
        <h3 className="mb-1 font-display text-[22px] font-bold tracking-[-0.015em]">
          {active.name}
        </h3>
        <p className="mb-5 text-[14.5px] leading-relaxed text-arah-dim">
          {active.description}
        </p>
        <div className="mb-2 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
          {labels.syntax}
        </div>
        <pre className="m-0 mb-5 whitespace-pre-wrap rounded-card border border-arah-hair bg-arah-code px-4 py-3 font-mono text-[13.5px] text-accent">
          {active.syntax}
        </pre>
        <div className="grid gap-4 sm:grid-cols-2">
          {active.reads ? (
            <Field label={labels.reads} text={active.reads} />
          ) : null}
          {active.writes ? (
            <Field label={labels.writes} text={active.writes} />
          ) : null}
        </div>
        {active.example ? (
          <div className="mt-5">
            <div className="mb-2 flex items-center justify-between">
              <span className="font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
                {labels.example} · {labels.output}
              </span>
              <CopyButton
                text={active.example}
                label={labels.copy}
                copiedLabel={labels.copied}
              />
            </div>
            <pre className="m-0 whitespace-pre-wrap rounded-card border border-arah-hair bg-arah-code px-4 py-3 font-mono text-[13px] text-arah-muted">
              <span className="text-arah-fade">$ </span>
              {active.example}
              {active.output ? (
                <>
                  {"\n"}
                  <span className="text-ok">{active.output}</span>
                </>
              ) : null}
            </pre>
          </div>
        ) : null}
      </div>
    </div>
  );
}

function Field({ label, text }: { label: string; text: string }) {
  return (
    <div>
      <div className="mb-1.5 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
        {label}
      </div>
      <p className="font-mono text-[12.5px] leading-relaxed text-arah-muted">
        {text}
      </p>
    </div>
  );
}

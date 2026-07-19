"use client";

import {
  useCallback,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import { useRouter, useSearchParams } from "next/navigation";

export type HarnessPart = {
  name: string;
  file: string;
  stage: string;
  definition: string;
  role: string;
  maturity: string;
  example: string;
  slug?: string;
};

export type HarnessExplorerLabels = {
  function: string;
  maturity: string;
  example: string;
};

function stageTone(stage: string): string {
  const s = stage.toLowerCase();
  if (s === "available" || s === "disponível" || s === "disponivel") {
    return "text-ok";
  }
  if (s === "experimental") {
    return "text-amber";
  }
  return "text-arah-faint";
}

function resolveIndex(parts: HarnessPart[], param: string | null): number {
  if (!param) return 0;
  const bySlug = parts.findIndex((p) => p.slug === param);
  if (bySlug >= 0) return bySlug;
  const asNum = Number.parseInt(param, 10);
  if (!Number.isNaN(asNum) && asNum >= 0 && asNum < parts.length) return asNum;
  if (!Number.isNaN(asNum) && asNum >= 1 && asNum <= parts.length) {
    return asNum - 1;
  }
  return 0;
}

export function HarnessExplorer({
  parts,
  labels,
}: {
  parts: HarnessPart[];
  labels: HarnessExplorerLabels;
}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const listId = useId();
  const buttonRefs = useRef<Array<HTMLButtonElement | null>>([]);
  const partParam = searchParams.get("part");

  const initial = useMemo(
    () => resolveIndex(parts, partParam),
    [parts, partParam],
  );
  const [selected, setSelected] = useState(initial);

  useEffect(() => {
    setSelected(resolveIndex(parts, partParam));
  }, [parts, partParam]);

  const syncUrl = useCallback(
    (index: number) => {
      const part = parts[index];
      if (!part) return;
      const value = part.slug ?? String(index);
      const params = new URLSearchParams(searchParams.toString());
      params.set("part", value);
      const qs = params.toString();
      router.replace(qs ? `?${qs}` : "?", { scroll: false });
    },
    [parts, router, searchParams],
  );

  const select = useCallback(
    (index: number, focus = false) => {
      const next = Math.max(0, Math.min(parts.length - 1, index));
      setSelected(next);
      syncUrl(next);
      if (focus) {
        buttonRefs.current[next]?.focus();
      }
    },
    [parts.length, syncUrl],
  );

  const onKeyDown = (event: React.KeyboardEvent, index: number) => {
    let next: number | null = null;
    if (event.key === "ArrowDown" || event.key === "ArrowRight") {
      next = (index + 1) % parts.length;
    } else if (event.key === "ArrowUp" || event.key === "ArrowLeft") {
      next = (index - 1 + parts.length) % parts.length;
    } else if (event.key === "Home") {
      next = 0;
    } else if (event.key === "End") {
      next = parts.length - 1;
    }
    if (next === null) return;
    event.preventDefault();
    select(next, true);
  };

  const current = parts[selected] ?? parts[0];
  if (!current) return null;

  return (
    <div className="grid items-start gap-5 md:grid-cols-2">
      <div
        role="listbox"
        aria-labelledby={listId}
        className="flex flex-col gap-1.5"
      >
        <span id={listId} className="sr-only">
          Harness parts
        </span>
        {parts.map((part, index) => {
          const active = index === selected;
          const num = String(index + 1).padStart(2, "0");
          return (
            <button
              key={part.slug ?? part.name}
              ref={(el) => {
                buttonRefs.current[index] = el;
              }}
              type="button"
              role="option"
              aria-selected={active}
              tabIndex={active ? 0 : -1}
              onClick={() => select(index)}
              onKeyDown={(e) => onKeyDown(e, index)}
              className={`flex items-center gap-3 rounded-card border px-3.5 py-2.5 text-left text-[14px] transition-colors hover:border-[#3A4553] ${
                active
                  ? "border-accent bg-accent-tint"
                  : "border-arah-line bg-transparent"
              }`}
            >
              <span className="w-5 shrink-0 font-mono text-[11px] text-arah-fade">
                {num}
              </span>
              <span className="flex-1 text-arah-text">{part.name}</span>
              <span
                className={`font-mono text-[10.5px] uppercase tracking-wide ${stageTone(part.stage)}`}
              >
                {part.stage}
              </span>
            </button>
          );
        })}
      </div>

      <div className="sticky top-[84px] min-h-[420px] rounded-panel border border-arah-line bg-arah-surface p-8">
        <p className="mb-2.5 font-mono text-[11.5px] uppercase tracking-[0.12em] text-accent">
          {current.stage} · {current.file}
        </p>
        <h3 className="mb-3.5 font-display text-[26px] font-bold leading-tight">
          {current.name}
        </h3>
        <p className="mb-5 text-[15.5px] leading-relaxed text-arah-muted text-pretty">
          {current.definition}
        </p>
        <div className="mb-5 grid gap-3.5 sm:grid-cols-2">
          <div className="rounded-card border border-arah-panel p-4">
            <div className="mb-2 font-mono text-[10.5px] uppercase tracking-[0.12em] text-arah-fade">
              {labels.function}
            </div>
            <div className="text-[13.5px] leading-relaxed text-arah-dim">
              {current.role}
            </div>
          </div>
          <div className="rounded-card border border-arah-panel p-4">
            <div className="mb-2 font-mono text-[10.5px] uppercase tracking-[0.12em] text-arah-fade">
              {labels.maturity}
            </div>
            <div className="text-[13.5px] leading-relaxed text-arah-dim">
              {current.maturity}
            </div>
          </div>
        </div>
        <div className="mb-2 font-mono text-[10.5px] uppercase tracking-[0.12em] text-arah-fade">
          {labels.example}
        </div>
        <pre className="overflow-auto whitespace-pre-wrap rounded-card border border-arah-panel bg-arah-code p-[18px] font-mono text-[12.5px] leading-relaxed text-[#A9C7D2]">
          {current.example}
        </pre>
      </div>
    </div>
  );
}

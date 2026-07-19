"use client";

import { useState } from "react";

export function CopyButton({
  text,
  label = "Copy",
  copiedLabel = "Copied ✓",
}: {
  text: string;
  label?: string;
  copiedLabel?: string;
}) {
  const [copied, setCopied] = useState(false);

  return (
    <button
      type="button"
      className="rounded-control border border-arah-control bg-transparent px-2.5 py-1 font-mono text-[11px] uppercase tracking-[0.12em] text-arah-dim hover:border-[#3A4553] hover:text-arah-text"
      onClick={async () => {
        try {
          await navigator.clipboard.writeText(text);
          setCopied(true);
          window.setTimeout(() => setCopied(false), 1500);
        } catch {
          /* ignore */
        }
      }}
    >
      {copied ? copiedLabel : label}
    </button>
  );
}

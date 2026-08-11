export type StatusChip = { id?: string; name: string; notes?: string } | string;

function chipLabel(item: StatusChip): string {
  return typeof item === "string" ? item : item.name;
}

function chipKey(item: StatusChip, index: number): string {
  if (typeof item === "string") return `${item}-${index}`;
  return item.id ?? `${item.name}-${index}`;
}

export function StatusPanels({
  available,
  experimental,
  planned,
  labels,
}: {
  available: StatusChip[];
  experimental: StatusChip[];
  planned: StatusChip[];
  labels?: {
    available?: string;
    experimental?: string;
    planned?: string;
  };
}) {
  const panels = [
    {
      key: "available",
      title: labels?.available ?? "AVAILABLE",
      items: available,
      border: "border-[oklch(70%_0.1_160_/_0.35)]",
      dot: "bg-[var(--arah-ok)]",
      titleClass: "text-ok",
      chipClass: "border-arah-control text-arah-muted",
    },
    {
      key: "experimental",
      title: labels?.experimental ?? "EXPERIMENTAL",
      items: experimental,
      border: "border-amber",
      dot: "bg-[var(--arah-amber)]",
      titleClass: "text-amber",
      chipClass: "border-arah-control text-arah-muted",
    },
    {
      key: "planned",
      title: labels?.planned ?? "PLANNED",
      items: planned,
      border: "border-arah-control",
      dot: "bg-arah-fade",
      titleClass: "text-arah-faint",
      chipClass: "border-arah-line text-arah-faint",
    },
  ] as const;

  return (
    <div className="grid items-start gap-3.5 md:grid-cols-3">
      {panels.map((panel) => (
        <div
          key={panel.key}
          className={`rounded-panel border bg-arah-surface p-[26px] ${panel.border}`}
        >
          <div className="mb-[18px] flex items-center gap-2.5">
            <span className={`h-2.5 w-2.5 rounded-full ${panel.dot}`} />
            <span
              className={`font-mono text-xs uppercase tracking-[0.12em] ${panel.titleClass}`}
            >
              {panel.title}
            </span>
          </div>
          <div className="flex flex-wrap gap-1.5 font-mono text-[12.5px]">
            {panel.items.map((item, i) => (
              <span
                key={chipKey(item, i)}
                className={`rounded-control border px-2.5 py-1.5 ${panel.chipClass}`}
              >
                {chipLabel(item)}
              </span>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

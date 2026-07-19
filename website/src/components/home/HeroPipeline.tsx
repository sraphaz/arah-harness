type Stage = { num: string; name: string; hint?: string };

export function HeroPipeline({
  stages,
  title,
}: {
  stages: Stage[];
  title: string;
}) {
  return (
    <div className="animate-fadeUp rounded-panel border border-arah-line bg-arah-deep p-5 md:p-6">
      <p className="eyebrow mb-4 text-accent">{title}</p>
      <div className="relative flex gap-4">
        <div className="relative w-px shrink-0 bg-arah-line">
          <span className="absolute left-1/2 h-2 w-2 -translate-x-1/2 rounded-full bg-[var(--arah-accent)] shadow-[0_0_0_3px_oklch(75%_0.09_200_/_0.2)] animate-flowDot" />
        </div>
        <ol className="flex flex-1 flex-col gap-2">
          {stages.map((s, idx) => {
            const accent = idx === 2;
            const amber = idx === 6;
            return (
              <li
                key={s.num}
                className="rounded-card border px-3 py-2.5"
                style={{
                  borderColor: accent
                    ? "var(--arah-accent-border)"
                    : amber
                      ? "var(--arah-amber-border)"
                      : "#1E2630",
                  background: accent
                    ? "var(--arah-accent-tint)"
                    : amber
                      ? "oklch(75% 0.09 80 / .06)"
                      : "#0D1117",
                  animation: "stagePulse 7s ease-in-out infinite",
                  animationDelay: `${idx * 0.9}s`,
                }}
              >
                <div className="flex items-baseline gap-2">
                  <span
                    className="font-mono text-[11px]"
                    style={{
                      color: accent
                        ? "var(--arah-accent-fg)"
                        : amber
                          ? "var(--arah-amber-fg)"
                          : "#8C97A5",
                    }}
                  >
                    {s.num}
                  </span>
                  <span className="text-[14px] text-arah-text">{s.name}</span>
                </div>
                {s.hint ? (
                  <p className="mt-0.5 font-mono text-[11px] text-arah-fade">{s.hint}</p>
                ) : null}
              </li>
            );
          })}
        </ol>
      </div>
    </div>
  );
}

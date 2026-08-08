type Stage = { num: string; name: string; hint?: string };

/** Delays synced to design (flowDot 7s cycle). */
const STAGE_DELAYS_S = [0, 1, 2, 3, 4.2, 5.2, 6.2] as const;

export function HeroPipeline({
  stages,
  title,
}: {
  stages: Stage[];
  title: string;
}) {
  return (
    <div
      className="hero-pipeline animate-fadeUp relative rounded-[14px] border border-arah-panel bg-gradient-to-b from-[#0D1117] to-[#0A0D11] px-8 py-7"
      style={{ animationDelay: "0.15s" }}
      aria-label={title}
    >
      <p className="mb-[18px] font-mono text-[11.5px] uppercase tracking-[0.12em] text-arah-fade">
        {title}
      </p>
      <div className="relative">
        {/* Rail + flowing dot — absolute so it never shifts layout (CLS-safe) */}
        <div
          className="pointer-events-none absolute bottom-3.5 left-[11px] top-3.5 w-0.5 bg-arah-line"
          aria-hidden
        >
          <span className="absolute left-1/2 h-2 w-2 -translate-x-1/2 rounded-full bg-[oklch(80%_0.11_200)] shadow-[0_0_12px_oklch(80%_0.11_200_/_0.8)] animate-flowDot" />
        </div>
        <ol className="relative flex flex-col gap-2.5">
          {stages.map((s, idx) => {
            const accent = idx === 2;
            const amber = idx === 6;
            const delay = STAGE_DELAYS_S[idx] ?? idx;
            return (
              <li
                key={s.num}
                className="relative z-[1] flex items-center gap-3.5 pr-3.5 first:pt-2.5"
              >
                <span
                  className="w-6 shrink-0 bg-[#0B0E12] text-center font-mono text-[11px]"
                  style={{
                    color: accent
                      ? "oklch(78% 0.09 200)"
                      : amber
                        ? "oklch(78% 0.09 80)"
                        : "#5A6675",
                  }}
                >
                  {s.num}
                </span>
                <div
                  className={`animate-stagePulse min-h-[42px] flex-1 rounded-lg border px-3.5 py-2.5 text-[14px] ${
                    accent
                      ? "border-[oklch(75%_0.09_200_/_0.35)] bg-[oklch(75%_0.09_200_/_0.05)]"
                      : amber
                        ? "border-[oklch(75%_0.09_80_/_0.35)] bg-[oklch(75%_0.09_80_/_0.05)]"
                        : "border-arah-line bg-transparent"
                  }`}
                  style={{ animationDelay: `${delay}s` }}
                >
                  <span
                    className="font-semibold"
                    style={{
                      color: accent
                        ? "oklch(84% 0.07 200)"
                        : amber
                          ? "oklch(84% 0.07 80)"
                          : undefined,
                    }}
                  >
                    {s.name}
                  </span>
                  {s.hint ? (
                    <span className="ml-2 font-mono text-[12.5px] text-arah-fade">
                      {s.hint}
                    </span>
                  ) : null}
                </div>
              </li>
            );
          })}
        </ol>
      </div>
    </div>
  );
}

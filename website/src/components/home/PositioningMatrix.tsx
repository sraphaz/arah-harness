const COLUMN_LABELS: Record<string, string> = {
  "coding-copilots": "Coding copilots",
  "multi-agent-frameworks": "Multi-agent frameworks",
  "spec-driven-kits": "Spec-driven kits",
  "ci-cd-platforms": "CI/CD platforms",
  "observability-tools": "Observability tools",
  "arah-harness": "ARAH Harness",
};

function mark(value: string | number): string {
  if (value === "yes" || value === 2) return "●";
  if (value === "partial" || value === 1) return "◐";
  return "—";
}

function markClass(value: string | number, highlight: boolean): string {
  if (highlight) return "text-ok";
  if (value === "yes" || value === 2) return "text-arah-faint";
  if (value === "partial" || value === 1) return "text-[#6B7684]";
  return "text-[#3A4553]";
}

export type MatrixRow = {
  dimension: string;
  slug?: string;
  values: Record<string, string>;
  scores?: Record<string, number>;
};

export type Matrix = {
  columns: string[];
  rows: MatrixRow[];
};

export function PositioningMatrix({
  matrix,
  dimensionLabel = "DIMENSION",
}: {
  matrix: Matrix;
  dimensionLabel?: string;
}) {
  const { columns, rows } = matrix;

  return (
    <div className="overflow-x-auto rounded-panel border border-arah-line bg-arah-surface">
      <div
        className="grid min-w-[960px] text-[12.5px]"
        style={{
          gridTemplateColumns: `230px repeat(${columns.length}, 1fr)`,
        }}
      >
        <div className="border-b border-arah-panel px-[18px] py-3.5 font-mono text-[10.5px] uppercase tracking-[0.1em] text-arah-fade">
          {dimensionLabel}
        </div>
        {columns.map((col) => {
          const arah = col === "arah-harness";
          return (
            <div
              key={col}
              className={`border-b border-arah-panel px-2.5 py-3.5 text-center ${
                arah
                  ? "bg-accent-tint font-semibold text-accent"
                  : "text-arah-faint"
              }`}
            >
              {COLUMN_LABELS[col] ?? col}
            </div>
          );
        })}

        {rows.map((row) => (
          <div key={row.slug ?? row.dimension} className="contents">
            <div className="border-b border-arah-hair px-[18px] py-2.5 text-arah-muted">
              {row.dimension}
            </div>
            {columns.map((col) => {
              const arah = col === "arah-harness";
              const score = row.scores?.[col];
              const value = score !== undefined ? score : row.values[col] ?? "no";
              return (
                <div
                  key={`${row.slug ?? row.dimension}-${col}`}
                  className={`border-b border-arah-hair px-2.5 py-2.5 text-center ${
                    arah ? "bg-accent-tint" : ""
                  } ${markClass(value, arah)}`}
                >
                  {mark(value)}
                </div>
              );
            })}
          </div>
        ))}
      </div>
    </div>
  );
}

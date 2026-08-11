import { CopyButton } from "@/components/CopyButton";

export type QuickStartStep = {
  label: string;
  code: string;
};

export function QuickStart({
  eyebrow,
  headline,
  body,
  commands,
  steps,
  copyLabel,
  copiedLabel,
}: {
  eyebrow: string;
  headline: string;
  body: React.ReactNode;
  commands: string[];
  steps: QuickStartStep[];
  copyLabel?: string;
  copiedLabel?: string;
}) {
  return (
    <div className="grid items-start gap-16 md:grid-cols-2">
      <div>
        <p className="eyebrow mb-4">{eyebrow}</p>
        <h2 className="mb-5 font-display text-[clamp(28px,4.2vw,38px)] font-bold leading-[1.15] tracking-[-0.015em]">
          {headline}
        </h2>
        <div className="mb-7 text-[16.5px] leading-relaxed text-arah-dim text-pretty">
          {body}
        </div>
        <div className="rounded-card border border-arah-panel bg-arah-surface px-[22px] py-[18px] font-mono text-[13.5px] leading-[2.2] text-[#A9C7D2]">
          {commands.map((cmd) => (
            <div key={cmd}>{cmd}</div>
          ))}
        </div>
      </div>

      <div className="flex flex-col gap-3">
        {steps.map((step) => (
          <div
            key={step.label}
            className="overflow-hidden rounded-card border border-arah-panel bg-arah-code"
          >
            <div className="flex items-center justify-between gap-3 border-b border-arah-hair px-4 py-2.5">
              <span className="font-mono text-[11px] text-arah-fade">
                {step.label}
              </span>
              <CopyButton
                text={step.code}
                label={copyLabel}
                copiedLabel={copiedLabel}
              />
            </div>
            <pre className="overflow-auto px-[18px] py-4 font-mono text-[12.5px] leading-[1.7] text-[#A9C7D2]">
              {step.code}
            </pre>
          </div>
        ))}
      </div>
    </div>
  );
}

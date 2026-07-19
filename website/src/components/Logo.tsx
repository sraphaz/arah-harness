export function Logo({ className = "" }: { className?: string }) {
  return (
    <svg
      className={className}
      width="28"
      height="28"
      viewBox="0 0 32 32"
      fill="none"
      aria-hidden
    >
      <line x1="8" y1="8" x2="16" y2="16" stroke="#5A6675" strokeWidth="1.5" />
      <line x1="24" y1="8" x2="16" y2="16" stroke="#5A6675" strokeWidth="1.5" />
      <line x1="16" y1="16" x2="16" y2="26" stroke="#5A6675" strokeWidth="1.5" />
      <circle cx="8" cy="8" r="3" fill="#0D1117" stroke="#9AA5B1" strokeWidth="1.5" />
      <circle cx="24" cy="8" r="3" fill="#0D1117" stroke="#9AA5B1" strokeWidth="1.5" />
      <circle cx="16" cy="26" r="3" fill="#0D1117" stroke="#9AA5B1" strokeWidth="1.5" />
      <circle cx="16" cy="16" r="4" fill="oklch(78% 0.09 200)" stroke="oklch(84% 0.07 200)" strokeWidth="1.5" />
    </svg>
  );
}

import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        arah: {
          bg: "#0A0C0F",
          bgAlt: "#0B0E12",
          surface: "#0D1117",
          code: "#080A0D",
          deep: "#0C0F14",
          line: "#1E2630",
          panel: "#1A212B",
          hair: "#141A22",
          hair2: "#161C24",
          control: "#232B36",
          chip: "#2A3340",
          text: "#E7ECF1",
          muted: "#B7C0CB",
          dim: "#9AA5B1",
          faint: "#8C97A5",
          fade: "#5A6675",
          min: "#4A5563",
        },
      },
      fontFamily: {
        display: ["var(--font-schibsted)", "system-ui", "sans-serif"],
        mono: ["var(--font-plex-mono)", "ui-monospace", "monospace"],
      },
      maxWidth: {
        site: "1200px",
        wide: "1360px",
      },
      borderRadius: {
        control: "7px",
        card: "10px",
        panel: "13px",
      },
      keyframes: {
        flowDot: {
          "0%": { top: "0%" },
          "100%": { top: "100%" },
        },
        stagePulse: {
          "0%, 100%": { borderColor: "#1E2630" },
          "40%, 60%": { borderColor: "oklch(75% 0.09 200 / .55)" },
        },
        cyclePulse: {
          "0%, 100%": { opacity: "0.4" },
          "50%": { opacity: "1" },
        },
        livePulse: {
          "0%, 100%": { opacity: "1", transform: "scale(1)" },
          "50%": { opacity: "0.45", transform: "scale(0.85)" },
        },
        fadeUp: {
          "0%": { opacity: "0", transform: "translateY(12px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
      },
      animation: {
        flowDot: "flowDot 7s linear infinite",
        stagePulse: "stagePulse 7s ease-in-out infinite",
        cyclePulse: "cyclePulse 7s ease-in-out infinite",
        livePulse: "livePulse 2s ease-in-out infinite",
        fadeUp: "fadeUp 0.6s ease both",
      },
    },
  },
  plugins: [],
};
export default config;

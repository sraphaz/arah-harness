export const locales = ["en", "pt"] as const;
export type Locale = (typeof locales)[number];

export function isLocale(value: string): value is Locale {
  return locales.includes(value as Locale);
}

export function otherLocale(locale: Locale): Locale {
  return locale === "en" ? "pt" : "en";
}

export function localePath(locale: Locale, path = ""): string {
  const clean = path.startsWith("/") ? path : path ? `/${path}` : "";
  return `/${locale}${clean}`;
}

export function swapLocalePath(pathname: string, next: Locale): string {
  const parts = pathname.split("/");
  if (parts.length > 1 && isLocale(parts[1])) {
    parts[1] = next;
    return parts.join("/") || `/${next}`;
  }
  return `/${next}`;
}

export const navCopy = {
  en: {
    home: "Home",
    architecture: "Architecture",
    how: "How It Works",
    organism: "TechOrganism",
    useCases: "Use Cases",
    docs: "Docs",
    console: "Live Console",
    github: "View on GitHub",
    start: "Get Started",
    tagline: "Agents propose. Humans select. The repository remembers.",
  },
  pt: {
    home: "Início",
    architecture: "Arquitetura",
    how: "Como Funciona",
    organism: "TechOrganism",
    useCases: "Casos de Uso",
    docs: "Docs",
    console: "Live Console",
    github: "Ver no GitHub",
    start: "Começar",
    tagline: "Agentes propõem. Humanos selecionam. O repositório lembra.",
  },
} as const;

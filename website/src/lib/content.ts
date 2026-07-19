import { readFileSync } from "fs";
import path from "path";
import type { Locale } from "./i18n";

const root = path.join(process.cwd(), "content");

function readJson<T>(rel: string): T {
  return JSON.parse(readFileSync(path.join(root, rel), "utf8")) as T;
}

export function getHome(locale: Locale) {
  return readJson<Record<string, unknown>>(`home/${locale}.json`);
}

export function getHowItWorks(locale: Locale) {
  return readJson<HowItWorksData>(`how-it-works/${locale}.json`);
}

export function getTechOrganism(locale: Locale) {
  return readJson<TechOrganismData>(`techorganism/${locale}.json`);
}

export function getUseCases(locale: Locale) {
  return readJson<UseCasesData>(`use-cases/${locale}.json`);
}

export function getCli(locale: Locale) {
  return readJson<{ locale: string; version?: string; commands: CliCommand[] }>(
    `cli/${locale}.json`,
  );
}

export function getDocsIndex(locale: Locale) {
  return readJson<DocsIndex>(`docs/${locale}/index.json`);
}

export function getConsoleMock() {
  return readJson<ConsoleMock>("console/mock.json");
}

export type CliCommand = {
  slug: string;
  name: string;
  syntax: string;
  description: string;
  reads?: string;
  writes?: string;
  example?: string;
  output?: string;
};

export type DocsBlock =
  | { type: "heading"; text: string }
  | { type: "paragraph"; text: string }
  | { type: "code"; code: string; label?: string; language?: string }
  | { type: "list"; items: string[] }
  | { type: "cli-reference" };

export type DocsPage = {
  slug: string;
  title: string;
  section: string;
  sectionSlug: string;
  intro?: string;
  blocks?: DocsBlock[];
};

export type DocsNavPage = {
  slug: string;
  title: string;
};

export type DocsNavSection = {
  slug: string;
  name: string;
  pages: DocsNavPage[];
};

export type DocsIndex = {
  locale?: string;
  version?: string;
  nav: DocsNavSection[];
  pages: DocsPage[];
};

export function flattenDocs(index: DocsIndex): DocsPage[] {
  return index.pages;
}

export function findDoc(
  index: DocsIndex,
  sectionSlug: string,
  slug: string,
): DocsPage | undefined {
  return index.pages.find(
    (p) => p.sectionSlug === sectionSlug && p.slug === slug,
  );
}

export type HowItWorksSection = {
  eyebrow?: string;
  headline?: string;
  body?: string;
};

export type HowItWorksStep = {
  slug: string;
  num: string;
  name: string;
  description: string;
  command?: string;
  input?: string;
  output?: string;
  artifacts?: string;
  order?: number;
};

export type HowItWorksDemoStep = {
  slug: string;
  phase: string;
  title: string;
  body: string;
  terminal: string;
  order?: number;
};

export type HowItWorksData = {
  locale: string;
  meta: { title: string; version?: string };
  hero: HowItWorksSection;
  walkthrough?: HowItWorksSection;
  steps: HowItWorksStep[];
  demo: HowItWorksDemoStep[];
};

export type TechOrganismData = {
  locale: string;
  meta: { title: string; version?: string };
  hero: HowItWorksSection;
  cycle: { nodes: Array<{ slug: string; name: string }> };
  mapping: Array<{ metaphor: string; mechanism: string; slug: string }>;
  boundaries: { body: string; items: string[] };
  cta: { headline: string; body: string };
};

export type UseCase = {
  id: string;
  slug: string;
  name: string;
  title: string;
  description: string;
};

export type UseCasesData = {
  locale: string;
  meta: { title: string; version?: string };
  hero: HowItWorksSection;
  useCases: UseCase[];
};

export type ConsoleFilter = { slug: string; name: string; key: string };

export type ConsoleGate = {
  slug: string;
  name: string;
  status: string;
  duration: string;
};

export type ConsoleTerritory = {
  slug: string;
  name: string;
  path: string;
  health: string;
  agents: number;
  signals: number;
  autonomy: string;
};

export type ConsoleQueueItem = {
  id: string;
  title: string;
  gates: string;
  autonomy: string;
  evidence: string;
  agent: string;
};

export type ConsoleProposal = {
  title: string;
  evidence: string;
};

export type ConsoleAutonomy = {
  slug: string;
  name: string;
  percent: number;
};

export type ConsoleFeedItem = {
  time: string;
  type: string;
  message: string;
  route: string;
};

export type ConsoleRepo = {
  slug: string;
  name: string;
  kernel: string;
  drift: string;
  driftOk: boolean;
  sync: string;
  kpis: {
    cells: number;
    signals24h: number;
    gateRate: string;
    gateOk: boolean;
    awaiting: number;
    proposals: number;
  };
  gates: ConsoleGate[];
  gateSummary: string;
  territories: ConsoleTerritory[];
  queue: ConsoleQueueItem[];
  proposals: ConsoleProposal[];
  autonomyMix: ConsoleAutonomy[];
  feed: ConsoleFeedItem[];
};

export type ConsoleMock = {
  version: string;
  locale: string;
  filters: ConsoleFilter[];
  repos: ConsoleRepo[];
};

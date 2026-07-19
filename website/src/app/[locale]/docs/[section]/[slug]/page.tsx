import { Suspense } from "react";
import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { CopyButton } from "@/components/CopyButton";
import { CliExplorer } from "@/components/docs/CliExplorer";
import { DocsShell } from "@/components/docs/DocsShell";
import {
  findDoc,
  flattenDocs,
  getCli,
  getDocsIndex,
  type DocsBlock,
} from "@/lib/content";
import { isLocale, localePath, type Locale } from "@/lib/i18n";

const ui = {
  en: {
    search: "Search docs…",
    empty: "No pages match",
    version: "Version",
    menu: "Menu",
    close: "Close",
    editOnGithub: "Edit this page on GitHub",
    prev: "Previous",
    next: "Next",
    breadcrumbDocs: "Docs",
    copy: "Copy",
    copied: "Copied ✓",
    syntax: "Syntax",
    reads: "Reads",
    writes: "Writes",
    example: "Example",
    output: "Result",
  },
  pt: {
    search: "Buscar nas docs…",
    empty: "Nenhuma página corresponde",
    version: "Versão",
    menu: "Menu",
    close: "Fechar",
    editOnGithub: "Editar esta página no GitHub",
    prev: "Anterior",
    next: "Próxima",
    breadcrumbDocs: "Docs",
    copy: "Copiar",
    copied: "Copiado ✓",
    syntax: "Sintaxe",
    reads: "Lê",
    writes: "Escreve",
    example: "Exemplo",
    output: "Resultado",
  },
} as const;

export function generateStaticParams() {
  const params: Array<{ locale: string; section: string; slug: string }> = [];
  for (const locale of ["en", "pt"] as const) {
    const index = getDocsIndex(locale);
    for (const page of index.pages) {
      params.push({
        locale,
        section: page.sectionSlug,
        slug: page.slug,
      });
    }
  }
  return params;
}

export function generateMetadata({
  params,
}: {
  params: { locale: string; section: string; slug: string };
}): Metadata {
  if (!isLocale(params.locale)) return {};
  const index = getDocsIndex(params.locale);
  const doc = findDoc(index, params.section, params.slug);
  if (!doc) return {};
  return { title: `${doc.title} — ARAH Harness` };
}

function isCliBlock(blocks?: DocsBlock[]): boolean {
  return !!blocks?.some((b) => b.type === "cli-reference");
}

export default function DocPage({
  params,
}: {
  params: { locale: string; section: string; slug: string };
}) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const t = ui[locale];
  const index = getDocsIndex(locale);
  const doc = findDoc(index, params.section, params.slug);
  if (!doc) notFound();

  const pages = flattenDocs(index);
  const currentIdx = pages.findIndex(
    (p) => p.sectionSlug === doc.sectionSlug && p.slug === doc.slug,
  );
  const prev = currentIdx > 0 ? pages[currentIdx - 1] : undefined;
  const next =
    currentIdx >= 0 && currentIdx < pages.length - 1
      ? pages[currentIdx + 1]
      : undefined;

  const editHref = `https://github.com/sraphaz/arah-harness/edit/main/website/content/docs/${locale}/index.json`;
  const isCli = isCliBlock(doc.blocks);

  return (
    <DocsShell
      locale={locale}
      nav={index.nav}
      version={index.version}
      activeSection={doc.sectionSlug}
      activeSlug={doc.slug}
      labels={{
        search: t.search,
        empty: t.empty,
        version: t.version,
        menu: t.menu,
        close: t.close,
      }}
    >
      <article className="mx-auto max-w-[820px] px-8 pb-24 pt-12">
        <nav className="mb-6 flex flex-wrap items-center gap-2 font-mono text-[11.5px] text-arah-dim">
          <Link
            href={localePath(locale, "/docs")}
            className="hover:text-arah-text"
          >
            {t.breadcrumbDocs}
          </Link>
          <span className="text-arah-fade">/</span>
          <span>{doc.section}</span>
          <span className="text-arah-fade">/</span>
          <span className="text-arah-text">{doc.title}</span>
          <a
            href={editHref}
            target="_blank"
            rel="noreferrer"
            className="ml-auto text-arah-dim hover:text-arah-text"
          >
            {t.editOnGithub} →
          </a>
        </nav>

        <h1 className="mb-4 font-display text-[clamp(28px,4vw,38px)] font-bold leading-[1.1] tracking-[-0.02em] text-balance">
          {doc.title}
        </h1>
        {doc.intro ? (
          <p className="mb-10 text-[17px] leading-relaxed text-arah-dim text-pretty">
            {doc.intro}
          </p>
        ) : null}

        {isCli ? (
          <Suspense
            fallback={
              <div className="h-[420px] rounded-panel border border-arah-line bg-arah-surface" />
            }
          >
            <CliExplorer
              commands={getCli(locale).commands}
              labels={{
                syntax: t.syntax,
                reads: t.reads,
                writes: t.writes,
                example: t.example,
                output: t.output,
                copy: t.copy,
                copied: t.copied,
              }}
            />
          </Suspense>
        ) : (
          <div className="flex flex-col gap-5">
            {doc.blocks?.map((block, i) => (
              <Block
                key={i}
                block={block}
                copy={t.copy}
                copied={t.copied}
              />
            ))}
          </div>
        )}

        <div className="mt-16 grid gap-3 border-t border-arah-hair pt-8 sm:grid-cols-2">
          {prev ? (
            <Link
              href={localePath(
                locale,
                `/docs/${prev.sectionSlug}/${prev.slug}`,
              )}
              className="rounded-card border border-arah-line bg-arah-surface p-4 text-left hover:border-[#3A4553]"
            >
              <div className="mb-1 font-mono text-[10.5px] uppercase tracking-[0.12em] text-arah-fade">
                ← {t.prev}
              </div>
              <div className="text-[14.5px] font-medium">{prev.title}</div>
            </Link>
          ) : (
            <div />
          )}
          {next ? (
            <Link
              href={localePath(
                locale,
                `/docs/${next.sectionSlug}/${next.slug}`,
              )}
              className="rounded-card border border-arah-line bg-arah-surface p-4 text-right hover:border-[#3A4553]"
            >
              <div className="mb-1 font-mono text-[10.5px] uppercase tracking-[0.12em] text-arah-fade">
                {t.next} →
              </div>
              <div className="text-[14.5px] font-medium">{next.title}</div>
            </Link>
          ) : (
            <div />
          )}
        </div>
      </article>
    </DocsShell>
  );
}

function Block({
  block,
  copy,
  copied,
}: {
  block: DocsBlock;
  copy: string;
  copied: string;
}) {
  switch (block.type) {
    case "heading":
      return (
        <h2 className="mt-6 font-display text-[22px] font-semibold tracking-[-0.015em]">
          {block.text}
        </h2>
      );
    case "paragraph":
      return (
        <p className="text-[16px] leading-relaxed text-arah-muted text-pretty">
          {block.text}
        </p>
      );
    case "list":
      return (
        <ul className="flex flex-col gap-2">
          {block.items.map((item, i) => (
            <li
              key={i}
              className="flex gap-3 text-[15px] leading-relaxed text-arah-muted"
            >
              <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-[oklch(78%_0.09_200)]" />
              <span>{item}</span>
            </li>
          ))}
        </ul>
      );
    case "code":
      return (
        <div className="overflow-hidden rounded-panel border border-arah-hair bg-arah-code">
          {block.label ? (
            <div className="flex items-center justify-between border-b border-arah-hair px-4 py-2">
              <span className="font-mono text-[11px] uppercase tracking-[0.12em] text-arah-fade">
                {block.label}
              </span>
              <CopyButton
                text={block.code}
                label={copy}
                copiedLabel={copied}
              />
            </div>
          ) : null}
          <pre className="m-0 overflow-auto px-4 py-4 font-mono text-[13.5px] leading-[1.7] text-arah-muted whitespace-pre-wrap">
            {block.code}
          </pre>
        </div>
      );
    case "cli-reference":
      return null;
  }
}

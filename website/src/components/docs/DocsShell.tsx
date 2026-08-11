"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useMemo, useState } from "react";
import type { DocsNavSection } from "@/lib/content";
import { localePath, type Locale } from "@/lib/i18n";

type Labels = {
  search: string;
  empty: string;
  version: string;
  menu: string;
  close: string;
};

export function DocsShell({
  locale,
  nav,
  version,
  activeSection,
  activeSlug,
  labels,
  children,
}: {
  locale: Locale;
  nav: DocsNavSection[];
  version?: string;
  activeSection?: string;
  activeSlug?: string;
  labels: Labels;
  children: React.ReactNode;
}) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const pathname = usePathname();

  const q = query.trim().toLowerCase();
  const results = useMemo(() => {
    if (!q) return null;
    const rows: Array<{
      sectionName: string;
      sectionSlug: string;
      slug: string;
      title: string;
    }> = [];
    for (const section of nav) {
      for (const page of section.pages) {
        if (
          page.title.toLowerCase().includes(q) ||
          section.name.toLowerCase().includes(q)
        ) {
          rows.push({
            sectionName: section.name,
            sectionSlug: section.slug,
            slug: page.slug,
            title: page.title,
          });
        }
      }
    }
    return rows;
  }, [nav, q]);

  function docHref(sectionSlug: string, slug: string) {
    return localePath(locale, `/docs/${sectionSlug}/${slug}`);
  }

  const sidebar = (
    <aside className="flex h-full flex-col gap-6 overflow-y-auto border-r border-arah-hair bg-arah-bg px-6 py-8 md:sticky md:top-16 md:h-[calc(100vh-64px)]">
      <div className="flex items-center justify-between">
        <Link
          href={localePath(locale, "/docs")}
          className="font-display text-[15px] font-semibold"
        >
          Docs
          {version ? (
            <span className="ml-2 rounded-control border border-accent bg-accent-tint px-1.5 py-0.5 font-mono text-[10px] text-accent">
              v{version}
            </span>
          ) : null}
        </Link>
        <button
          type="button"
          onClick={() => setOpen(false)}
          className="rounded-control border border-arah-control px-2 py-1 font-mono text-[11px] text-arah-dim md:hidden"
        >
          {labels.close}
        </button>
      </div>
      <input
        type="search"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder={labels.search}
        className="w-full rounded-control border border-arah-control bg-arah-surface px-3 py-2 font-mono text-[12.5px] text-arah-text placeholder:text-arah-fade focus:border-accent focus:outline-none"
      />
      {results ? (
        results.length === 0 ? (
          <p className="font-mono text-[12px] text-arah-fade">{labels.empty}</p>
        ) : (
          <ul className="flex flex-col gap-1">
            {results.map((r) => (
              <li key={`${r.sectionSlug}/${r.slug}`}>
                <Link
                  href={docHref(r.sectionSlug, r.slug)}
                  onClick={() => {
                    setQuery("");
                    setOpen(false);
                  }}
                  className="block rounded-control px-3 py-2 text-[13.5px] text-arah-text hover:bg-arah-surface"
                >
                  <span className="mb-0.5 block font-mono text-[10.5px] text-arah-fade">
                    {r.sectionName}
                  </span>
                  {r.title}
                </Link>
              </li>
            ))}
          </ul>
        )
      ) : (
        <nav className="flex flex-col gap-6">
          {nav.map((section) => (
            <div key={section.slug}>
              <div className="mb-2 font-mono text-[10.5px] uppercase tracking-[0.12em] text-arah-fade">
                {section.name}
              </div>
              <ul className="flex flex-col">
                {section.pages.map((page) => {
                  const active =
                    activeSection === section.slug &&
                    activeSlug === page.slug;
                  const href = docHref(section.slug, page.slug);
                  return (
                    <li key={page.slug}>
                      <Link
                        href={href}
                        onClick={() => setOpen(false)}
                        aria-current={active ? "page" : undefined}
                        className={`block border-l-2 px-3 py-1.5 text-[13.5px] ${
                          active
                            ? "border-[oklch(78%_0.09_200)] bg-[#10151C] text-arah-text"
                            : "border-transparent text-arah-dim hover:border-arah-chip hover:text-arah-text"
                        }`}
                      >
                        {page.title}
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </nav>
      )}
    </aside>
  );

  return (
    <div className="grid md:grid-cols-[264px_minmax(0,1fr)]">
      {/* Mobile toggle */}
      <div className="flex items-center justify-between border-b border-arah-hair px-6 py-3 md:hidden">
        <Link
          href={localePath(locale, "/docs")}
          className="font-display text-[15px] font-semibold"
        >
          Docs {version ? <span className="text-arah-dim">v{version}</span> : null}
        </Link>
        <button
          type="button"
          onClick={() => setOpen(true)}
          className="rounded-control border border-arah-control px-3 py-1.5 font-mono text-[12px] text-arah-dim"
        >
          {labels.menu}
        </button>
      </div>

      {/* Drawer on mobile */}
      {open ? (
        <div
          className="fixed inset-0 z-50 md:hidden"
          onClick={() => setOpen(false)}
        >
          <div className="absolute inset-0 bg-black/70" />
          <div
            className="absolute inset-y-0 left-0 w-[300px] max-w-[85vw]"
            onClick={(e) => e.stopPropagation()}
          >
            {sidebar}
          </div>
        </div>
      ) : null}

      {/* Sidebar on desktop */}
      <div className="hidden md:block">{sidebar}</div>

      <div className="min-w-0" data-pathname={pathname}>
        {children}
      </div>
    </div>
  );
}

"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Logo } from "./Logo";
import {
  localePath,
  navCopy,
  otherLocale,
  swapLocalePath,
  type Locale,
} from "@/lib/i18n";

export function Nav({ locale }: { locale: Locale }) {
  const pathname = usePathname() || `/${locale}`;
  const t = navCopy[locale];
  const links = [
    { href: localePath(locale), label: t.home },
    { href: localePath(locale, "/architecture"), label: t.architecture },
    { href: localePath(locale, "/how-it-works"), label: t.how },
    { href: localePath(locale, "/techorganism"), label: t.organism },
    { href: localePath(locale, "/use-cases"), label: t.useCases },
    { href: localePath(locale, "/docs"), label: t.docs },
    { href: localePath(locale, "/console"), label: t.console },
  ];
  const next = otherLocale(locale);

  return (
    <header className="sticky top-0 z-50 border-b border-arah-hair2 bg-[rgba(10,12,15,0.82)] backdrop-blur-[14px]">
      <nav className="mx-auto flex min-h-16 max-w-wide flex-wrap items-center gap-x-5 gap-y-2 px-8 py-3">
        <Link href={localePath(locale)} className="flex items-center gap-2.5 font-display text-[15px] font-semibold text-arah-text">
          <Logo />
          <span>
            ARAH <span className="font-normal text-arah-dim">Harness</span>
          </span>
        </Link>
        <div className="flex flex-1 flex-wrap items-center gap-x-4 gap-y-1">
          {links.map((l) => {
            const active =
              l.href === localePath(locale)
                ? pathname === l.href
                : pathname.startsWith(l.href);
            return (
              <Link
                key={l.href}
                href={l.href}
                className={`font-mono text-[12.5px] tracking-wide ${
                  active ? "text-arah-text" : "text-arah-dim hover:text-arah-text"
                }`}
              >
                {l.label}
              </Link>
            );
          })}
        </div>
        <div className="flex items-center gap-2">
          <div className="flex overflow-hidden rounded-control border border-arah-control font-mono text-[11px]">
            <Link
              href={swapLocalePath(pathname, "en")}
              className={`px-2.5 py-1.5 ${locale === "en" ? "bg-accent-tint text-accent" : "text-arah-dim hover:text-arah-text"}`}
              hrefLang="en"
            >
              EN
            </Link>
            <Link
              href={swapLocalePath(pathname, "pt")}
              className={`px-2.5 py-1.5 ${locale === "pt" ? "bg-accent-tint text-accent" : "text-arah-dim hover:text-arah-text"}`}
              hrefLang="pt"
            >
              PT
            </Link>
          </div>
          <a
            href="https://github.com/sraphaz/arah-harness"
            className="rounded-control border border-arah-control px-3 py-1.5 font-mono text-[12px] text-arah-muted hover:border-[#3A4553] hover:text-arah-text"
            target="_blank"
            rel="noreferrer"
          >
            {t.github}
          </a>
          <Link
            href={localePath(locale, "/docs/getting-started/install")}
            className="rounded-control bg-accent px-3 py-1.5 font-mono text-[12px] font-semibold"
          >
            {t.start}
          </Link>
          <span className="sr-only">{next}</span>
        </div>
      </nav>
    </header>
  );
}

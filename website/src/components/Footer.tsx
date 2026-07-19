import Link from "next/link";
import { localePath, navCopy, type Locale } from "@/lib/i18n";

export function Footer({ locale }: { locale: Locale }) {
  const t = navCopy[locale];
  return (
    <footer className="border-t border-arah-hair bg-arah-bg">
      <div className="mx-auto flex max-w-site flex-col gap-6 px-8 py-12 md:flex-row md:items-end md:justify-between">
        <div>
          <p className="font-display text-lg text-arah-text">ARAH Harness</p>
          <p className="mt-2 max-w-md text-[14.5px] text-arah-dim">{t.tagline}</p>
        </div>
        <div className="flex flex-wrap gap-x-5 gap-y-2 font-mono text-[12px] text-arah-dim">
          <a href="https://github.com/sraphaz/arah-harness" className="hover:text-arah-text">
            GitHub
          </a>
          <a
            href="https://github.com/sraphaz/arah-harness/releases"
            className="hover:text-arah-text"
          >
            Releases
          </a>
          <a
            href="https://github.com/sraphaz/arah-harness/blob/main/CHANGELOG.md"
            className="hover:text-arah-text"
          >
            Changelog
          </a>
          <a
            href="https://github.com/sraphaz/arah-harness/blob/main/CONTRIBUTING.md"
            className="hover:text-arah-text"
          >
            Contributing
          </a>
          <a
            href="https://github.com/sraphaz/arah-harness/blob/main/LICENSE"
            className="hover:text-arah-text"
          >
            License
          </a>
          <Link href={localePath(locale, "/docs")} className="hover:text-arah-text">
            Docs
          </Link>
        </div>
      </div>
    </footer>
  );
}

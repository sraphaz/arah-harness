import { notFound } from "next/navigation";
import { ClientRedirect } from "@/components/ClientRedirect";
import { getDocsIndex } from "@/lib/content";
import { isLocale, localePath, type Locale } from "@/lib/i18n";

export function generateStaticParams() {
  return [{ locale: "en" }, { locale: "pt" }];
}

export default function DocsRootPage({
  params,
}: {
  params: { locale: string };
}) {
  if (!isLocale(params.locale)) notFound();
  const locale = params.locale as Locale;
  const index = getDocsIndex(locale);
  const first = index.nav[0]?.pages[0];
  const firstSection = index.nav[0]?.slug;
  if (first && firstSection) {
    return (
      <ClientRedirect
        href={localePath(locale, `/docs/${firstSection}/${first.slug}`)}
      />
    );
  }
  notFound();
}

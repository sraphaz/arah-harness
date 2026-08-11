import Link from "next/link";

const dict = {
  en: {
    code: "404",
    title: "This page is not part of the harness.",
    body: "The path you tried to visit does not exist. Head back to the home page, or jump straight into the docs.",
    home: "Home",
    docs: "Docs",
    architecture: "Architecture",
  },
  pt: {
    code: "404",
    title: "Esta página não faz parte do harness.",
    body: "O caminho que você tentou visitar não existe. Volte para a página inicial ou vá direto para as docs.",
    home: "Início",
    docs: "Docs",
    architecture: "Arquitetura",
  },
} as const;

export default function LocaleNotFound() {
  // NOTE: not-found renders outside of params context; default to EN.
  const t = dict.en;
  return (
    <div className="mx-auto flex min-h-[60vh] max-w-[720px] flex-col items-start justify-center gap-6 px-8 py-24">
      <p className="font-mono text-[13px] uppercase tracking-[0.18em] text-accent">
        {t.code}
      </p>
      <h1 className="font-display text-[clamp(28px,4vw,40px)] font-bold leading-[1.1] tracking-[-0.02em]">
        {t.title}
      </h1>
      <p className="text-[16.5px] leading-relaxed text-arah-dim text-pretty">
        {t.body}
      </p>
      <div className="flex flex-wrap gap-3">
        <Link
          href="/en"
          className="rounded-control bg-accent px-5 py-2.5 text-[14px] font-semibold"
        >
          {t.home} · EN
        </Link>
        <Link
          href="/pt"
          className="rounded-control border border-arah-chip px-5 py-2.5 text-[14px] font-medium hover:border-[#4A5563]"
        >
          {dict.pt.home} · PT
        </Link>
        <Link
          href="/en/docs"
          className="rounded-control border border-arah-chip px-5 py-2.5 text-[14px] font-medium hover:border-[#4A5563]"
        >
          {t.docs}
        </Link>
        <Link
          href="/en/architecture"
          className="rounded-control border border-arah-chip px-5 py-2.5 text-[14px] font-medium hover:border-[#4A5563]"
        >
          {t.architecture}
        </Link>
      </div>
    </div>
  );
}

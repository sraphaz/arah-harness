"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

/** Static-export-safe redirect (next/navigation redirect() is not supported with output: 'export'). */
export function ClientRedirect({ href }: { href: string }) {
  const router = useRouter();

  useEffect(() => {
    router.replace(href);
  }, [href, router]);

  return (
    <p className="mx-auto max-w-site px-8 py-16 font-mono text-sm text-arah-dim">
      Redirecting…
      <noscript>
        {" "}
        <a href={href} className="text-accent underline">
          Continue
        </a>
      </noscript>
    </p>
  );
}

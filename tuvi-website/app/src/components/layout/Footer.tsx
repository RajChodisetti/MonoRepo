import { siteContent } from "@/content/site";
import { getContactEmail, getLinkedInUrl } from "@/lib/env";

export default function Footer() {
  const year = new Date().getFullYear();

  return (
    <footer className="border-t border-border bg-bg-elevated px-5 py-12 md:px-8">
      <div className="mx-auto flex max-w-6xl flex-col gap-8 md:flex-row md:items-center md:justify-between">
        <div>
          <p className="font-display text-xl font-bold text-text">
            {siteContent.brand.name}
          </p>
          <p className="mt-1 text-sm text-muted">{siteContent.brand.tagline}</p>
          <a
            href={`mailto:${getContactEmail()}`}
            className="mt-3 inline-block text-sm text-cyan transition hover:text-gold"
          >
            {getContactEmail()}
          </a>
        </div>

        <div className="flex flex-wrap items-center gap-4 text-sm text-muted">
          <a
            href={getLinkedInUrl()}
            target="_blank"
            rel="noopener noreferrer"
            className="transition hover:text-cyan"
          >
            LinkedIn
          </a>
          {siteContent.footer.legal.map((link) => (
            <a key={link.label} href={link.href} className="transition hover:text-text">
              {link.label}
            </a>
          ))}
        </div>
      </div>

      <p className="mx-auto mt-10 max-w-6xl text-center text-xs text-muted/70 md:text-left">
        © {year} {siteContent.brand.name}. All rights reserved.
      </p>
    </footer>
  );
}

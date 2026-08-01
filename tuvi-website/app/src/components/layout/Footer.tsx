import Link from "next/link";
import { siteContent } from "@/content/site";
import BrandLogo from "@/components/layout/BrandLogo";
import { getCallInDisplay, getCallInTelHref, getContactEmail, getLinkedInUrl } from "@/lib/env";

export default function Footer() {
  const year = new Date().getFullYear();
  const callInDisplay = getCallInDisplay();
  const callInHref = getCallInTelHref();

  return (
    <footer className="border-t border-border bg-bg-elevated px-5 py-12 md:px-8 md:py-16">
      <div className="mx-auto grid max-w-6xl gap-10 md:grid-cols-[1.2fr_0.8fr] md:items-end">
        <div>
          <BrandLogo size="footer" />
          <p className="mt-5 max-w-md text-sm leading-6 text-muted">
            Thoughtful websites, applications, and AI systems for businesses ready to grow with
            stronger software.
          </p>
          <div className="mt-5 flex flex-wrap gap-x-5 gap-y-2 text-sm">
            <a href={`mailto:${getContactEmail()}`} className="font-semibold text-primary transition-colors hover:text-ink">
              {getContactEmail()}
            </a>
            {callInHref && callInDisplay ? (
              <a href={callInHref} className="text-muted transition-colors hover:text-ink">
                AI assistant · {callInDisplay}
              </a>
            ) : null}
          </div>
        </div>

        <nav aria-label="Footer navigation" className="md:text-right">
          <div className="flex flex-wrap gap-x-5 gap-y-3 text-sm text-muted md:justify-end">
            <a href={getLinkedInUrl()} target="_blank" rel="noopener noreferrer" className="transition-colors hover:text-ink">
              LinkedIn
            </a>
            {siteContent.footer.legal.map((link) => (
              <Link key={link.label} href={link.href} className="transition-colors hover:text-ink">
                {link.label}
              </Link>
            ))}
          </div>
          <p className="mt-6 text-xs leading-5 text-muted/80">
            © {year} {siteContent.brand.name}. All rights reserved.
          </p>
        </nav>
      </div>
    </footer>
  );
}

import type { ReactNode } from "react";
import SiteFooter from "@/components/layout/SiteFooter";
import { getContactEmail } from "@/lib/env";

type LegalPageProps = {
  eyebrow: string;
  title: string;
  summary: string;
  updated: string;
  children: ReactNode;
};

export function LegalPage({ eyebrow, title, summary, updated, children }: LegalPageProps) {
  const contactEmail = getContactEmail();

  return (
    <>
      <article className="hero-atmosphere relative overflow-hidden px-4 py-14 sm:px-8 sm:py-20 md:px-12 md:py-24">
        <div className="hero-grid pointer-events-none absolute inset-0 opacity-40" />
        <div className="relative mx-auto max-w-4xl">
          <header className="border-b border-border pb-10">
            <p className="text-xs font-semibold uppercase tracking-[0.16em] text-primary">{eyebrow}</p>
            <h1 className="mt-3 max-w-3xl font-display text-[clamp(2.6rem,6vw,4.8rem)] font-semibold leading-none tracking-[-0.04em] text-ink">
              {title}
            </h1>
            <p className="mt-5 max-w-3xl text-base leading-7 text-muted sm:text-lg">{summary}</p>
            <p className="mt-5 text-sm font-medium text-muted">Last updated: {updated}</p>
          </header>

          <div className="mt-10 space-y-10">{children}</div>

          <aside className="mt-12 rounded-3xl border border-border bg-bg p-6 shadow-[0_16px_50px_rgba(15,39,31,0.08)] sm:p-8">
            <h2 className="font-display text-2xl font-semibold tracking-[-0.02em] text-ink">Questions or requests</h2>
            <p className="mt-2 leading-7 text-muted">
              Contact us at{" "}
              <a href={`mailto:${contactEmail}`} className="font-semibold text-primary underline decoration-primary/25 underline-offset-4 transition hover:text-accent">
                {contactEmail}
              </a>
              .
            </p>
          </aside>
        </div>
      </article>
      <SiteFooter />
    </>
  );
}

export function LegalSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section>
      <h2 className="font-display text-2xl font-semibold tracking-[-0.025em] text-ink">{title}</h2>
      <div className="mt-3 space-y-4 text-[15px] leading-7 text-muted sm:text-base">{children}</div>
    </section>
  );
}

export function LegalList({ children }: { children: ReactNode }) {
  return <ul className="list-disc space-y-2 pl-6 marker:text-primary">{children}</ul>;
}

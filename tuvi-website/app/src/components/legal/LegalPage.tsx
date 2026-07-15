import type { ReactNode } from "react";
import Nav from "@/components/layout/Nav";
import Footer from "@/components/layout/Footer";
import { getContactEmail } from "@/lib/env";

type LegalPageProps = {
  eyebrow: string;
  title: string;
  summary: string;
  updated: string;
  children: ReactNode;
};

export function LegalPage({
  eyebrow,
  title,
  summary,
  updated,
  children,
}: LegalPageProps) {
  const contactEmail = getContactEmail();

  return (
    <>
      <Nav />
      <main id="main-content" tabIndex={-1} className="relative min-h-screen overflow-hidden pt-24">
        <div className="pointer-events-none absolute inset-0 grid-bg opacity-35" />
        <div className="pointer-events-none absolute left-0 top-28 h-72 w-72 -translate-x-1/3 rounded-full bg-sage/70 blur-[100px]" />
        <div className="pointer-events-none absolute right-0 top-[32rem] h-72 w-72 translate-x-1/3 rounded-full bg-parchment blur-[100px]" />

        <article className="relative mx-auto max-w-4xl px-5 py-12 md:px-8 md:py-20">
          <header className="border-b border-border pb-10">
            <p className="text-xs font-bold uppercase tracking-[0.16em] text-primary">
              {eyebrow}
            </p>
            <h1 className="mt-3 max-w-3xl font-display text-4xl font-bold tracking-tight text-ink md:text-5xl">
              {title}
            </h1>
            <p className="mt-5 max-w-3xl text-base leading-7 text-muted md:text-lg">
              {summary}
            </p>
            <p className="mt-5 text-sm font-medium text-muted">Last updated: {updated}</p>
          </header>

          <div className="mt-10 space-y-10">{children}</div>

          <aside className="mt-12 rounded-3xl border border-border bg-bg-elevated p-6 shadow-sm md:p-8">
            <h2 className="font-display text-xl font-bold text-ink">Questions or requests</h2>
            <p className="mt-2 leading-7 text-muted">
              Contact us at{" "}
              <a
                href={`mailto:${contactEmail}`}
                className="font-semibold text-primary transition hover:text-accent"
              >
                {contactEmail}
              </a>
              .
            </p>
          </aside>
        </article>
      </main>
      <Footer />
    </>
  );
}

export function LegalSection({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <section>
      <h2 className="font-display text-2xl font-bold tracking-tight text-ink">{title}</h2>
      <div className="mt-3 space-y-4 text-[15px] leading-7 text-muted md:text-base">
        {children}
      </div>
    </section>
  );
}

export function LegalList({ children }: { children: ReactNode }) {
  return <ul className="list-disc space-y-2 pl-6 marker:text-primary">{children}</ul>;
}

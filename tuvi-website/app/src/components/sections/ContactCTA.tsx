"use client";

import { siteContent } from "@/content/site";
import { getBookCallUrl, getContactEmail } from "@/lib/env";
import Button from "@/components/ui/Button";
import SectionShell from "@/components/ui/SectionShell";
import Reveal from "@/components/ui/Reveal";

export default function ContactCTA() {
  const { contact } = siteContent;

  return (
    <SectionShell id={contact.id}>
      <Reveal>
        <div className="relative overflow-hidden rounded-3xl border border-border bg-gradient-to-br from-surface via-bg-elevated to-surface p-10 text-center md:p-16">
          <div className="pointer-events-none absolute -left-20 top-0 h-64 w-64 rounded-full bg-gold/10 blur-[80px]" />
          <div className="pointer-events-none absolute -right-20 bottom-0 h-64 w-64 rounded-full bg-cyan/10 blur-[80px]" />

          <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-cyan">
            {contact.eyebrow}
          </p>
          <h2 className="relative mt-4 font-display text-3xl font-bold leading-tight text-text md:text-4xl lg:text-5xl">
            {contact.title}
          </h2>
          <p className="relative mx-auto mt-4 max-w-xl text-base text-muted md:text-lg">
            {contact.description}
          </p>

          <div className="relative mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button href={getBookCallUrl()}>
              {contact.primaryCta}
            </Button>
            <a
              href={`mailto:${getContactEmail()}`}
              className="text-sm text-muted transition hover:text-cyan"
            >
              {getContactEmail()}
            </a>
          </div>
        </div>
      </Reveal>
    </SectionShell>
  );
}

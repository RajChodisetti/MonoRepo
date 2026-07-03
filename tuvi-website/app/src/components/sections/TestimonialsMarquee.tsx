"use client";

import { siteContent } from "@/content/site";
import SectionHeading from "@/components/ui/SectionHeading";
import SectionShell from "@/components/ui/SectionShell";

function TestimonialCard({
  quote,
  author,
  company,
  initial,
}: {
  quote: string;
  author: string;
  company: string;
  initial: string;
}) {
  return (
    <div className="w-[min(380px,85vw)] shrink-0 rounded-2xl border border-border bg-surface/80 p-6 backdrop-blur transition hover:border-gold/30 hover:shadow-[0_0_40px_rgba(212,168,83,0.08)]">
      <p className="text-sm leading-relaxed text-muted md:text-base">&ldquo;{quote}&rdquo;</p>
      <div className="mt-6 flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-gold/30 to-cyan/20 font-display text-sm font-bold text-gold">
          {initial}
        </div>
        <div>
          <p className="text-sm font-semibold text-text">{author}</p>
          <p className="text-xs text-muted">{company}</p>
        </div>
      </div>
    </div>
  );
}

export default function TestimonialsMarquee() {
  const { testimonials } = siteContent;
  const items = [...testimonials.items, ...testimonials.items];

  return (
    <SectionShell id={testimonials.id} className="overflow-hidden">
      <SectionHeading
        eyebrow={testimonials.eyebrow}
        title={testimonials.title}
        align="center"
      />

      <div className="relative overflow-hidden">
        <div className="pointer-events-none absolute inset-y-0 left-0 z-10 w-12 bg-gradient-to-r from-bg to-transparent md:w-20" />
        <div className="pointer-events-none absolute inset-y-0 right-0 z-10 w-12 bg-gradient-to-l from-bg to-transparent md:w-20" />

        <div className="flex w-max animate-marquee gap-5 md:gap-6">
          {items.map((item, i) => (
            <TestimonialCard key={`${item.company}-${i}`} {...item} />
          ))}
        </div>
      </div>
    </SectionShell>
  );
}

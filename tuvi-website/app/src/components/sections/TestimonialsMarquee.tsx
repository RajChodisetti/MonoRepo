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
    <div className="card-soft w-[min(400px,85vw)] shrink-0 rounded-2xl p-6 transition hover:border-primary/30">
      <span className="font-display text-3xl font-bold leading-none text-primary" aria-hidden>
        &ldquo;
      </span>
      <p className="mt-1 text-sm leading-relaxed text-ink md:text-[15px]">{quote}</p>
      <div className="mt-5 flex items-center gap-3 border-t border-border pt-4">
        <div className="flex h-9 w-9 items-center justify-center rounded-full bg-ink font-display text-sm font-bold text-white">
          {initial}
        </div>
        <div>
          <p className="text-sm font-semibold text-ink">{author}</p>
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

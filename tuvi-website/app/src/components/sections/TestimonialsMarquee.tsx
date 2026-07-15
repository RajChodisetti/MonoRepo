"use client";

import { siteContent } from "@/content/site";
import SectionHeading from "@/components/ui/SectionHeading";
import SectionShell from "@/components/ui/SectionShell";
import Reveal from "@/components/ui/Reveal";

export default function TestimonialsMarquee() {
  const { testimonials } = siteContent;

  return (
    <SectionShell id={testimonials.id} className="bg-surface/70">
      <SectionHeading eyebrow={testimonials.eyebrow} title={testimonials.title} align="center" />
      <div className="grid gap-4 md:grid-cols-3">
        {testimonials.items.map((item, index) => (
          <Reveal key={item.number} delay={index * 0.06}>
            <article className="card-soft card-lift h-full rounded-3xl p-7">
              <span className="font-display text-4xl font-semibold text-primary/35">{item.number}</span>
              <h3 className="mt-8 font-display text-2xl font-semibold tracking-[-0.02em] text-ink">
                {item.title}
              </h3>
              <p className="mt-3 text-sm leading-6 text-muted md:text-[15px]">{item.description}</p>
            </article>
          </Reveal>
        ))}
      </div>
    </SectionShell>
  );
}

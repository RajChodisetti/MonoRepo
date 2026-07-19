"use client";

import type { ExperienceCard } from "@/data/types/restaurant";
import SourceAwareImage, { mediaForURL } from "@/components/SourceAwareImage";

export default function ExperienceCards({ cards }: { cards: ExperienceCard[] }) {
  return (
    <section className="bg-[#141210] py-24">
      <div className="mx-auto max-w-6xl px-6">
        <p className="text-xs uppercase tracking-[0.2em] text-brass">Experiences</p>
        <h2 className="font-display mt-3 text-4xl text-cream md:text-5xl">More ways to dine</h2>

        <div className="mt-12 grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {cards.map((card) => (
            <article
              key={card.id}
              className="group relative min-h-[320px] overflow-hidden rounded-xl"
            >
              <SourceAwareImage
                media={mediaForURL(card.image, card.title)}
                fill
                loading="lazy"
                className="object-cover transition duration-700 group-hover:scale-110"
                sizes="(max-width: 768px) 100vw, 25vw"
              />
              <div className="absolute inset-0 bg-gradient-to-t from-charcoal via-charcoal/40 to-transparent transition group-hover:from-charcoal/95" />
              <div className="absolute inset-x-0 bottom-0 p-6 transition group-hover:-translate-y-2">
                <h3 className="font-display text-2xl text-cream">{card.title}</h3>
                <p className="mt-2 text-sm text-cream/70">{card.description}</p>
                <a
                  href={card.cta.href}
                  className="mt-4 inline-flex text-xs font-semibold uppercase tracking-wider text-brass opacity-0 transition group-hover:opacity-100"
                >
                  {card.cta.label} →
                </a>
              </div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}

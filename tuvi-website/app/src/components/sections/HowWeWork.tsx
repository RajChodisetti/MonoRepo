"use client";

import SectionShell from "@/components/ui/SectionShell";
import SectionHeading from "@/components/ui/SectionHeading";
import Reveal from "@/components/ui/Reveal";

const steps = [
  {
    number: "01",
    title: "Tell us the goal",
    description:
      "A 30-minute call. You talk business, we listen — then we map the shortest path to a product that moves your numbers.",
  },
  {
    number: "02",
    title: "We design & build",
    description:
      "You see real, working software every week — not slide decks. Your first $1,000 of work is on us.",
  },
  {
    number: "03",
    title: "Launch & grow",
    description:
      "We ship it, wire up the analytics, and stay close — improving what works and cutting what doesn't.",
  },
];

export default function HowWeWork() {
  return (
    <SectionShell id="process">
      <SectionHeading
        eyebrow="How it works"
        title="From idea to launch, in three steps."
        align="center"
      />

      <div className="grid gap-4 md:grid-cols-3">
        {steps.map((step, i) => (
          <Reveal key={step.number} delay={i * 0.08}>
            <article className="card-soft card-lift relative h-full overflow-hidden rounded-2xl p-7">
              <span className="font-display text-5xl font-bold tracking-tight text-zinc-200 transition group-hover:text-violet-200 md:text-6xl">
                {step.number}
              </span>
              <h3 className="mt-4 font-display text-xl font-bold tracking-tight text-ink">
                {step.title}
              </h3>
              <p className="mt-2.5 text-sm leading-relaxed text-muted md:text-[15px]">
                {step.description}
              </p>
              {i < steps.length - 1 && (
                <span
                  className="absolute -right-2 top-8 hidden text-2xl text-zinc-300 md:block"
                  aria-hidden
                >
                  →
                </span>
              )}
            </article>
          </Reveal>
        ))}
      </div>
    </SectionShell>
  );
}

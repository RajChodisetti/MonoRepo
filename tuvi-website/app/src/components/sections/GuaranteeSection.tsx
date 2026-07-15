"use client";

import { motion } from "framer-motion";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";
import SectionHeading from "@/components/ui/SectionHeading";
import SectionShell from "@/components/ui/SectionShell";
import Reveal from "@/components/ui/Reveal";

export default function GuaranteeSection() {
  const { guarantee } = siteContent;

  return (
    <SectionShell id={guarantee.id} className="bg-bg-elevated">
      <SectionHeading
        eyebrow={guarantee.eyebrow}
        title={guarantee.title}
        description={guarantee.description}
        align="center"
      />

      <div className="grid gap-4 md:grid-cols-3">
        {guarantee.pillars.map((pillar, i) => (
          <Reveal key={pillar.title} delay={i * 0.08}>
            <motion.div className="card-soft card-lift h-full rounded-3xl p-7">
              <span className="inline-flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 font-display text-sm font-bold text-primary">
                {i + 1}
              </span>
              <h3 className="mt-4 font-display text-xl font-bold tracking-tight text-ink">
                {pillar.title}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted md:text-[15px]">
                {pillar.description}
              </p>
            </motion.div>
          </Reveal>
        ))}
      </div>

      <Reveal className="mt-12 text-center">
        <Button href={getBookCallUrl()}>
          {guarantee.cta} <span aria-hidden>→</span>
        </Button>
      </Reveal>
    </SectionShell>
  );
}

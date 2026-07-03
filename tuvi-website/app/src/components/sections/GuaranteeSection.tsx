"use client";

import { motion } from "framer-motion";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";
import SectionHeading from "@/components/ui/SectionHeading";
import SectionShell from "@/components/ui/SectionShell";
import Reveal from "@/components/ui/Reveal";

const icons = ["◆", "◇", "○"];

export default function GuaranteeSection() {
  const { guarantee } = siteContent;

  return (
    <SectionShell id={guarantee.id} className="bg-bg-elevated/40">
      <SectionHeading
        eyebrow={guarantee.eyebrow}
        title={guarantee.title}
        description={guarantee.description}
        align="center"
      />

      <div className="grid gap-6 md:grid-cols-3">
        {guarantee.pillars.map((pillar, i) => (
          <Reveal key={pillar.title} delay={i * 0.1}>
            <motion.div
              whileHover={{ y: -6 }}
              className="glass h-full rounded-2xl p-6 transition-shadow hover:shadow-[0_0_48px_rgba(56,189,248,0.08)]"
            >
              <span className="text-2xl text-gold">{icons[i]}</span>
              <h3 className="mt-4 font-display text-xl font-bold text-text">{pillar.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-muted">{pillar.description}</p>
            </motion.div>
          </Reveal>
        ))}
      </div>

      <Reveal className="mt-12 text-center">
        <Button href={getBookCallUrl()}>
          {guarantee.cta}
        </Button>
      </Reveal>
    </SectionShell>
  );
}

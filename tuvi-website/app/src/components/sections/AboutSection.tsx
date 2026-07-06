"use client";

import { motion } from "framer-motion";
import { siteContent } from "@/content/site";
import SectionHeading from "@/components/ui/SectionHeading";
import SectionShell from "@/components/ui/SectionShell";
import Reveal from "@/components/ui/Reveal";

export default function AboutSection() {
  const { about } = siteContent;

  return (
    <SectionShell id={about.id}>
      <div className="grid gap-12 lg:grid-cols-2 lg:items-center">
        <div>
          <SectionHeading eyebrow={about.eyebrow} title={about.title} />
          <div className="space-y-4 text-base leading-relaxed text-muted md:text-lg">
            {about.paragraphs.map((p) => (
              <Reveal key={p.slice(0, 24)}>
                <p>{p}</p>
              </Reveal>
            ))}
          </div>
        </div>

        <Reveal delay={0.15}>
          <div className="relative">
            <div className="absolute inset-0 rounded-3xl bg-gradient-to-br from-gold/10 to-cyan/10 blur-2xl" />
            <motion.div
              whileHover={{ scale: 1.02 }}
              transition={{ type: "spring", stiffness: 300 }}
              className="relative rounded-3xl border border-border bg-surface p-8"
            >
              <p className="text-xs font-bold uppercase tracking-widest text-gold">Core strengths</p>
              <ul className="mt-6 space-y-4">
                {about.highlights.map((item, i) => (
                  <motion.li
                    key={item}
                    initial={{ opacity: 0, x: 20 }}
                    whileInView={{ opacity: 1, x: 0 }}
                    viewport={{ once: true }}
                    transition={{ delay: i * 0.1 }}
                    className="flex items-start gap-3 border-b border-border pb-4 last:border-0 last:pb-0"
                  >
                    <span className="mt-1 flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-gold/15 text-xs font-bold text-gold">
                      {i + 1}
                    </span>
                    <span className="text-text">{item}</span>
                  </motion.li>
                ))}
              </ul>
            </motion.div>
          </div>
        </Reveal>
      </div>
    </SectionShell>
  );
}

"use client";

import { motion } from "framer-motion";
import { siteContent } from "@/content/site";
import SectionHeading from "@/components/ui/SectionHeading";
import SectionShell from "@/components/ui/SectionShell";
import Reveal from "@/components/ui/Reveal";

export default function TeamSection() {
  const { team } = siteContent;

  return (
    <SectionShell id={team.id}>
      <SectionHeading
        eyebrow={team.eyebrow}
        title={team.title}
        align="center"
      />

      <div className="mx-auto max-w-2xl">
        {team.members.map((member, i) => (
          <Reveal key={member.name} delay={i * 0.12}>
            <motion.article
              whileHover={{ scale: 1.02 }}
              className="glass group relative overflow-hidden rounded-3xl p-8"
            >
              <div className="pointer-events-none absolute -right-8 -top-8 h-32 w-32 rounded-full bg-gold/5 transition group-hover:bg-gold/10" />

              <div className="flex items-start gap-6">
                <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl border-2 border-gold/40 bg-gradient-to-br from-gold/20 to-cyan/10 font-display text-2xl font-bold text-gold shadow-[0_0_32px_rgba(212,168,83,0.15)]">
                  {member.initials}
                </div>

                <div className="min-w-0">
                  <h3 className="font-display text-2xl font-bold text-text">{member.name}</h3>
                  <p className="mt-1 text-sm font-semibold uppercase tracking-[0.14em] text-cyan">
                    {member.role}
                  </p>
                  <p className="mt-4 text-sm leading-relaxed text-muted md:text-base">
                    {member.bio}
                  </p>
                </div>
              </div>
            </motion.article>
          </Reveal>
        ))}
      </div>
    </SectionShell>
  );
}

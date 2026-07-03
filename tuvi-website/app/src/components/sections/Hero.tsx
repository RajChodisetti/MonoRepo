"use client";

import { motion } from "framer-motion";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";

const capabilities = ["System integrations", "AI infrastructure", "Mobile & web apps"];

function CapabilityCard() {
  return (
    <div className="relative w-full">
      <div className="relative overflow-hidden rounded-3xl border border-white/10 bg-gradient-to-br from-surface to-bg-elevated p-7 shadow-2xl md:p-8">
        <div className="pointer-events-none absolute -right-10 -top-10 h-40 w-40 rounded-full bg-cyan/10 blur-2xl" />
        <div className="pointer-events-none absolute -bottom-8 -left-8 h-32 w-32 rounded-full bg-gold/10 blur-2xl" />

        <div className="relative">
          <p className="text-xs font-bold uppercase tracking-widest text-cyan">Capability</p>
          <p className="mt-3 font-display text-xl font-bold leading-snug text-text md:text-2xl">
            Software that scales with your ambition
          </p>

          <ul className="mt-6 space-y-3">
            {capabilities.map((item, i) => (
              <motion.li
                key={item}
                initial={{ opacity: 0, x: 12 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.5 + i * 0.1, duration: 0.45 }}
                className="flex items-center gap-3 rounded-xl border border-white/8 bg-white/5 px-4 py-3.5 text-sm text-text"
              >
                <span className="h-2 w-2 shrink-0 rounded-full bg-gold" />
                {item}
              </motion.li>
            ))}
          </ul>
        </div>
      </div>

      <div
        className="pointer-events-none absolute -bottom-3 -left-3 hidden h-20 w-20 rounded-2xl border border-cyan/20 bg-cyan/5 backdrop-blur-sm lg:block"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute -right-2 -top-2 hidden h-14 w-14 rounded-full border border-gold/30 bg-gold/10 lg:block"
        aria-hidden
      />
    </div>
  );
}

export default function Hero() {
  const { hero } = siteContent;

  return (
    <section className="relative flex min-h-screen items-center pt-[72px]">
      <div className="pointer-events-none absolute inset-0 grid-bg opacity-50" />
      <div className="pointer-events-none absolute left-0 top-24 h-72 w-72 -translate-x-1/3 rounded-full bg-gold/10 blur-[100px]" />
      <div className="pointer-events-none absolute bottom-16 right-0 h-64 w-64 translate-x-1/4 rounded-full bg-cyan/10 blur-[90px]" />

      <div className="relative mx-auto w-full max-w-6xl px-5 py-12 md:px-8 md:py-20 lg:py-24">
        <div className="grid grid-cols-1 items-center gap-10 lg:grid-cols-2 lg:gap-14 xl:gap-20">
          <div className="min-w-0 max-w-xl lg:max-w-none">
            <motion.p
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
              className="text-[11px] font-bold uppercase tracking-[0.2em] text-cyan"
            >
              {hero.eyebrow}
            </motion.p>

            <h1 className="mt-4 font-display text-[2rem] font-bold leading-[1.1] sm:text-4xl md:text-5xl xl:text-[3.25rem]">
              {hero.headline.map((line, i) => (
                <motion.span
                  key={line}
                  className="block"
                  initial={{ opacity: 0, y: 28 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.6, delay: 0.15 + i * 0.12, ease: [0.22, 1, 0.36, 1] }}
                >
                  {i === 1 ? <span className="text-gradient">{line}</span> : line}
                </motion.span>
              ))}
            </h1>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.45 }}
              className="mt-5 text-base leading-relaxed text-muted md:mt-6 md:text-lg"
            >
              {hero.subcopy}
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.55 }}
              className="mt-7 flex flex-col gap-3 sm:mt-8 sm:flex-row sm:flex-wrap"
            >
              <Button href={getBookCallUrl()} className="w-full sm:w-auto">
                {hero.primaryCta}
              </Button>
              <Button href={hero.secondaryHref} variant="ghost" className="w-full sm:w-auto">
                {hero.secondaryCta}
              </Button>
            </motion.div>

            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.75 }}
              className="mt-7 flex flex-col gap-3 text-sm text-muted sm:mt-8 sm:flex-row sm:flex-wrap sm:items-center sm:gap-4"
            >
              <span className="w-fit rounded-full border border-gold/30 bg-gold/10 px-3 py-1 text-gold">
                {hero.trust}
              </span>
              <span>{hero.note}</span>
            </motion.div>
          </div>

          <motion.div
            initial={{ opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.65, delay: 0.35 }}
            className="min-w-0 w-full lg:max-w-[420px] lg:justify-self-end xl:max-w-[440px]"
          >
            <CapabilityCard />
          </motion.div>
        </div>
      </div>
    </section>
  );
}

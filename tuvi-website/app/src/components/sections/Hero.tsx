"use client";

import { motion, useReducedMotion } from "framer-motion";
import BrandLogo from "@/components/layout/BrandLogo";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";

const disciplines = ["Web", "Apps", "AI systems"] as const;

export default function Hero() {
  const { hero } = siteContent;
  const reduceMotion = useReducedMotion();
  const rise = reduceMotion ? { opacity: 1, y: 0 } : { opacity: 0, y: 22 };

  return (
    <section className="hero-blob relative flex min-h-[min(900px,100svh)] items-center overflow-hidden pt-24">
      <div className="pointer-events-none absolute inset-0 grid-bg opacity-35 [mask-image:radial-gradient(52rem_36rem_at_50%_30%,black,transparent)]" />

      <div className="relative mx-auto w-full max-w-6xl px-5 py-14 md:px-8 md:py-20 lg:py-24">
        <div className="grid items-center gap-12 lg:grid-cols-[1.05fr_0.95fr] lg:gap-16">
          <div className="max-w-2xl">
            <motion.p
              initial={rise}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.45 }}
              className="inline-flex items-center gap-2.5 rounded-full border border-border bg-bg-elevated/90 px-4 py-2 text-xs font-bold uppercase tracking-[0.14em] text-primary shadow-sm"
            >
              <span className="h-1.5 w-1.5 rounded-full bg-primary" aria-hidden="true" />
              {hero.eyebrow}
            </motion.p>

            <h1 className="mt-7 font-display text-[2.9rem] font-semibold leading-[0.98] tracking-[-0.04em] text-ink sm:text-6xl md:text-7xl lg:text-[5.25rem]">
              {hero.headline.map((line, index) => (
                <motion.span
                  key={line}
                  className={`block ${index === 1 ? "text-primary" : ""}`}
                  initial={rise}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.58, delay: reduceMotion ? 0 : 0.1 + index * 0.1 }}
                >
                  {line}
                </motion.span>
              ))}
            </h1>

            <motion.p
              initial={rise}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.55, delay: reduceMotion ? 0 : 0.32 }}
              className="mt-7 max-w-xl text-base leading-7 text-muted md:text-lg md:leading-8"
            >
              {hero.subcopy}
            </motion.p>

            <motion.div
              initial={rise}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.55, delay: reduceMotion ? 0 : 0.42 }}
              className="mt-8 flex flex-col gap-3 sm:flex-row sm:flex-wrap"
            >
              <Button href={getBookCallUrl()} className="w-full sm:w-auto">
                {hero.primaryCta} <span aria-hidden="true">→</span>
              </Button>
              <Button href={hero.secondaryHref} variant="ghost" className="w-full sm:w-auto">
                {hero.secondaryCta}
              </Button>
            </motion.div>

            <motion.div
              initial={{ opacity: reduceMotion ? 1 : 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: reduceMotion ? 0 : 0.58 }}
              className="mt-8 flex flex-col gap-2 border-l-2 border-primary/30 pl-4 text-sm text-muted sm:flex-row sm:items-center sm:gap-4"
            >
              <span className="font-semibold text-ink">{hero.trust}</span>
              <span className="hidden h-1 w-1 rounded-full bg-secondary sm:block" aria-hidden="true" />
              <span>{hero.note}</span>
            </motion.div>
          </div>

          <motion.div
            initial={reduceMotion ? { opacity: 1 } : { opacity: 0, y: 26, rotate: 1.5 }}
            animate={{ opacity: 1, y: 0, rotate: 0 }}
            transition={{ duration: 0.72, delay: reduceMotion ? 0 : 0.24, ease: [0.22, 1, 0.36, 1] }}
            className="mx-auto w-full max-w-[520px]"
          >
            <div className="logo-orbit relative aspect-square overflow-hidden rounded-[2.5rem] border border-border bg-bg-elevated shadow-[0_34px_90px_-40px_rgba(15,39,31,0.42)]">
              <div className="absolute inset-[8%]">
                <BrandLogo size="hero" showName={false} priority className="h-full w-full" />
              </div>
              <div className="absolute inset-x-5 bottom-5 grid grid-cols-3 gap-2 rounded-2xl border border-border bg-bg-elevated/90 p-2.5 shadow-lg backdrop-blur-md sm:inset-x-8 sm:bottom-8">
                {disciplines.map((item) => (
                  <span key={item} className="rounded-xl bg-surface px-2 py-2.5 text-center text-[0.68rem] font-bold uppercase tracking-[0.08em] text-ink sm:text-xs">
                    {item}
                  </span>
                ))}
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}

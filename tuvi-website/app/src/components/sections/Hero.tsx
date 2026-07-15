"use client";

import { motion, useReducedMotion } from "framer-motion";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";
import ServiceIcon, { type ServiceIconName } from "@/components/ui/ServiceIcon";

const solutionTracks: Array<{
  icon: ServiceIconName;
  title: string;
  description: string;
  number: string;
}> = [
  {
    icon: "ai",
    title: "AI systems",
    description: "Voice, chat and workflow automation",
    number: "01",
  },
  {
    icon: "website",
    title: "Web platforms",
    description: "Fast websites and dependable web products",
    number: "02",
  },
  {
    icon: "app",
    title: "Apps & integrations",
    description: "Connected tools for customers and teams",
    number: "03",
  },
];

export default function Hero() {
  const { hero } = siteContent;
  const reduceMotion = useReducedMotion();
  const rise = reduceMotion ? { opacity: 1, y: 0 } : { opacity: 0, y: 22 };

  return (
    <section className="hero-blob relative flex items-center overflow-hidden pt-24 xl:min-h-[760px]">
      <div className="pointer-events-none absolute inset-0 grid-bg opacity-35 [mask-image:radial-gradient(52rem_36rem_at_50%_30%,black,transparent)]" />

      <div className="relative mx-auto w-full max-w-6xl px-5 py-14 md:px-8 md:py-20 lg:py-24">
        <div className="grid items-center gap-12 xl:grid-cols-[1.18fr_0.82fr] xl:gap-14">
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

            <h1 className="mt-7 font-display text-[2.9rem] font-semibold leading-[0.98] tracking-[-0.04em] text-ink sm:text-6xl md:text-7xl lg:text-[4.5rem] xl:text-[4.65rem]">
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
            initial={reduceMotion ? { opacity: 1 } : { opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.72, delay: reduceMotion ? 0 : 0.24, ease: [0.22, 1, 0.36, 1] }}
            className="mx-auto hidden w-full max-w-[460px] xl:block"
          >
            <div className="ink-panel relative overflow-hidden rounded-[2rem] border border-white/10 p-5 text-[#fffef8] shadow-[0_34px_90px_-44px_rgba(15,39,31,0.72)] sm:p-7">
              <div className="pointer-events-none absolute -right-24 -top-28 h-72 w-72 rounded-full border border-white/10" />
              <div className="pointer-events-none absolute -bottom-32 -left-20 h-72 w-72 rounded-full border border-white/[0.07]" />

              <div className="relative flex items-center justify-between gap-4">
                <p className="text-xs font-semibold uppercase tracking-[0.16em] text-white/55">
                  What Tuvi builds
                </p>
                <span className="inline-flex items-center gap-2 rounded-full border border-sage/25 bg-sage/10 px-3 py-1.5 text-[11px] font-semibold uppercase tracking-[0.1em] text-sage">
                  <span className="h-1.5 w-1.5 rounded-full bg-sage" aria-hidden="true" />
                  AI + software
                </span>
              </div>

              <div className="relative mt-8 max-w-md">
                <p className="font-display text-3xl font-semibold leading-[1.08] tracking-[-0.03em] sm:text-4xl">
                  From a strong idea to software people can rely on.
                </p>
                <p className="mt-3 max-w-sm text-sm leading-6 text-white/60">
                  One focused team for strategy, design, engineering and the systems behind it.
                </p>
              </div>

              <div className="relative mt-8 divide-y divide-white/10 border-y border-white/10">
                {solutionTracks.map((item) => (
                  <div
                    key={item.title}
                    className="grid grid-cols-[auto_1fr_auto] items-center gap-4 py-4"
                  >
                    <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-white/[0.08] text-sage">
                      <ServiceIcon name={item.icon} className="h-5 w-5" />
                    </span>
                    <div className="min-w-0">
                      <p className="font-semibold text-[#fffef8]">{item.title}</p>
                      <p className="mt-0.5 text-xs leading-5 text-white/50 sm:text-sm">
                        {item.description}
                      </p>
                    </div>
                    <span className="font-mono text-xs text-white/30" aria-hidden="true">
                      {item.number}
                    </span>
                  </div>
                ))}
              </div>

              <div className="relative mt-6 grid grid-cols-4 gap-2 text-[11px] font-semibold uppercase tracking-[0.1em] text-white/45">
                {["Strategy", "Design", "Build", "Support"].map((stage) => (
                  <span key={stage} className="rounded-lg bg-white/[0.05] px-3 py-2 text-center">
                    {stage}
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

"use client";

import { motion } from "framer-motion";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";

/** Browser-window mockup — "we build websites" at a glance. */
function BrowserMockup() {
  return (
    <div className="relative w-full">
      <div className="card-soft relative overflow-hidden rounded-2xl shadow-[0_24px_64px_-24px_rgba(9,9,11,0.25)]">
        {/* Title bar */}
        <div className="flex items-center gap-2 border-b border-border bg-zinc-50 px-4 py-3">
          <span className="h-2.5 w-2.5 rounded-full bg-zinc-300" />
          <span className="h-2.5 w-2.5 rounded-full bg-zinc-300" />
          <span className="h-2.5 w-2.5 rounded-full bg-zinc-300" />
          <span className="ml-3 flex-1 truncate rounded-md bg-white px-3 py-1 text-[11px] font-medium text-muted ring-1 ring-border">
            yourbusiness.com
          </span>
        </div>

        {/* Mini site being "built" */}
        <div className="p-5">
          <div className="ink-panel rounded-xl p-5">
            <div className="h-2 w-16 rounded-full bg-white/40" />
            <div className="mt-3 h-3.5 w-44 rounded-full bg-white" />
            <div className="mt-2 h-3.5 w-32 rounded-full bg-white/70" />
            <div className="mt-4 flex gap-2">
              <div className="h-7 w-24 rounded-full bg-primary" />
              <div className="h-7 w-20 rounded-full bg-white/15 ring-1 ring-white/40" />
            </div>
          </div>

          <div className="mt-4 grid grid-cols-3 gap-3">
            {[0, 1, 2].map((i) => (
              <div key={i} className="rounded-lg border border-border bg-white p-3">
                <div
                  className={`h-6 w-6 rounded-md ${
                    ["bg-violet-100", "bg-blue-100", "bg-cyan-100"][i]
                  }`}
                />
                <div className="mt-2.5 h-2 w-full rounded-full bg-zinc-100" />
                <div className="mt-1.5 h-2 w-2/3 rounded-full bg-zinc-100" />
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Floating chips — the "other stuff" we build */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.7 }}
        className="animate-float absolute -left-5 top-16 hidden items-center gap-2 rounded-xl bg-white px-3.5 py-2.5 text-sm font-semibold text-ink shadow-lg ring-1 ring-border sm:flex lg:-left-10"
      >
        <span className="h-2 w-2 rounded-full bg-blue-500" /> Mobile & web apps
      </motion.div>
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.9 }}
        className="animate-float-delayed absolute -right-4 top-44 hidden items-center gap-2 rounded-xl bg-white px-3.5 py-2.5 text-sm font-semibold text-ink shadow-lg ring-1 ring-border sm:flex lg:-right-8"
      >
        <span className="h-2 w-2 rounded-full bg-primary" /> AI assistants
      </motion.div>
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 1.1 }}
        className="animate-float absolute -bottom-4 left-10 hidden items-center gap-2 rounded-xl bg-white px-3.5 py-2.5 text-sm font-semibold text-ink shadow-lg ring-1 ring-border sm:flex"
      >
        <span className="h-2 w-2 rounded-full bg-cyan-500" /> Systems & integrations
      </motion.div>
    </div>
  );
}

export default function Hero() {
  const { hero } = siteContent;

  return (
    <section className="hero-blob relative flex min-h-screen items-center pt-[72px]">
      <div className="pointer-events-none absolute inset-0 grid-bg opacity-60 [mask-image:radial-gradient(60rem_40rem_at_50%_0%,black,transparent)]" />

      <div className="relative mx-auto w-full max-w-6xl px-5 py-12 md:px-8 md:py-20 lg:py-24">
        <div className="grid grid-cols-1 items-center gap-12 lg:grid-cols-2 lg:gap-14 xl:gap-20">
          <div className="min-w-0 max-w-xl lg:max-w-none">
            <motion.span
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
              className="inline-flex items-center gap-2.5 rounded-full border border-border bg-white px-3.5 py-1.5 text-xs font-semibold text-ink shadow-sm"
            >
              <span className="h-2 w-2 rounded-full bg-primary" />
              {hero.eyebrow}
            </motion.span>

            <h1 className="mt-6 font-display text-[2.2rem] font-bold leading-[1.05] tracking-tight text-ink sm:text-4xl md:text-5xl xl:text-[3.5rem]">
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
              className="mt-6 text-base leading-relaxed text-muted md:text-lg"
            >
              {hero.subcopy}
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: 0.55 }}
              className="mt-8 flex flex-col gap-3 sm:flex-row sm:flex-wrap"
            >
              <Button href={getBookCallUrl()} className="w-full sm:w-auto">
                {hero.primaryCta} <span aria-hidden>→</span>
              </Button>
              <Button href={hero.secondaryHref} variant="ghost" className="w-full sm:w-auto">
                {hero.secondaryCta}
              </Button>
            </motion.div>

            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: 0.75 }}
              className="mt-8 flex flex-col gap-3 text-sm text-muted sm:flex-row sm:flex-wrap sm:items-center sm:gap-5"
            >
              <span className="flex items-center gap-1.5 font-medium text-ink">
                <span className="text-amber-500" aria-hidden>
                  ★★★★★
                </span>
                {hero.trust}
              </span>
              <span>{hero.note}</span>
            </motion.div>
          </div>

          <motion.div
            initial={{ opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.65, delay: 0.35 }}
            className="min-w-0 w-full lg:max-w-[460px] lg:justify-self-end"
          >
            <BrowserMockup />
          </motion.div>
        </div>
      </div>
    </section>
  );
}

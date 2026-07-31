"use client";

import { useEffect, useState } from "react";
import dynamic from "next/dynamic";
import { motion, useReducedMotion } from "framer-motion";
import { siteContent } from "@/content/site";
import { getBookCallUrl } from "@/lib/env";
import Button from "@/components/ui/Button";
import HeroServiceChip from "@/components/hero/HeroServiceChip";
import type { ServiceIconName } from "@/components/ui/ServiceIcon";

const HeroOrb3D = dynamic(() => import("@/components/hero/HeroOrb3D"), {
  ssr: false,
  loading: () => null,
});

const CHIP_CYCLE_MS = 3000;

const leftChips: Array<{
  icon: ServiceIconName;
  label: string;
  float: string;
  imageSrc: string;
}> = [
  {
    icon: "ai",
    label: "AI Systems",
    float: "animate-float",
    imageSrc: "/hero/chips/ai-robot.png",
  },
  {
    icon: "app",
    label: "App Dev",
    float: "animate-float-slow",
    imageSrc: "/hero/chips/app-dev.png",
  },
];

const rightChips: Array<{
  icon: ServiceIconName;
  label: string;
  float: string;
  imageSrc: string;
}> = [
  {
    icon: "website",
    label: "Web Platforms",
    float: "animate-float-delayed",
    imageSrc: "/hero/chips/web-platform.png",
  },
  {
    icon: "restaurant",
    label: "Restaurants",
    float: "animate-float",
    imageSrc: "/hero/chips/restaurants.png",
  },
];

const heroStats = [
  { value: "50+", label: "Projects complete" },
  { value: "7+", label: "Countries" },
  { value: "25+", label: "In progress" },
] as const;

const ease = [0.22, 1, 0.36, 1] as const;

export default function Hero() {
  const { hero } = siteContent;
  const reduceMotion = useReducedMotion();
  const [showChipImages, setShowChipImages] = useState(false);

  // One shared clock — all chips flip together
  useEffect(() => {
    if (reduceMotion) return;
    const id = window.setInterval(() => {
      setShowChipImages((v) => !v);
    }, CHIP_CYCLE_MS);
    return () => window.clearInterval(id);
  }, [reduceMotion]);

  return (
    <section className="hero-blob relative flex h-[100svh] max-h-[100svh] flex-col overflow-hidden pt-[5.5rem] md:pt-24">
      <div className="pointer-events-none absolute inset-0 grid-bg opacity-20 [mask-image:radial-gradient(50rem_34rem_at_50%_36%,black,transparent)]" />

      <div className="relative z-10 mx-auto flex h-full w-full max-w-5xl flex-col px-5 pb-3 md:px-8 md:pb-4">
        <div className="relative z-20 shrink-0 text-center">
          <motion.p
            initial={reduceMotion ? false : { opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease }}
            className="inline-flex items-center gap-2 text-[10px] font-semibold uppercase tracking-[0.24em] text-white/45 md:text-[11px]"
          >
            <span className="h-1.5 w-1.5 rounded-full bg-primary" aria-hidden="true" />
            {hero.eyebrow}
          </motion.p>

          <h1 className="mt-2 font-body text-[clamp(2.25rem,6.2vw,4.75rem)] font-bold leading-[0.92] tracking-[-0.055em] md:mt-3">
            <motion.span
              className="block text-white/28"
              initial={reduceMotion ? false : { opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.6, delay: reduceMotion ? 0 : 0.08, ease }}
            >
              {hero.headline[0]}
            </motion.span>
            <motion.span
              className="block text-white"
              initial={reduceMotion ? false : { opacity: 0, y: 24 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.65, delay: reduceMotion ? 0 : 0.16, ease }}
            >
              {hero.headline[1]}
            </motion.span>
          </h1>
        </div>

        <div className="relative mx-auto mt-3 flex min-h-0 w-full max-w-4xl flex-1 items-center overflow-visible md:mt-4">
          <div className="relative z-10 grid w-full grid-cols-1 items-center gap-3 overflow-visible md:grid-cols-[1fr_minmax(240px,380px)_1fr] md:gap-5">
            <div className="hidden h-[min(48vh,360px)] flex-col justify-between py-3 md:flex">
              {leftChips.map((chip, i) => (
                <HeroServiceChip
                  key={chip.label}
                  {...chip}
                  showImage={showChipImages}
                  align="left"
                  delay={reduceMotion ? 0 : 0.32 + i * 0.1}
                />
              ))}
            </div>

            <motion.div
              className="relative mx-auto aspect-square w-[min(58vw,300px)] overflow-visible bg-transparent sm:w-[min(48vw,340px)] md:w-full md:max-w-[380px]"
              initial={reduceMotion ? false : { opacity: 0, scale: 0.82 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.85, delay: reduceMotion ? 0 : 0.22, ease }}
              style={{ background: "transparent" }}
            >
              <div className="absolute inset-0 overflow-visible bg-transparent">
                <HeroOrb3D />
              </div>
            </motion.div>

            <div className="hidden h-[min(48vh,360px)] flex-col justify-between py-3 md:flex">
              {rightChips.map((chip, i) => (
                <HeroServiceChip
                  key={chip.label}
                  {...chip}
                  showImage={showChipImages}
                  align="right"
                  delay={reduceMotion ? 0 : 0.36 + i * 0.1}
                />
              ))}
            </div>
          </div>
        </div>

        <div className="mt-3 flex shrink-0 justify-center gap-2 overflow-x-auto pb-1 md:hidden">
          {[...leftChips, ...rightChips].map((chip, i) => (
            <HeroServiceChip
              key={chip.label}
              {...chip}
              showImage={showChipImages}
              compact
              delay={reduceMotion ? 0 : 0.35 + i * 0.06}
            />
          ))}
        </div>

        <div className="relative z-10 mx-auto mt-3 w-full max-w-lg shrink-0 text-center md:mt-4">
          <motion.p
            initial={reduceMotion ? false : { opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: reduceMotion ? 0 : 0.42, ease }}
            className="line-clamp-2 text-[13px] leading-6 text-white/55 md:text-sm md:leading-7"
          >
            {hero.subcopy}
          </motion.p>

          <motion.div
            initial={reduceMotion ? false : { opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: reduceMotion ? 0 : 0.5, ease }}
            className="mt-4 flex flex-wrap items-center justify-center gap-2.5 md:mt-5 md:gap-3"
          >
            <Button href={getBookCallUrl()} className="!bg-white !px-5 !text-black hover:!bg-white/90 md:!px-6">
              {hero.primaryCta}
            </Button>
            <Button
              href={hero.secondaryHref}
              variant="ghost"
              className="!border-white/20 !bg-transparent !text-white hover:!border-white/40 hover:!bg-white/[0.05]"
            >
              {hero.secondaryCta}
            </Button>
          </motion.div>
        </div>

        <motion.div
          initial={reduceMotion ? false : { opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.55, delay: reduceMotion ? 0 : 0.58, ease }}
          className="hero-stats-bar relative z-20 mt-4 grid shrink-0 grid-cols-3 gap-1 pt-4 md:mt-5 md:pt-5"
        >
          {heroStats.map((stat, index) => (
            <div
              key={stat.label}
              className={`flex flex-col items-center text-center ${
                index > 0 ? "border-l border-white/10" : ""
              }`}
            >
              <p className="font-body text-xl font-bold tracking-tight text-white sm:text-2xl md:text-3xl">
                {stat.value}
              </p>
              <p className="mt-1 text-[8px] font-semibold uppercase tracking-[0.18em] text-white/40 sm:text-[9px] md:text-[10px]">
                {stat.label}
              </p>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}

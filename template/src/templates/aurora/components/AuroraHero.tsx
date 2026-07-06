"use client";

import Image from "next/image";
import { motion } from "framer-motion";
import type { AuroraContent } from "../lib/mapContent";
import MagneticButton from "./ui/MagneticButton";
import BlurReveal from "./ui/BlurReveal";
import Marquee from "./ui/Marquee";

export default function AuroraHero({ content }: { content: AuroraContent["hero"] }) {
  return (
    <section id="hero" className="relative flex min-h-screen flex-col justify-center overflow-hidden pt-16">
      <div className="aurora-container relative z-10 grid items-center gap-12 py-20 lg:grid-cols-2">
        <div>
          <BlurReveal>
            <p className="mb-4 text-xs font-semibold uppercase tracking-[0.25em] text-cyan-400">
              {content.priceLevel || "Premium Dining"}
            </p>
          </BlurReveal>
          <BlurReveal delay={0.1}>
            <h1 className="aurora-heading text-5xl font-bold leading-[1.05] text-white md:text-7xl">
              {content.name}
            </h1>
          </BlurReveal>
          <BlurReveal delay={0.2}>
            <p className="aurora-gradient-text mt-4 text-xl font-medium md:text-2xl">
              {content.tagline}
            </p>
          </BlurReveal>
          <BlurReveal delay={0.3}>
            <p className="mt-4 max-w-lg text-white/60">{content.subheadline}</p>
          </BlurReveal>
          {content.rating && (
            <BlurReveal delay={0.35}>
              <p className="mt-4 text-sm text-purple-400">
                ★ {content.rating} · {content.reviewsCount || 0} reviews
              </p>
            </BlurReveal>
          )}
          <BlurReveal delay={0.4}>
            <div className="mt-8 flex flex-wrap gap-4">
              <MagneticButton href={content.primaryCTA.href}>
                {content.primaryCTA.label}
              </MagneticButton>
              <MagneticButton href={content.secondaryCTA.href} variant="secondary">
                {content.secondaryCTA.label}
              </MagneticButton>
            </div>
          </BlurReveal>
        </div>

        <BlurReveal delay={0.3} className="relative">
          <motion.div
            animate={{ y: [0, -12, 0] }}
            transition={{ duration: 6, repeat: Infinity, ease: "easeInOut" }}
            className="glass-card relative overflow-hidden p-2"
          >
            {content.poster && (
              <div className="relative aspect-[4/3] overflow-hidden rounded-xl">
                <Image
                  src={content.poster}
                  alt={`${content.name} preview`}
                  fill
                  priority
                  className="object-cover"
                  sizes="(max-width: 768px) 100vw, 50vw"
                />
                <div className="absolute inset-0 bg-gradient-to-t from-[#09090B] via-transparent to-transparent" />
              </div>
            )}
            <div className="absolute bottom-4 left-4 right-4 grid grid-cols-3 gap-2">
              {["Menu", "Reserve", "Gallery"].map((label) => (
                <div
                  key={label}
                  className="rounded-lg border border-white/10 bg-white/10 px-3 py-2 text-center text-[10px] uppercase tracking-wider text-white/80 backdrop-blur-md"
                >
                  {label}
                </div>
              ))}
            </div>
          </motion.div>

          <motion.div
            animate={{ y: [0, 8, 0] }}
            transition={{ duration: 5, repeat: Infinity, ease: "easeInOut", delay: 1 }}
            className="glass-card absolute -right-4 top-8 hidden w-40 p-4 md:block"
          >
            <p className="text-[10px] uppercase tracking-wider text-cyan-400">Rating</p>
            <p className="aurora-heading mt-1 text-2xl font-bold text-white">
              {content.rating || "4.9"}★
            </p>
          </motion.div>
        </BlurReveal>
      </div>

      <Marquee items={content.marqueeItems} />

      <a href="#menu" aria-label="Scroll" className="absolute bottom-24 left-1/2 -translate-x-1/2">
        <motion.span
          animate={{ opacity: [0.3, 1, 0.3] }}
          transition={{ duration: 2, repeat: Infinity }}
          className="block h-10 w-px bg-gradient-to-b from-cyan-400 to-transparent"
        />
      </a>
    </section>
  );
}

"use client";

import Link from "next/link";
import SectionShell from "@/components/ui/SectionShell";
import SectionHeading from "@/components/ui/SectionHeading";
import Reveal from "@/components/ui/Reveal";
import ServiceIcon, { type ServiceIconName } from "@/components/ui/ServiceIcon";

const offerings: Array<{
  icon: ServiceIconName;
  title: string;
  description: string;
  href?: string;
  linkLabel?: string;
  featured?: boolean;
}> = [
  {
    icon: "ai",
    title: "Useful AI systems",
    description: "Voice, chat, and workflow assistants with human review where decisions or communication matter.",
  },
  {
    icon: "website",
    title: "Websites that earn attention",
    description: "Fast, distinctive websites shaped around your brand and the action you want customers to take.",
  },
  {
    icon: "app",
    title: "Mobile & web applications",
    description: "Purpose-built products for customers and teams, designed to feel clear on every screen.",
  },
  {
    icon: "restaurant",
    title: "A connected digital guest journey for restaurants",
    description:
      "See how websites, QR ordering, rewards, reservation requests, voice AI, and guest campaigns can work as one considered restaurant experience.",
    href: "/services/restaurants",
    linkLabel: "Explore restaurant services",
    featured: true,
  },
  {
    icon: "growth",
    title: "Systems, data & growth",
    description: "Dashboards, integrations, and careful automation that reduce busywork and make the next decision easier.",
  },
];

export default function WhatWeBuild() {
  return (
    <SectionShell id="services" className="bg-surface/70">
      <SectionHeading
        eyebrow="What we build"
        title="Practical AI and software, built for real use."
        description="From an AI assistant or customer-facing website to a connected operations platform, we design every part around a clear outcome."
        align="center"
      />

      <div className="grid gap-4 lg:grid-cols-6">
        {offerings.map((item, index) => (
          <Reveal
            key={item.title}
            delay={index * 0.04}
            className={item.featured ? "lg:col-span-4" : "lg:col-span-2"}
          >
            <article
              className={`card-lift group flex h-full flex-col overflow-hidden rounded-3xl p-6 md:p-7 ${
                item.featured
                  ? "relative border border-white/15 bg-gradient-to-br from-white/10 to-white/[0.03] text-ink shadow-[0_24px_64px_-36px_rgba(0,0,0,0.7)]"
                  : "card-soft"
              }`}
            >
              {item.featured ? (
                <div className="pointer-events-none absolute -right-16 -top-20 h-64 w-64 rounded-full border border-white/15" />
              ) : null}
              <span
                className={`relative flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl ${
                  item.featured ? "bg-white/10 text-ink" : "bg-sage/70 text-primary"
                }`}
              >
                <ServiceIcon name={item.icon} />
              </span>
              <h3
                className={`relative mt-6 font-display text-2xl font-semibold leading-tight tracking-[-0.02em] ${
                  item.featured ? "text-ink md:max-w-xl md:text-3xl" : "text-ink"
                }`}
              >
                {item.title}
              </h3>
              <p
                className={`relative mt-3 flex-1 text-sm leading-6 md:text-[15px] ${
                  item.featured ? "max-w-2xl text-muted" : "text-muted"
                }`}
              >
                {item.description}
              </p>
              {item.href ? (
                <Link
                  href={item.href}
                  className="relative mt-7 inline-flex w-fit items-center gap-2 rounded-full bg-white px-5 py-2.5 text-sm font-semibold text-black transition-colors duration-200 hover:bg-white/90"
                >
                  {item.linkLabel} <span aria-hidden="true">→</span>
                </Link>
              ) : null}
            </article>
          </Reveal>
        ))}
      </div>
    </SectionShell>
  );
}

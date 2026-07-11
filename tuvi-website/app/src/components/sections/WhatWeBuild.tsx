"use client";

import Link from "next/link";
import SectionShell from "@/components/ui/SectionShell";
import SectionHeading from "@/components/ui/SectionHeading";
import Reveal from "@/components/ui/Reveal";

const offerings = [
  {
    icon: "🖥️",
    title: "Websites that win customers",
    description:
      "Fast, sharp websites built around your brand — from a landing page to a full online presence.",
    href: null,
  },
  {
    icon: "📱",
    title: "Mobile & web apps",
    description: "Custom apps your customers and team actually enjoy using — on any device.",
    href: null,
  },
  {
    icon: "🤖",
    title: "AI assistants & voice agents",
    description: "Assistants that answer calls, chat with customers, and book appointments 24/7.",
    href: null,
  },
  {
    icon: "🍽️",
    title: "Restaurant systems",
    description: "QR ordering, rewards, and reservations that turn first-time guests into regulars.",
    href: "/services/restaurants",
    linkLabel: "Watch the demos",
  },
  {
    icon: "📊",
    title: "Dashboards & integrations",
    description: "Your tools, connected. Your whole business, visible in one place.",
    href: null,
  },
  {
    icon: "🚀",
    title: "Growth & automation",
    description: "Email, SMS, and smart automations that bring customers back on autopilot.",
    href: null,
  },
];

export default function WhatWeBuild() {
  return (
    <SectionShell id="services" className="bg-white">
      <SectionHeading
        eyebrow="What we build"
        title="One studio. Everything digital."
        description="Tell us what your business needs — we design it, build it, and launch it."
        align="center"
      />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {offerings.map((item, i) => (
          <Reveal key={item.title} delay={i * 0.05}>
            <article className="card-soft card-lift group flex h-full flex-col rounded-2xl p-6">
              <div className="flex items-center gap-3.5">
                <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-border bg-zinc-50 text-xl">
                  {item.icon}
                </span>
                <h3 className="font-display text-lg font-bold tracking-tight text-ink">
                  {item.title}
                </h3>
              </div>
              <p className="mt-3.5 flex-1 text-sm leading-relaxed text-muted md:text-[15px]">
                {item.description}
              </p>
              {item.href && (
                <Link
                  href={item.href}
                  className="mt-4 inline-flex items-center gap-1.5 text-sm font-semibold text-primary transition group-hover:gap-2.5"
                >
                  {item.linkLabel} <span aria-hidden>→</span>
                </Link>
              )}
            </article>
          </Reveal>
        ))}
      </div>
    </SectionShell>
  );
}

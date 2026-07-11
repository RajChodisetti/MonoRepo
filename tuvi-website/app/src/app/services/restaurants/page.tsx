import type { Metadata } from "next";
import Link from "next/link";
import Nav from "@/components/layout/Nav";
import Footer from "@/components/layout/Footer";
import VoiceAssistantWidget from "@/components/VoiceAssistantWidget";
import Button from "@/components/ui/Button";

export const metadata: Metadata = {
  title: "Restaurant Systems | Tuvi Solutions",
  description:
    "QR ordering, rewards, reservations, voice AI, and guest growth systems — built for restaurants by Tuvi Solutions.",
};

const mediaBase = "/services/restaurants";

const featureVideos = [
  {
    id: "qr-ordering",
    badge: "Demo 1 — QR ordering",
    title: "Guests scan. Orders flow.",
    description:
      "Guests scan at the table, browse the menu, customize, and send clean orders straight to the team. No app to download, no waving for a waiter.",
    outcomes: ["Faster table turns", "Fewer order errors", "Works on any phone"],
    video: `${mediaBase}/qr-ordering-kitchen-v2.mp4`,
    poster: `${mediaBase}/qr-ordering-kitchen-v2-poster.png`,
  },
  {
    id: "rewards",
    badge: "Demo 2 — Rewards",
    title: "First visits become regulars.",
    description:
      "Guests check in, earn points, and see offers worth coming back for — without slowing down your counter for a second.",
    outcomes: ["Repeat visits tracked", "Points & perks built in", "Zero counter slowdown"],
    video: `${mediaBase}/rewards-reception-v3-pro.mp4`,
    poster: `${mediaBase}/rewards-reception-v3-pro-poster.png`,
  },
];

const serviceCards = [
  {
    icon: "🌐",
    title: "Restaurant websites",
    description: "A sharp digital front door built around your venue, menu, and brand.",
  },
  {
    icon: "📱",
    title: "QR ordering",
    description: "A table-ordering journey that gets guests to purchase faster.",
  },
  {
    icon: "🎁",
    title: "Rewards & membership",
    description: "Points, perks, and visit tracking that bring guests back.",
  },
  {
    icon: "🎙️",
    title: "AI voice receptionist",
    description: "Answers common questions and captures bookings — even at 2am.",
  },
  {
    icon: "📅",
    title: "Reservations",
    description: "Capture demand safely without overpromising tables.",
  },
  {
    icon: "💌",
    title: "Guest campaigns",
    description: "Email and SMS for specials, events, rewards, and win-backs.",
  },
];

function FeatureVideo({
  feature,
  index,
}: {
  feature: (typeof featureVideos)[number];
  index: number;
}) {
  const reversed = index % 2 === 1;

  return (
    <article id={feature.id} className="grid items-center gap-8 lg:grid-cols-[7fr_5fr] lg:gap-12">
      <div className={`relative ${reversed ? "lg:order-2" : ""}`}>
        <div className="relative overflow-hidden rounded-2xl shadow-[0_32px_64px_-24px_rgba(9,9,11,0.35)] ring-1 ring-border">
          <video
            className="aspect-video w-full bg-zinc-900 object-cover"
            poster={feature.poster}
            autoPlay
            loop
            muted
            playsInline
            preload="auto"
          >
            <source src={feature.video} type="video/mp4" />
          </video>
          <span className="absolute left-4 top-4 rounded-full bg-ink/90 px-3.5 py-1.5 text-[11px] font-semibold uppercase tracking-[0.08em] text-white backdrop-blur">
            {feature.badge}
          </span>
        </div>
      </div>

      <div className={reversed ? "lg:order-1" : ""}>
        <h3 className="font-display text-2xl font-bold tracking-tight text-ink md:text-3xl">
          {feature.title}
        </h3>
        <p className="mt-3 text-base leading-relaxed text-muted">{feature.description}</p>
        <ul className="mt-6 space-y-3">
          {feature.outcomes.map((outcome) => (
            <li
              key={outcome}
              className="flex items-center gap-3 text-sm font-medium text-ink md:text-base"
            >
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-[11px] font-bold text-primary">
                ✓
              </span>
              {outcome}
            </li>
          ))}
        </ul>
      </div>
    </article>
  );
}

export default function RestaurantServicesPage() {
  return (
    <>
      <Nav />
      <main className="min-h-screen bg-bg pb-24 text-ink md:pb-0">
        {/* Hero */}
        <section className="hero-blob relative overflow-hidden px-5 pb-14 pt-32 md:px-8 md:pb-24 md:pt-44">
          <div className="pointer-events-none absolute inset-0 grid-bg opacity-60 [mask-image:radial-gradient(60rem_36rem_at_50%_0%,black,transparent)]" />
          <div className="relative mx-auto max-w-6xl text-center">
            <span className="inline-flex items-center gap-2.5 rounded-full border border-border bg-white px-3.5 py-1.5 text-xs font-semibold text-ink shadow-sm">
              <span className="h-2 w-2 rounded-full bg-primary" aria-hidden />
              Tuvi for restaurants
            </span>
            <h1 className="mx-auto mt-6 max-w-3xl font-display text-3xl font-bold leading-[1.05] tracking-tight text-ink sm:text-4xl md:text-6xl">
              Restaurant systems <span className="text-gradient">guests love to use.</span>
            </h1>
            <p className="mx-auto mt-5 max-w-2xl text-base leading-relaxed text-muted md:text-lg">
              QR ordering, rewards, reservations, voice AI, and guest campaigns — one connected
              growth system for your restaurant.
            </p>
            <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row sm:flex-wrap">
              <Button href="#features" className="w-full sm:w-auto">
                Watch the demos <span aria-hidden>→</span>
              </Button>
              <Button href="#services" variant="ghost" className="w-full sm:w-auto">
                Explore services
              </Button>
            </div>
          </div>
        </section>

        {/* Feature demo videos */}
        <section id="features" className="border-t border-border bg-white px-5 py-16 md:px-8 md:py-24">
          <div className="mx-auto max-w-6xl">
            <div className="mx-auto max-w-3xl text-center">
              <span className="flex items-center justify-center gap-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-muted">
                <span className="h-2 w-2 rounded-[2px] bg-primary" aria-hidden />
                See it in action
              </span>
              <h2 className="mt-4 font-display text-3xl font-bold leading-[1.08] tracking-tight text-ink md:text-4xl lg:text-5xl">
                Watch what we&apos;d build for you.
              </h2>
              <p className="mt-4 text-base leading-relaxed text-muted md:text-lg">
                Two demos, the full guest experience: scan, order, check in, come back — with zero
                extra busywork for your staff.
              </p>
            </div>

            <div className="mt-14 space-y-16 md:mt-20 md:space-y-24">
              {featureVideos.map((feature, index) => (
                <FeatureVideo key={feature.id} feature={feature} index={index} />
              ))}
            </div>
          </div>
        </section>

        {/* Service grid */}
        <section id="services" className="border-t border-border px-5 py-16 md:px-8 md:py-24">
          <div className="mx-auto max-w-6xl">
            <div className="mx-auto max-w-3xl text-center">
              <span className="flex items-center justify-center gap-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-muted">
                <span className="h-2 w-2 rounded-[2px] bg-primary" aria-hidden />
                Restaurant services
              </span>
              <h2 className="mt-4 font-display text-3xl font-bold leading-[1.08] tracking-tight text-ink md:text-4xl lg:text-5xl">
                Everything around the guest journey.
              </h2>
            </div>

            <div className="mt-12 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {serviceCards.map((service) => (
                <article key={service.title} className="card-soft card-lift rounded-2xl p-6">
                  <div className="flex items-center gap-3.5">
                    <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-border bg-zinc-50 text-xl">
                      {service.icon}
                    </span>
                    <h3 className="font-display text-lg font-bold tracking-tight text-ink">
                      {service.title}
                    </h3>
                  </div>
                  <p className="mt-3.5 text-sm leading-relaxed text-muted md:text-[15px]">
                    {service.description}
                  </p>
                </article>
              ))}
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="px-5 py-16 md:px-8 md:py-20">
          <div className="ink-panel relative mx-auto max-w-6xl overflow-hidden rounded-[2rem] p-8 text-center md:p-14">
            <span className="relative inline-flex items-center gap-2.5 rounded-full border border-white/20 bg-white/10 px-3.5 py-1.5 text-xs font-semibold text-white">
              <span className="h-2 w-2 rounded-full bg-primary" aria-hidden />
              Next step
            </span>
            <h2 className="relative mx-auto mt-5 max-w-2xl font-display text-3xl font-bold leading-[1.08] tracking-tight text-white md:text-4xl lg:text-5xl">
              See what Tuvi can launch for your restaurant.
            </h2>
            <p className="relative mx-auto mt-4 max-w-2xl text-base leading-relaxed text-zinc-300 md:text-lg">
              Walk through websites, QR ordering, rewards, reservations, voice AI, and guest growth
              — on one free call.
            </p>
            <div className="relative mt-8">
              <Link
                href="/#contact"
                className="inline-flex w-full items-center justify-center gap-2 rounded-full bg-white px-6 py-3 text-sm font-semibold text-ink shadow-lg transition duration-300 hover:-translate-y-0.5 hover:shadow-xl sm:w-auto"
              >
                Schedule the walkthrough <span aria-hidden>→</span>
              </Link>
            </div>
            <p className="relative mt-4 text-sm text-zinc-400">
              Or ask the <span className="font-semibold text-white">Talk to Tuvi AI</span> assistant
              in the corner — it books the same calendar.
            </p>
          </div>
        </section>
      </main>
      <Footer />
      <VoiceAssistantWidget />
    </>
  );
}

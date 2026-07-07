import type { Metadata } from "next";
import Link from "next/link";
import VoiceAssistantWidget from "@/components/VoiceAssistantWidget";

export const metadata: Metadata = {
  title: "Restaurant Systems | Tuvi Solutions",
  description:
    "Restaurant QR ordering, rewards, reservations, voice AI, and customer growth systems from Tuvi Solutions.",
};

const mediaBase = "/services/restaurants";

const featureVideos = [
  {
    eyebrow: "Feature 01",
    title: "QR scanning to order",
    description:
      "Guests scan at the table, browse the menu, customize orders, and send clean requests to the team.",
    video: `${mediaBase}/qr-ordering-kitchen-v2.mp4`,
    poster: `${mediaBase}/qr-ordering-kitchen-v2-poster.png`,
  },
  {
    eyebrow: "Feature 02",
    title: "Rewards system",
    description:
      "Guests check in, earn points, and see redeemable offers without slowing down the counter.",
    video: `${mediaBase}/rewards-reception-v3-pro.mp4`,
    poster: `${mediaBase}/rewards-reception-v3-pro-poster.png`,
  },
];

const serviceCards = [
  {
    title: "Premium restaurant websites",
    description: "A polished digital front door built around the venue, menu, and brand.",
  },
  {
    title: "QR ordering",
    description: "A table-ordering journey that gives guests a faster path to purchase.",
  },
  {
    title: "Rewards and membership",
    description: "Points, member perks, visit tracking, and offers that bring guests back.",
  },
  {
    title: "AI voice receptionist",
    description: "A voice assistant that answers common questions and captures booking intent.",
  },
  {
    title: "Reservations",
    description: "A safer request workflow for demand capture without overpromising availability.",
  },
  {
    title: "Customer promotions",
    description: "Email and SMS-ready campaigns for specials, events, rewards, and reactivation.",
  },
];

export default function RestaurantServicesPage() {
  return (
    <>
      <main className="min-h-screen bg-[#070807] text-[#f7f2e8]">
        <header className="fixed inset-x-0 top-0 z-50 border-b border-white/10 bg-[#070807]/75 backdrop-blur-xl">
          <nav className="mx-auto flex h-16 max-w-6xl items-center justify-between px-5 md:h-[72px] md:px-8">
            <Link href="/" className="font-display text-lg font-bold tracking-tight md:text-xl">
              Tuvi Solutions<span className="text-gold">.</span>
            </Link>
            <div className="flex items-center gap-3">
              <Link
                href="/"
                className="rounded-full border border-white/15 px-4 py-2 text-xs font-semibold uppercase tracking-[0.14em] text-[#f7f2e8] transition hover:border-cyan/60 hover:text-cyan"
              >
                Home
              </Link>
              <Link
                href="/#contact"
                className="hidden rounded-full bg-gradient-to-r from-gold-dim to-gold px-4 py-2 text-xs font-bold uppercase tracking-[0.12em] text-bg transition hover:-translate-y-0.5 sm:inline-flex"
              >
                Book a Call
              </Link>
            </div>
          </nav>
        </header>

        <section className="relative overflow-hidden px-5 pb-14 pt-28 md:px-8 md:pb-20 md:pt-36">
          <div
            className="absolute inset-0 opacity-55"
            style={{
              backgroundImage: `linear-gradient(90deg, rgba(7,8,7,0.96), rgba(7,8,7,0.7) 45%, rgba(7,8,7,0.18)), url("${mediaBase}/qr-rewards-product-scene.jpg")`,
              backgroundPosition: "center",
              backgroundSize: "cover",
            }}
          />
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_20%,rgba(56,189,248,0.18),transparent_30rem),radial-gradient(circle_at_70%_70%,rgba(212,168,83,0.16),transparent_28rem)]" />
          <div className="relative mx-auto max-w-6xl">
            <p className="text-xs font-bold uppercase tracking-[0.18em] text-cyan">
              Tuvi Solutions for restaurants
            </p>
            <h1 className="mt-5 max-w-3xl font-display text-4xl font-bold leading-tight text-text md:text-6xl">
              Smart restaurant systems guests can use right away.
            </h1>
            <p className="mt-5 max-w-2xl text-base leading-8 text-white/70 md:text-lg">
              QR ordering, rewards, reservations, voice AI, and customer campaigns connected into one
              practical growth system for restaurants.
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Link
                href="#features"
                className="rounded-full bg-gradient-to-r from-gold-dim to-gold px-5 py-3 text-sm font-bold text-bg transition hover:-translate-y-0.5"
              >
                Watch main features
              </Link>
              <Link
                href="#services"
                className="rounded-full border border-white/15 px-5 py-3 text-sm font-semibold text-text transition hover:border-cyan/60 hover:text-cyan"
              >
                Explore services
              </Link>
            </div>
          </div>
        </section>

        <section id="features" className="border-t border-white/10 px-5 py-16 md:px-8 md:py-20">
          <div className="mx-auto max-w-6xl">
            <div className="max-w-3xl">
              <p className="text-xs font-bold uppercase tracking-[0.18em] text-cyan">Main features</p>
              <h2 className="mt-4 font-display text-3xl font-bold leading-tight text-text md:text-5xl">
                Ordering and rewards are the moments guests notice first.
              </h2>
              <p className="mt-4 text-base leading-8 text-white/65">
                These two videos show the front-of-house experience: guests scan, order, check in,
                and keep coming back without adding more busywork for staff.
              </p>
            </div>

            <div className="mt-10 grid gap-5 lg:grid-cols-2">
              {featureVideos.map((item) => (
                <article
                  key={item.title}
                  className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.04] shadow-2xl shadow-black/30"
                >
                  <div className="relative">
                    <video
                      className="aspect-video w-full bg-black object-cover"
                      poster={item.poster}
                      controls
                      muted
                      playsInline
                      preload="metadata"
                    >
                      <source src={item.video} type="video/mp4" />
                    </video>
                    <span className="absolute left-4 top-4 rounded-full border border-cyan/40 bg-bg/80 px-3 py-1 text-[10px] font-bold uppercase tracking-[0.14em] text-text backdrop-blur">
                      {item.eyebrow}
                    </span>
                  </div>
                  <div className="p-5">
                    <h3 className="font-display text-2xl font-bold text-text">{item.title}</h3>
                    <p className="mt-2 text-sm leading-6 text-white/65">{item.description}</p>
                  </div>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section id="services" className="border-t border-white/10 px-5 py-16 md:px-8 md:py-20">
          <div className="mx-auto max-w-6xl">
            <div className="max-w-3xl">
              <p className="text-xs font-bold uppercase tracking-[0.18em] text-cyan">Restaurant services</p>
              <h2 className="mt-4 font-display text-3xl font-bold leading-tight text-text md:text-5xl">
                More launch-ready systems around the guest journey.
              </h2>
            </div>

            <div className="mt-10 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {serviceCards.map((service, index) => (
                <article
                  key={service.title}
                  className="rounded-2xl border border-white/10 bg-white/[0.04] p-5 transition hover:-translate-y-1 hover:border-gold/50"
                >
                  <p className="text-xs font-bold uppercase tracking-[0.18em] text-gold">
                    {String(index + 1).padStart(2, "0")}
                  </p>
                  <h3 className="mt-4 font-display text-xl font-bold text-text">{service.title}</h3>
                  <p className="mt-3 text-sm leading-6 text-white/65">{service.description}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className="border-t border-white/10 px-5 py-16 md:px-8">
          <div className="mx-auto max-w-6xl rounded-3xl border border-white/10 bg-white/[0.04] p-6 text-center md:p-10">
            <p className="text-xs font-bold uppercase tracking-[0.18em] text-cyan">Next step</p>
            <h2 className="mx-auto mt-4 max-w-2xl font-display text-3xl font-bold leading-tight text-text md:text-5xl">
              See what Tuvi can launch for your restaurant.
            </h2>
            <p className="mx-auto mt-4 max-w-2xl text-base leading-8 text-white/65">
              Talk through websites, QR ordering, rewards, reservations, voice AI, and customer growth.
            </p>
            <Link
              href="/#contact"
              className="mt-7 inline-flex rounded-full bg-gradient-to-r from-gold-dim to-gold px-6 py-3 text-sm font-bold text-bg transition hover:-translate-y-0.5"
            >
              Schedule the walkthrough
            </Link>
          </div>
        </section>
      </main>
      <VoiceAssistantWidget />
    </>
  );
}

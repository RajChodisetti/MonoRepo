import type { Metadata } from "next";
import Link from "next/link";
import Nav from "@/components/layout/Nav";
import Footer from "@/components/layout/Footer";
import BrandLogo from "@/components/layout/BrandLogo";
import RestaurantFeatureVideo from "@/components/RestaurantFeatureVideo";
import VoiceAssistantWidget from "@/components/VoiceAssistantWidget";
import ServiceIcon, { type ServiceIconName } from "@/components/ui/ServiceIcon";
import Button from "@/components/ui/Button";

export const metadata: Metadata = {
  title: "Restaurant Services | Tuvi Solutions",
  description:
    "Websites, QR ordering, rewards, reservation requests, voice AI, and guest growth systems designed for restaurants by Tuvi Solutions.",
  alternates: { canonical: "/services/restaurants" },
  openGraph: {
    title: "Restaurant Services | Tuvi Solutions",
    description:
      "A considered digital guest journey for restaurants—from discovery and ordering to rewards and return visits.",
    type: "website",
    url: "/services/restaurants",
  },
};

const mediaBase =
  "https://sw-prod-files-syd1.syd1.cdn.digitaloceanspaces.com/tuvi/public/restaurant-services/v2";

const featureVideos = [
  {
    id: "qr-ordering",
    badge: "Guest flow 01 · Ordering",
    title: "From table scan to a clearer order request.",
    description:
      "Guests scan at the table, browse a mobile-first menu, customize their selection, and submit a structured request to the team—without downloading an app.",
    outcomes: ["Mobile-first menu journey", "Structured order requests", "No guest app download"],
    video: `${mediaBase}/qr-ordering-kitchen-v3-web.mp4`,
    poster: `${mediaBase}/qr-ordering-kitchen-v3-web-poster.jpg`,
  },
  {
    id: "rewards",
    badge: "Guest flow 02 · Rewards",
    title: "A return visit has a reason to happen.",
    description:
      "Guests can check in, see points, and understand available perks through a quick branded experience that complements the way your counter already works.",
    outcomes: ["Return visits can be tracked", "Flexible points and perks", "Designed for quick check-in"],
    video: `${mediaBase}/rewards-reception-v4-web.mp4`,
    poster: `${mediaBase}/rewards-reception-v4-web-poster.jpg`,
  },
] as const;

const serviceCards: Array<{
  icon: ServiceIconName;
  title: string;
  description: string;
}> = [
  {
    icon: "website",
    title: "Restaurant websites",
    description: "A distinctive digital front door shaped around the venue, menu, and next guest action.",
  },
  {
    icon: "qr",
    title: "QR ordering",
    description: "A table-ordering journey designed to make browsing and requesting items feel straightforward.",
  },
  {
    icon: "rewards",
    title: "Rewards & membership",
    description: "Points, perks, and visit tracking configured around the restaurant's actual loyalty model.",
  },
  {
    icon: "voice",
    title: "AI voice receptionist",
    description: "Answers approved common questions and captures booking requests when the team is occupied.",
  },
  {
    icon: "calendar",
    title: "Reservation requests",
    description: "Capture guest demand as pending requests without promising a table before confirmation.",
  },
  {
    icon: "campaign",
    title: "Reviewed guest campaigns",
    description: "Human-approved email drafts for specials, events, rewards, and return visits.",
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
    <article
      className={`grid items-center gap-8 lg:gap-16 ${
        reversed ? "lg:grid-cols-[0.78fr_1.22fr]" : "lg:grid-cols-[1.22fr_0.78fr]"
      }`}
    >
      <div className={`relative ${reversed ? "lg:order-2" : ""}`}>
        <div className="relative w-full overflow-hidden rounded-[2.25rem] border border-border bg-ink shadow-[0_34px_76px_-34px_rgba(15,39,31,0.45)]">
          <RestaurantFeatureVideo title={feature.title} poster={feature.poster} src={feature.video} />
          <span className="absolute left-4 top-4 rounded-full border border-white/20 bg-ink/90 px-3.5 py-1.5 text-[11px] font-semibold uppercase tracking-[0.1em] text-[#fffef8] backdrop-blur">
            {feature.badge}
          </span>
        </div>
      </div>

      <div className={reversed ? "lg:order-1" : ""}>
        <p className="text-xs font-bold uppercase tracking-[0.16em] text-primary">Designed around service</p>
        <h3 className="mt-3 font-display text-4xl font-semibold leading-[1.02] tracking-[-0.03em] text-ink md:text-5xl">
          {feature.title}
        </h3>
        <p className="mt-5 text-base leading-7 text-muted">{feature.description}</p>
        <ul className="mt-7 space-y-3">
          {feature.outcomes.map((outcome) => (
            <li key={outcome} className="flex items-center gap-3 text-sm font-semibold text-ink md:text-base">
              <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-sage text-xs font-bold text-primary" aria-hidden="true">
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
      <main id="main-content" tabIndex={-1} className="min-h-screen bg-bg text-ink">
        <section className="relative overflow-hidden bg-ink px-5 pb-16 pt-32 text-[#fffef8] md:px-8 md:pb-24 md:pt-40">
          <div className="pointer-events-none absolute -right-24 top-12 h-[34rem] w-[34rem] rounded-full border border-white/10" />
          <div className="pointer-events-none absolute -right-8 top-28 h-[26rem] w-[26rem] rounded-full border border-white/10" />
          <div className="relative mx-auto grid max-w-6xl items-center gap-12 lg:grid-cols-[1.08fr_0.92fr]">
            <div>
              <span className="inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/5 px-4 py-2 text-xs font-bold uppercase tracking-[0.14em] text-white/80">
                <span className="h-1.5 w-1.5 rounded-full bg-sage" aria-hidden="true" />
                Tuvi for restaurants
              </span>
              <h1 className="mt-7 max-w-3xl font-display text-5xl font-semibold leading-[0.98] tracking-[-0.04em] text-[#fffef8] sm:text-6xl md:text-7xl">
                Digital guest journeys, built around real service.
              </h1>
              <p className="mt-6 max-w-2xl text-base leading-7 text-white/70 md:text-lg md:leading-8">
                Bring discovery, ordering, rewards, reservation requests, voice AI, and guest communication into one considered restaurant experience.
              </p>
              <div className="mt-8 flex flex-col gap-3 sm:flex-row sm:flex-wrap">
                <Link href="#features" className="inline-flex items-center justify-center gap-2 rounded-full bg-[#fffef8] px-6 py-3 text-sm font-semibold text-ink transition-colors hover:bg-sage">
                  Watch the demos <span aria-hidden="true">↓</span>
                </Link>
                <Link href="/book" className="inline-flex items-center justify-center rounded-full border border-white/25 px-6 py-3 text-sm font-semibold text-[#fffef8] transition-colors hover:bg-white/10">
                  Plan a restaurant walkthrough
                </Link>
              </div>
            </div>

            <div className="mx-auto w-full max-w-[430px]">
              <div className="relative aspect-square overflow-hidden rounded-[2.5rem] bg-[#fffef8] shadow-[0_36px_100px_-38px_rgba(0,0,0,0.72)]">
                <div className="absolute inset-[7%]">
                  <BrandLogo size="hero" showName={false} priority className="h-full w-full" />
                </div>
                <span className="absolute bottom-6 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-full bg-ink px-4 py-2 text-[11px] font-bold uppercase tracking-[0.14em] text-[#fffef8]">
                  Restaurant growth systems
                </span>
              </div>
            </div>
          </div>
        </section>

        <section id="features" className="scroll-mt-28 border-b border-border bg-bg-elevated px-5 py-20 md:px-8 md:py-28">
          <div className="mx-auto max-w-7xl">
            <div className="max-w-3xl">
              <p className="text-xs font-bold uppercase tracking-[0.16em] text-primary">See the guest flows</p>
              <h2 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-[-0.03em] text-ink md:text-6xl">
                The experience is easier to judge when you can watch it.
              </h2>
              <p className="mt-5 max-w-2xl text-base leading-7 text-muted md:text-lg">
                These product demonstrations show the interaction model. A restaurant build is then shaped around its brand, menu, team, and operating rules.
              </p>
            </div>

            <div className="mt-16 space-y-20 md:mt-20 md:space-y-28">
              {featureVideos.map((feature, index) => (
                <FeatureVideo key={feature.id} feature={feature} index={index} />
              ))}
            </div>
          </div>
        </section>

        <section id="services" className="scroll-mt-28 bg-surface/65 px-5 py-20 md:px-8 md:py-28">
          <div className="mx-auto max-w-6xl">
            <div className="mx-auto max-w-3xl text-center">
              <p className="text-xs font-bold uppercase tracking-[0.16em] text-primary">Restaurant services</p>
              <h2 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-[-0.03em] text-ink md:text-6xl">
                Every module supports a moment in the guest journey.
              </h2>
            </div>

            <div className="mt-12 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {serviceCards.map((service, index) => (
                <article key={service.title} className="card-soft card-lift rounded-3xl p-7">
                  <div className="flex items-start justify-between gap-4">
                    <span className="flex h-12 w-12 items-center justify-center rounded-2xl bg-sage/75 text-primary">
                      <ServiceIcon name={service.icon} />
                    </span>
                    <span className="font-display text-2xl font-semibold text-primary/25">
                      {String(index + 1).padStart(2, "0")}
                    </span>
                  </div>
                  <h3 className="mt-6 font-display text-2xl font-semibold tracking-[-0.02em] text-ink">
                    {service.title}
                  </h3>
                  <p className="mt-3 text-sm leading-6 text-muted md:text-[15px]">{service.description}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className="bg-bg px-5 py-20 md:px-8 md:py-24">
          <div className="ink-panel relative mx-auto max-w-6xl overflow-hidden rounded-[2.5rem] p-8 md:p-14">
            <div className="pointer-events-none absolute -right-20 -top-20 h-72 w-72 rounded-full border border-white/10" />
            <div className="relative grid items-end gap-8 lg:grid-cols-[1fr_auto]">
              <div>
                <p className="text-xs font-bold uppercase tracking-[0.16em] text-white/60">Next step</p>
                <h2 className="mt-4 max-w-3xl font-display text-4xl font-semibold leading-[1.02] tracking-[-0.03em] text-[#fffef8] md:text-6xl">
                  Map the guest journey before choosing the software.
                </h2>
                <p className="mt-5 max-w-2xl text-base leading-7 text-white/70">
                  Bring your current menu, booking flow, and biggest service bottleneck. We&apos;ll use the call to shape a practical first milestone.
                </p>
              </div>
              <Button href="/book" variant="ghost" className="!border-white/20 !bg-[#fffef8] !text-ink hover:!bg-sage">
                Book the walkthrough <span aria-hidden="true">→</span>
              </Button>
            </div>
          </div>
        </section>
      </main>
      <Footer />
      <VoiceAssistantWidget />
    </>
  );
}

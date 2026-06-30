"use client";

import MagneticButton from "./ui/MagneticButton";
import BlurReveal from "./ui/BlurReveal";
import type { RestaurantContent } from "@/data/types/restaurant";
import { gmailComposeUrl, telHref } from "@/lib/reservation";

export default function CTABanner({ restaurant }: { restaurant: RestaurantContent }) {
  const email = restaurant.email
    ? gmailComposeUrl(restaurant.email, `Reservation — ${restaurant.name}`)
    : "";

  return (
    <section className="aurora-section">
      <div className="aurora-container">
        <BlurReveal>
          <div className="relative overflow-hidden rounded-3xl border border-purple-500/20 bg-gradient-to-br from-purple-900/30 via-[#0B1220] to-blue-900/30 p-12 text-center md:p-16">
            <div className="aurora-blob aurora-blob-1 !opacity-20" />
            <h2 className="aurora-heading relative text-4xl font-bold text-white md:text-5xl">
              Ready for an unforgettable meal?
            </h2>
            <p className="relative mx-auto mt-4 max-w-xl text-white/60">
              Join us at {restaurant.name} — reserve your table and experience hospitality
              reimagined.
            </p>
            <div className="relative mt-8 flex flex-wrap justify-center gap-4">
              <MagneticButton href={restaurant.primaryCTA.href}>
                {restaurant.primaryCTA.label}
              </MagneticButton>
              {restaurant.phone && (
                <MagneticButton href={telHref(restaurant.phone)} variant="secondary">
                  Call Now
                </MagneticButton>
              )}
              {email && (
                <MagneticButton href={email} variant="secondary">
                  Email Us
                </MagneticButton>
              )}
            </div>
          </div>
        </BlurReveal>
      </div>
    </section>
  );
}

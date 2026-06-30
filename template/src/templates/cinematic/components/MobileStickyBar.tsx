"use client";

import type { RestaurantContent } from "@/data/types/restaurant";
import { gmailComposeUrl, mapsHref, telHref } from "@/lib/reservation";

export default function MobileStickyBar({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const phone = telHref(restaurant.phone);
  const directions = mapsHref(restaurant.address, restaurant.coordinates);
  const email = restaurant.email
    ? gmailComposeUrl(restaurant.email, `Table reservation — ${restaurant.name}`)
    : "";

  return (
    <div className="fixed inset-x-0 bottom-0 z-50 grid grid-cols-4 border-t border-cream/10 bg-charcoal/95 backdrop-blur-md md:hidden">
      {phone && (
        <a href={phone} className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-cream/70">
          Call
        </a>
      )}
      <a href={directions} target="_blank" rel="noopener noreferrer" className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-cream/70">
        Directions
      </a>
      <a href={restaurant.primaryCTA.href} className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-brass">
        Reserve
      </a>
      <a href="#menu" className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-cream/70">
        Menu
      </a>
      {email && (
        <span className="sr-only">
          <a href={email}>Email</a>
        </span>
      )}
    </div>
  );
}

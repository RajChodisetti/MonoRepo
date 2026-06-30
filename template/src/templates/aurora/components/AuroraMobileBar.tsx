"use client";

import type { RestaurantContent } from "@/data/types/restaurant";
import { mapsHref, telHref } from "@/lib/reservation";

export default function AuroraMobileBar({ restaurant }: { restaurant: RestaurantContent }) {
  const phone = telHref(restaurant.phone);
  const directions = mapsHref(restaurant.address, restaurant.coordinates);

  return (
    <div className="fixed inset-x-0 bottom-0 z-50 grid grid-cols-4 border-t border-white/10 bg-[#09090B]/95 backdrop-blur-xl md:hidden">
      {phone && (
        <a href={phone} className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-white/60">
          Call
        </a>
      )}
      <a href={directions} target="_blank" rel="noopener noreferrer" className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-white/60">
        Map
      </a>
      <a href={restaurant.primaryCTA.href} className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-purple-400">
        Reserve
      </a>
      <a href="#menu" className="flex flex-col items-center justify-center py-3 text-[10px] uppercase tracking-wider text-white/60">
        Menu
      </a>
    </div>
  );
}

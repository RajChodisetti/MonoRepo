import type { RestaurantContent } from "@/data/types/restaurant";
import { mapsHref, telHref } from "@/lib/reservation";

export default function AuroraFooter({ restaurant }: { restaurant: RestaurantContent }) {
  return (
    <footer id="contact" className="border-t border-white/10 py-12 pb-24 md:pb-12">
      <div className="aurora-container grid gap-8 md:grid-cols-3">
        <div>
          <p className="aurora-heading text-xl font-bold text-white">{restaurant.name}</p>
          <p className="mt-2 text-sm text-white/50">{restaurant.address}</p>
        </div>
        <div className="text-sm text-white/50">
          <p>{restaurant.cuisine}</p>
          <p className="mt-1">{restaurant.locationLabel}</p>
          {restaurant.phone && (
            <a href={telHref(restaurant.phone)} className="mt-2 block text-cyan-400 hover:underline">
              {restaurant.phone}
            </a>
          )}
        </div>
        <div className="flex flex-wrap gap-3 text-sm">
          <a href={mapsHref(restaurant.address, restaurant.coordinates)} className="text-purple-400 hover:underline">
            Directions
          </a>
          <a href="#menu" className="text-purple-400 hover:underline">Menu</a>
          <a href={restaurant.primaryCTA.href} className="text-purple-400 hover:underline">Reserve</a>
        </div>
      </div>
      <p className="aurora-container mt-8 text-center text-xs text-white/30">
        Demo site · Powered by Tuvi Restaurant Platform · Aurora Template
      </p>
    </footer>
  );
}

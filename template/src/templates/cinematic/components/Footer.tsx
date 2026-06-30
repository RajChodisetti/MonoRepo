import type { RestaurantContent } from "@/data/types/restaurant";

export default function Footer({ restaurant }: { restaurant: RestaurantContent }) {
  return (
    <footer className="border-t border-cream/10 bg-charcoal py-8 pb-24 md:pb-8">
      <div className="mx-auto max-w-6xl px-6 text-center">
        <p className="font-display text-xl text-cream">{restaurant.name}</p>
        <p className="mt-2 text-xs text-cream/40">
          Demo site · Powered by Tuvi Restaurant Platform
        </p>
      </div>
    </footer>
  );
}

import type { RestaurantContent } from "@/data/types/restaurant";
import { mapsEmbedSrc } from "@/lib/reservation";
import BlurReveal from "./ui/BlurReveal";

const DAY_ORDER = [
  "monday", "tuesday", "wednesday", "thursday",
  "friday", "saturday", "sunday",
];

export default function AuroraLocation({ restaurant }: { restaurant: RestaurantContent }) {
  return (
    <section className="aurora-section">
      <div className="aurora-container grid gap-12 lg:grid-cols-2">
        <BlurReveal>
          <h2 className="aurora-heading text-3xl font-bold text-white">Location & Hours</h2>
          <p className="mt-4 text-white/60">{restaurant.address}</p>
          <table className="mt-8 w-full text-sm">
            <tbody>
              {DAY_ORDER.map((day) => (
                <tr key={day} className="border-b border-white/10">
                  <td className="py-2 capitalize text-white">{day}</td>
                  <td className="py-2 text-white/50">{restaurant.hours[day] || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </BlurReveal>
        {restaurant.coordinates && (
          <BlurReveal delay={0.1}>
            <div className="overflow-hidden rounded-2xl border border-white/10 aspect-[4/3]">
              <iframe
                title="Map"
                src={mapsEmbedSrc(restaurant.coordinates)}
                className="h-full w-full border-0"
                loading="lazy"
              />
            </div>
          </BlurReveal>
        )}
      </div>
    </section>
  );
}

import type { RestaurantContent } from "@/data/types/restaurant";
import { gmailComposeUrl, mapsEmbedSrc, mapsHref, telHref } from "@/lib/reservation";

const DAY_ORDER = [
  "monday", "tuesday", "wednesday", "thursday",
  "friday", "saturday", "sunday",
];

export default function LocationHours({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const directions = mapsHref(restaurant.address, restaurant.coordinates);
  const phone = telHref(restaurant.phone);
  const email = restaurant.email
    ? gmailComposeUrl(restaurant.email, `Inquiry — ${restaurant.name}`)
    : "";

  return (
    <section id="contact" className="bg-charcoal py-24 pb-32 md:pb-24">
      <div className="mx-auto grid max-w-6xl grid-cols-1 gap-12 px-6 md:grid-cols-2">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-brass">Find Us</p>
          <h2 className="font-display mt-3 text-4xl text-cream">Visit Us</h2>
          <address className="mt-6 not-italic text-cream/70 whitespace-pre-line">
            {restaurant.address}
          </address>

          <div className="mt-6 space-y-2 text-cream/80">
            {phone && (
              <a href={phone} className="block text-brass hover:underline">
                {restaurant.phone}
              </a>
            )}
            {email && (
              <a href={email} target="_blank" rel="noopener noreferrer" className="block text-brass hover:underline">
                {restaurant.email}
              </a>
            )}
            {restaurant.website && (
              <a href={restaurant.website} target="_blank" rel="noopener noreferrer" className="block text-brass hover:underline">
                Website
              </a>
            )}
          </div>

          <table className="mt-8 w-full text-sm">
            <caption className="mb-3 text-left text-xs font-semibold uppercase tracking-wider text-brass">
              Opening Hours
            </caption>
            <tbody>
              {DAY_ORDER.map((day) => (
                <tr key={day} className="border-b border-cream/10">
                  <td className="py-2 capitalize text-cream">{day}</td>
                  <td className="py-2 text-cream/60">{restaurant.hours[day] || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>

          <div className="mt-8 flex flex-wrap gap-3">
            <a href={directions} target="_blank" rel="noopener noreferrer" className="btn-primary">
              Get Directions
            </a>
            {phone && <a href={phone} className="btn-ghost">Call Now</a>}
            <a href={restaurant.primaryCTA.href} className="btn-ghost">Reserve</a>
          </div>
        </div>

        {restaurant.coordinates && (
          <div className="overflow-hidden rounded-xl border border-cream/10 aspect-[4/3]">
            <iframe
              title="Map"
              src={mapsEmbedSrc(restaurant.coordinates)}
              className="h-full w-full border-0"
              loading="lazy"
              referrerPolicy="no-referrer-when-downgrade"
            />
          </div>
        )}
      </div>
    </section>
  );
}

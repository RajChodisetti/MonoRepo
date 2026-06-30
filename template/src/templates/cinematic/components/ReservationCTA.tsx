import type { RestaurantContent } from "@/data/types/restaurant";
import { gmailComposeUrl } from "@/lib/reservation";

export default function ReservationCTA({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const emailUrl = restaurant.email
    ? gmailComposeUrl(restaurant.email, `Table reservation — ${restaurant.name}`)
    : "";

  const hoursPreview = Object.entries(restaurant.hours).slice(0, 3);

  return (
    <section id="reserve" className="bg-charcoal py-24">
      <div className="mx-auto max-w-3xl px-6 text-center">
        <h2 className="font-display text-4xl text-cream md:text-5xl">Your table is waiting.</h2>
        <p className="mt-4 text-cream/70">
          Join us at {restaurant.name} in {restaurant.city}. Call or email to reserve your table for
          dinner, celebrations, cocktails, or a quiet night built around unforgettable food.
        </p>

        {hoursPreview.length > 0 && (
          <div className="mt-6 text-sm text-cream/50">
            {hoursPreview.map(([day, hours]) => (
              <p key={day}>
                <span className="capitalize text-cream/70">{day}</span>: {hours}
              </p>
            ))}
          </div>
        )}

        <div className="mt-10 flex flex-wrap justify-center gap-4">
          <a href={restaurant.primaryCTA.href} className="btn-primary">
            {restaurant.primaryCTA.label}
          </a>
          {restaurant.phone && (
            <a href={`tel:${restaurant.phone.replace(/\s/g, "")}`} className="btn-ghost">
              Call the Restaurant
            </a>
          )}
          {emailUrl && (
            <a href={emailUrl} target="_blank" rel="noopener noreferrer" className="btn-ghost">
              Email Us
            </a>
          )}
        </div>
      </div>
    </section>
  );
}

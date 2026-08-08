import { permanentRedirect } from "next/navigation";

/**
 * Preserve the restaurant-services URL used by active outreach campaigns and
 * production configuration while the canonical corporate site lives at `/`.
 */
export default function RestaurantServicesCompatibilityPage() {
  permanentRedirect("/");
}

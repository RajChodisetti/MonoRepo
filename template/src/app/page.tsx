import type { Metadata } from "next";
import CinematicTemplate from "@/templates/cinematic/CinematicTemplate";
import AuroraTemplate from "@/templates/aurora/AuroraTemplate";
import {
  loadRestaurant,
  parseRestaurantIndex,
  getRestaurantCount,
} from "@/lib/adapters/restaurantLoader";
import { resolveTemplate } from "@/lib/templateConfig";
import { buildMetadata as buildCinematicMetadata } from "@/templates/cinematic/seo";
import { buildAuroraMetadata } from "@/templates/aurora/seo";

interface PageProps {
  searchParams: Promise<{ id?: string; template?: string }>;
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const params = await searchParams;
  const index = parseRestaurantIndex(params.id);
  const template = resolveTemplate(params.template);

  try {
    const restaurant = await loadRestaurant(index);
    return template === "2"
      ? buildAuroraMetadata(restaurant)
      : buildCinematicMetadata(restaurant);
  } catch {
    return { title: "Restaurant not found" };
  }
}

export default async function HomePage({ searchParams }: PageProps) {
  const params = await searchParams;
  const index = parseRestaurantIndex(params.id);
  const template = resolveTemplate(params.template);

  try {
    const restaurant = await loadRestaurant(index);

    if (template === "2") {
      return <AuroraTemplate restaurant={restaurant} />;
    }
    return <CinematicTemplate restaurant={restaurant} />;
  } catch (err) {
    const message = err instanceof Error ? err.message : "Unknown error";
    const total = await getRestaurantCount();
    return (
      <main className="flex min-h-screen items-center justify-center px-6 text-center">
        <div>
          <h1 className="text-3xl font-semibold text-white">Could not load restaurant</h1>
          <p className="mt-4 text-white/60">{message}</p>
          <p className="mt-2 text-sm text-white/40">
            Use ?id=0–{total - 1} · Template {template} active
          </p>
        </div>
      </main>
    );
  }
}

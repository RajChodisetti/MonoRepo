import type { Metadata } from "next";
import CinematicTemplate from "@/templates/cinematic/CinematicTemplate";
import AuroraTemplate from "@/templates/aurora/AuroraTemplate";
import ElysianTemplate from "@/templates/elysian/ElysianTemplate";
import {
  loadRestaurant,
  loadRestaurantFromApiOnly,
  parseRestaurantIndex,
  getRestaurantCount,
} from "@/lib/adapters/restaurantLoader";
import { resolveTemplate } from "@/lib/templateConfig";
import { buildMetadata as buildCinematicMetadata } from "@/templates/cinematic/seo";
import { buildAuroraMetadata } from "@/templates/aurora/seo";
import { buildElysianMetadata, buildElysianJsonLd } from "@/templates/elysian/seo";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

interface PageProps {
  searchParams: Promise<{ id?: string; template?: string }>;
}

async function loadForTemplate(index: number, template: "1" | "2" | "3") {
  if (template === "3") return loadRestaurantFromApiOnly(index);
  return loadRestaurant(index);
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const params = await searchParams;
  const index = parseRestaurantIndex(params.id);
  const template = resolveTemplate(params.template);

  try {
    const restaurant = await loadForTemplate(index, template);
    if (template === "3") return buildElysianMetadata(restaurant);
    if (template === "2") return buildAuroraMetadata(restaurant);
    return buildCinematicMetadata(restaurant);
  } catch {
    return { title: "Restaurant not found" };
  }
}

export default async function HomePage({ searchParams }: PageProps) {
  const params = await searchParams;
  const index = parseRestaurantIndex(params.id);
  const template = resolveTemplate(params.template);

  try {
    const restaurant = await loadForTemplate(index, template);

    if (template === "3") {
      const jsonLd = buildElysianJsonLd(restaurant);
      return (
        <>
          <script
            type="application/ld+json"
            dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
          />
          <ElysianTemplate restaurant={restaurant} />
        </>
      );
    }
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
            {template === "3" ? " (Elysian requires API — set NEXT_PUBLIC_API_URL)" : ""}
          </p>
        </div>
      </main>
    );
  }
}

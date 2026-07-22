import type { Metadata } from "next";
import CinematicTemplate from "@/templates/cinematic/CinematicTemplate";
import AuroraTemplate from "@/templates/aurora/AuroraTemplate";
import ElysianTemplate from "@/templates/elysian/ElysianTemplate";
import FoodieTemplate from "@/templates/foodie/FoodieTemplate";
import {
  loadRestaurant,
  loadRestaurantFromApiOnly,
  loadSignedDemo,
  parseRestaurantIndex,
  getRestaurantCount,
} from "@/lib/adapters/restaurantLoader";
import { resolveTemplate } from "@/lib/templateConfig";
import { buildMetadata as buildCinematicMetadata } from "@/templates/cinematic/seo";
import { buildAuroraMetadata } from "@/templates/aurora/seo";
import { buildElysianMetadata, buildElysianJsonLd } from "@/templates/elysian/seo";
import { buildFoodieMetadata } from "@/templates/foodie/seo";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

interface PageProps {
  searchParams: Promise<{ id?: string; template?: string; slug?: string; token?: string }>;
}

async function loadForTemplate(
  index: number,
  template: "1" | "2" | "3",
  slug?: string,
  token?: string,
) {
  if (slug || token) {
    if (!slug || !token) throw new Error("The signed demo link is incomplete.");
    return loadSignedDemo(slug, token, index);
  }
  if (template === "3") return loadRestaurantFromApiOnly(index);
  return loadRestaurant(index);
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const params = await searchParams;
  const index = parseRestaurantIndex(params.id);
  const template = resolveTemplate(params.template);

  // Foodie is a static template (no restaurant data yet).
  if (template === "4") return buildFoodieMetadata();

  try {
    const restaurant = await loadForTemplate(index, template, params.slug, params.token);
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

  // Foodie is a static landing template — render directly without loading data.
  if (template === "4") {
    return <FoodieTemplate />;
  }

  try {
    const restaurant = await loadForTemplate(index, template, params.slug, params.token);

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
    const signedDemo = Boolean(params.slug || params.token);
    const total = signedDemo ? 0 : await getRestaurantCount();
    return (
      <main className="flex min-h-screen items-center justify-center px-6 text-center">
        <div>
          <h1 className="text-3xl font-semibold text-white">Could not load restaurant</h1>
          <p className="mt-4 text-white/60">{message}</p>
          {signedDemo ? (
            <p className="mt-2 text-sm text-white/40">
              This demo may be unpublished, expired, or opened with an invalid token.
            </p>
          ) : (
            <p className="mt-2 text-sm text-white/40">
              Use ?id=0–{total - 1} · Template {template} active
              {template === "3" ? " (Elysian requires API — set NEXT_PUBLIC_API_URL)" : ""}
            </p>
          )}
        </div>
      </main>
    );
  }
}

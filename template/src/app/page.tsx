import type { Metadata } from "next";
import CinematicTemplate from "@/templates/cinematic/CinematicTemplate";
import AuroraTemplate from "@/templates/aurora/AuroraTemplate";
import ElysianTemplate from "@/templates/elysian/ElysianTemplate";
import {
  loadRestaurant,
  loadRestaurantFromApiOnly,
  loadRestaurantByID,
  loadSignedDemo,
  parseRestaurantIndex,
  getRestaurantCount,
} from "@/lib/adapters/restaurantLoader";
import { resolveTemplate } from "@/lib/templateConfig";
import { buildMetadata as buildCinematicMetadata } from "@/templates/cinematic/seo";
import { buildAuroraMetadata } from "@/templates/aurora/seo";
import { buildElysianMetadata, buildElysianJsonLd } from "@/templates/elysian/seo";
import DemoEngagementTracker from "@/components/DemoEngagementTracker";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

interface PageProps {
  searchParams: Promise<{ id?: string; restaurant_id?: string; template?: string; slug?: string; token?: string }>;
}

async function loadForTemplate(
  index: number,
  template: "1" | "2" | "3",
  slug?: string,
  token?: string,
  restaurantID?: string,
) {
  if (slug || token) {
    if (!slug || !token) throw new Error("The signed demo link is incomplete.");
    return loadSignedDemo(slug, token, index);
  }
  if (restaurantID) return loadRestaurantByID(restaurantID);
  if (template === "3") return loadRestaurantFromApiOnly(index);
  return loadRestaurant(index);
}

export async function generateMetadata({ searchParams }: PageProps): Promise<Metadata> {
  const params = await searchParams;
  const index = parseRestaurantIndex(params.id);
  const template = resolveTemplate(params.template);

  try {
    const restaurant = await loadForTemplate(index, template, params.slug, params.token, params.restaurant_id);
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
    const restaurant = await loadForTemplate(index, template, params.slug, params.token, params.restaurant_id);

    if (template === "3") {
      const jsonLd = buildElysianJsonLd(restaurant);
      return (
        <>
          <DemoEngagementTracker slug={params.slug} demoToken={params.token} templateID="3" />
          <script
            type="application/ld+json"
            dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
          />
          <ElysianTemplate restaurant={restaurant} />
        </>
      );
    }
    if (template === "2") {
      return (
        <>
          <DemoEngagementTracker slug={params.slug} demoToken={params.token} templateID="2" />
          <AuroraTemplate restaurant={restaurant} />
        </>
      );
    }
    return (
      <>
        <DemoEngagementTracker slug={params.slug} demoToken={params.token} templateID="1" />
        <CinematicTemplate restaurant={restaurant} />
      </>
    );
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

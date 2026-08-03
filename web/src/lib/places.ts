/**
 * Types + MonoRepo SEO API helpers for the marketing web BFF.
 * Places credentials live only on the Go API — never in this app.
 */

export type RestaurantSearchResult = {
  placeId: string;
  name: string;
  address: string;
  rating?: number;
  userRatingCount?: number;
  source: "monorepo" | "places";
};

export type RestaurantDetails = {
  placeId: string;
  name: string;
  address: string;
  phone?: string;
  email?: string;
  website?: string;
  mapsUri?: string;
  rating?: number;
  userRatingCount?: number;
  priceLevel?: string;
  businessStatus?: string;
  types?: string[];
  editorialSummary?: string;
  source: "monorepo" | "places";
  media?: PlaceMedia;
};

export type MediaCard = {
  kind: string;
  label: string;
  subtitle?: string;
  imageUrl?: string;
  photoName?: string;
  href?: string;
};

export type PlaceMedia = {
  menuAndHighlights?: MediaCard[];
  photosAndVideos?: MediaCard[];
  mapsUri?: string;
};

function monorepoBase(): string {
  return (process.env.MONOREPO_API_URL?.trim() || "http://localhost:8080").replace(/\/$/, "");
}

export async function searchSeoRestaurants(
  query: string,
  location = "Australia",
): Promise<{
  results: RestaurantSearchResult[];
  meta: { placesEnabled: boolean; inventoryEnabled: boolean };
}> {
  const q = query.trim();
  if (q.length < 2) {
    return { results: [], meta: { placesEnabled: false, inventoryEnabled: false } };
  }

  const loc = (location || "Australia").trim() || "Australia";
  const params = new URLSearchParams({ q, location: loc });
  const res = await fetch(
    `${monorepoBase()}/api/public/v1/seo/search?${params.toString()}`,
    { cache: "no-store" },
  );
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`SEO search failed (${res.status}): ${text.slice(0, 200)}`);
  }

  const data = (await res.json()) as {
    results?: Array<{
      placeId?: string;
      name?: string;
      address?: string;
      rating?: number;
      userRatingCount?: number;
      source?: string;
    }>;
    meta?: { placesEnabled?: boolean; inventoryEnabled?: boolean };
  };

  const results: RestaurantSearchResult[] = [];
  for (const item of data.results || []) {
    const placeId = String(item.placeId || "").trim();
    if (!placeId) continue;
    results.push({
      placeId,
      name: String(item.name || "Restaurant"),
      address: String(item.address || ""),
      rating: typeof item.rating === "number" ? item.rating : undefined,
      userRatingCount:
        typeof item.userRatingCount === "number" ? item.userRatingCount : undefined,
      source: item.source === "monorepo" ? "monorepo" : "places",
    });
  }

  return {
    results,
    meta: {
      placesEnabled: Boolean(data.meta?.placesEnabled),
      inventoryEnabled: Boolean(data.meta?.inventoryEnabled),
    },
  };
}

export async function getSeoReport(
  placeId: string,
  unlockToken?: string,
): Promise<{
  place: RestaurantDetails;
  report: import("@/lib/report").RestaurantReport;
}> {
  const id = placeId.trim();
  if (!id) {
    throw new Error("Missing placeId");
  }

  const qs = unlockToken?.trim()
    ? `?unlock=${encodeURIComponent(unlockToken.trim())}`
    : "";
  const res = await fetch(
    `${monorepoBase()}/api/public/v1/seo/report/${encodeURIComponent(id)}${qs}`,
    { cache: "no-store" },
  );
  if (res.status === 404) {
    const err = new Error("Restaurant not found");
    (err as Error & { status?: number }).status = 404;
    throw err;
  }
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`SEO report failed (${res.status}): ${text.slice(0, 200)}`);
  }

  return (await res.json()) as {
    place: RestaurantDetails;
    report: import("@/lib/report").RestaurantReport;
  };
}

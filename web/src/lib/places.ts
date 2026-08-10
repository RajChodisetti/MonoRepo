/**
 * Types + MonoRepo SEO API helpers for the marketing web BFF.
 * Places credentials live only on the Go API — never in this app.
 */

export type RestaurantSearchResult = {
  placeId: string;
  name: string;
  address: string;
  latitude?: number;
  longitude?: number;
  rating?: number;
  userRatingCount?: number;
  attributions?: PlaceAttribution[];
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
  latitude?: number;
  longitude?: number;
  rating?: number;
  userRatingCount?: number;
  priceLevel?: string;
  businessStatus?: string;
  types?: string[];
  primaryType?: string;
  editorialSummary?: string;
  attributions?: PlaceAttribution[];
  source: "monorepo" | "places";
  media?: PlaceMedia;
};

export type PlaceAttribution = {
  provider?: string;
  providerUri?: string;
};

export type AuthorAttribution = {
  displayName?: string;
  uri?: string;
  photoUri?: string;
};

export type MediaCard = {
  kind: string;
  label: string;
  subtitle?: string;
  imageUrl?: string;
  photoName?: string;
  href?: string;
  authorAttributions?: AuthorAttribution[];
  googleMapsUri?: string;
  flagContentUri?: string;
};

export type PlaceMedia = {
  menuAndHighlights?: MediaCard[];
  photosAndVideos?: MediaCard[];
  mapsUri?: string;
};

function monorepoBase(): string {
  return (process.env.MONOREPO_API_URL?.trim() || "http://localhost:8080").replace(/\/$/, "");
}

async function throwSEOAPIError(response: Response, operation: string): Promise<never> {
  const text = await response.text().catch(() => "");
  const error = new Error(`${operation} failed (${response.status}): ${text.slice(0, 200)}`) as Error & {
    status?: number;
    retryAfter?: string;
  };
  error.status = response.status;
  const retryAfter = response.headers.get("retry-after")?.trim();
  if (retryAfter) error.retryAfter = retryAfter;
  throw error;
}

export async function searchSeoRestaurants(
  query: string,
  location = "Australia",
  forwardedHeaders: Record<string, string> = {},
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
    { cache: "no-store", headers: forwardedHeaders },
  );
  if (!res.ok) {
    await throwSEOAPIError(res, "SEO search");
  }

  const data = (await res.json()) as {
    results?: Array<{
      placeId?: string;
      name?: string;
      address?: string;
      latitude?: number;
      longitude?: number;
      rating?: number;
      userRatingCount?: number;
      attributions?: Array<{ provider?: unknown; providerUri?: unknown }>;
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
      latitude: typeof item.latitude === "number" ? item.latitude : undefined,
      longitude: typeof item.longitude === "number" ? item.longitude : undefined,
      rating: typeof item.rating === "number" ? item.rating : undefined,
      userRatingCount:
        typeof item.userRatingCount === "number" ? item.userRatingCount : undefined,
      attributions: (item.attributions || []).flatMap((attribution) => {
        const provider =
          typeof attribution.provider === "string" ? attribution.provider.trim() : "";
        const providerUri =
          typeof attribution.providerUri === "string"
            ? attribution.providerUri.trim()
            : "";
        return provider || providerUri ? [{ provider, providerUri }] : [];
      }),
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
  forwardedHeaders: Record<string, string> = {},
): Promise<{
  place: RestaurantDetails;
  report: import("@/lib/report").RestaurantReport;
}> {
  const id = placeId.trim();
  if (!id) {
    throw new Error("Missing placeId");
  }

  const headers = { ...forwardedHeaders };
  if (unlockToken?.trim()) {
    headers.Authorization = `Bearer ${unlockToken.trim()}`;
  }
  const res = await fetch(`${monorepoBase()}/api/public/v1/seo/report/${encodeURIComponent(id)}`, {
    cache: "no-store",
    headers,
  });
  if (res.status === 404) {
    const err = new Error("Restaurant not found");
    (err as Error & { status?: number }).status = 404;
    throw err;
  }
  if (!res.ok) {
    await throwSEOAPIError(res, "SEO report");
  }

  return (await res.json()) as {
    place: RestaurantDetails;
    report: import("@/lib/report").RestaurantReport;
  };
}

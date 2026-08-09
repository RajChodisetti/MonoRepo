import type { RestaurantContent } from "@/data/types/restaurant";
import { fetchSignedDemo, fetchSiteRestaurant, fetchSiteRestaurantByID, getRestaurantCountFromApi } from "./restaurantSiteApi";

export function parseRestaurantIndex(raw?: string | string[] | null): number {
  const value = Array.isArray(raw) ? raw[0] : raw;
  const parsed = Number.parseInt(value ?? "0", 10);
  return Number.isNaN(parsed) || parsed < 0 ? 0 : parsed;
}

export async function loadRestaurant(index: number): Promise<RestaurantContent> {
  const fromApi = await fetchSiteRestaurant(index);
  if (!fromApi) {
    throw new Error(
      `Restaurant index ${index} was not found via the restaurant API.`,
    );
  }
  return fromApi;
}

export async function loadRestaurantFromApiOnly(index: number): Promise<RestaurantContent> {
  const fromApi = await fetchSiteRestaurant(index);
  if (!fromApi) {
    throw new Error(
      `Restaurant index ${index} not found via API. Ensure NEXT_PUBLIC_API_URL is set and the API is running.`,
    );
  }
  return fromApi;
}

export async function loadRestaurantByID(restaurantID: string): Promise<RestaurantContent> {
  const fromApi = await fetchSiteRestaurantByID(restaurantID);
  if (!fromApi) {
    throw new Error(`Restaurant ${restaurantID} was not found via the restaurant API.`);
  }
  return fromApi;
}

export async function loadSignedDemo(slug: string, token: string, index: number): Promise<RestaurantContent> {
  const fromApi = await fetchSignedDemo(slug, token, index);
  if (!fromApi) {
    throw new Error("This signed demo is invalid, unpublished, or expired.");
  }
  return fromApi;
}

export async function getRestaurantCount(): Promise<number> {
  const apiCount = await getRestaurantCountFromApi();
  return apiCount ?? 0;
}

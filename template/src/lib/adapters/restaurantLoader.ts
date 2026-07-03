import type { RestaurantContent } from "@/data/types/restaurant";
import { fetchSiteRestaurant, getRestaurantCountFromApi } from "./restaurantSiteApi";
import { loadRestaurant as loadRestaurantFromJson, getRestaurantCount as getJsonRestaurantCount } from "./scrapedRestaurant";

export { parseRestaurantIndex } from "./scrapedRestaurant";

export async function loadRestaurant(index: number): Promise<RestaurantContent> {
  const fromApi = await fetchSiteRestaurant(index);
  if (fromApi) return fromApi;
  return loadRestaurantFromJson(index);
}

export async function getRestaurantCount(): Promise<number> {
  const apiCount = await getRestaurantCountFromApi();
  if (apiCount != null && apiCount > 0) return apiCount;
  return getJsonRestaurantCount();
}

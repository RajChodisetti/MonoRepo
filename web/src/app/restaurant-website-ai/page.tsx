import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { restaurantWebsite } from "@/content/products/restaurant-website";

export const metadata: Metadata = {
  title: restaurantWebsite.meta.title,
  description: restaurantWebsite.meta.description,
};

export default function RestaurantWebsiteAiPage() {
  return <ProductPage config={restaurantWebsite} />;
}

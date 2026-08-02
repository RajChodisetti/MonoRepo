import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { restaurantSeo } from "@/content/products/restaurant-seo";

export const metadata: Metadata = {
  title: restaurantSeo.meta.title,
  description: restaurantSeo.meta.description,
};

export default function RestaurantSeoPage() {
  return <ProductPage config={restaurantSeo} />;
}

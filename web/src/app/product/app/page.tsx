import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { brandedRestaurantApp } from "@/content/products/branded-restaurant-app";

export const metadata: Metadata = {
  title: brandedRestaurantApp.meta.title,
  description: brandedRestaurantApp.meta.description,
};

export default function BrandedRestaurantAppPage() {
  return <ProductPage config={brandedRestaurantApp} />;
}

import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { kitchenTablet } from "@/content/products/kitchen-tablet";

export const metadata: Metadata = {
  title: kitchenTablet.meta.title,
  description: kitchenTablet.meta.description,
};

export default function KitchenTabletPage() {
  return <ProductPage config={kitchenTablet} />;
}

import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { smartUpsells } from "@/content/products/smart-upsells";

export const metadata: Metadata = {
  title: smartUpsells.meta.title,
  description: smartUpsells.meta.description,
};

export default function SmartUpsellsPage() {
  return <ProductPage config={smartUpsells} />;
}

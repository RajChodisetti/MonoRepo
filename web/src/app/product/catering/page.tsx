import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { catering } from "@/content/products/catering";

export const metadata: Metadata = {
  title: catering.meta.title,
  description: catering.meta.description,
};

export default function CateringPage() {
  return <ProductPage config={catering} />;
}

import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { onlineOrdering } from "@/content/products/online-ordering";

export const metadata: Metadata = {
  title: onlineOrdering.meta.title,
  description: onlineOrdering.meta.description,
};

export default function OnlineOrderingPage() {
  return <ProductPage config={onlineOrdering} />;
}

import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { delivery } from "@/content/products/delivery";

export const metadata: Metadata = {
  title: delivery.meta.title,
  description: delivery.meta.description,
};

export default function DeliveryPage() {
  return <ProductPage config={delivery} />;
}

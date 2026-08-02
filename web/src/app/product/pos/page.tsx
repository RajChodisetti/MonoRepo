import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { posIntegrations } from "@/content/products/pos-integrations";

export const metadata: Metadata = {
  title: posIntegrations.meta.title,
  description: posIntegrations.meta.description,
};

export default function PosIntegrationsPage() {
  return <ProductPage config={posIntegrations} />;
}

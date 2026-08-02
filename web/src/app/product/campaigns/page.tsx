import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { marketingCampaigns } from "@/content/products/marketing-campaigns";

export const metadata: Metadata = {
  title: marketingCampaigns.meta.title,
  description: marketingCampaigns.meta.description,
};

export default function MarketingCampaignsPage() {
  return <ProductPage config={marketingCampaigns} />;
}

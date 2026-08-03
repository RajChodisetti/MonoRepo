import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { loyaltyRewards } from "@/content/products/loyalty-rewards";

export const metadata: Metadata = {
  title: loyaltyRewards.meta.title,
  description: loyaltyRewards.meta.description,
};

export default function LoyaltyRewardsPage() {
  return <ProductPage config={loyaltyRewards} />;
}

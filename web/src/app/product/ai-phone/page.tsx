import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { aiPhoneOrdering } from "@/content/products/ai-phone-ordering";

export const metadata: Metadata = {
  title: aiPhoneOrdering.meta.title,
  description: aiPhoneOrdering.meta.description,
};

export default function AiPhoneOrderingPage() {
  return <ProductPage config={aiPhoneOrdering} />;
}

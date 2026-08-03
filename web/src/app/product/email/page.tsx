import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { emailTextMarketing } from "@/content/products/email-text-marketing";

export const metadata: Metadata = {
  title: emailTextMarketing.meta.title,
  description: emailTextMarketing.meta.description,
};

export default function EmailTextMarketingPage() {
  return <ProductPage config={emailTextMarketing} />;
}

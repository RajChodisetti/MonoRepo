import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { reviewsEngine } from "@/content/products/reviews-engine";

export const metadata: Metadata = {
  title: reviewsEngine.meta.title,
  description: reviewsEngine.meta.description,
};

export default function ReviewsEnginePage() {
  return <ProductPage config={reviewsEngine} />;
}

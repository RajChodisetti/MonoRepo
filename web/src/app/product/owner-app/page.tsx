import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { ownerApp } from "@/content/products/owner-app";

export const metadata: Metadata = {
  title: ownerApp.meta.title,
  description: ownerApp.meta.description,
};

export default function OwnerAppPage() {
  return <ProductPage config={ownerApp} />;
}

import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { listingsManagement } from "@/content/products/listings-management";

export const metadata: Metadata = {
  title: listingsManagement.meta.title,
  description: listingsManagement.meta.description,
};

export default function ListingsManagementPage() {
  return <ProductPage config={listingsManagement} />;
}

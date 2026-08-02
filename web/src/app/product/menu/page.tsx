import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { onlineMenu } from "@/content/products/online-menu";

export const metadata: Metadata = {
  title: onlineMenu.meta.title,
  description: onlineMenu.meta.description,
};

export default function OnlineMenuPage() {
  return <ProductPage config={onlineMenu} />;
}

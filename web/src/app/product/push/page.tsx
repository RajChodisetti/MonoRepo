import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { pushNotifications } from "@/content/products/push-notifications";

export const metadata: Metadata = {
  title: pushNotifications.meta.title,
  description: pushNotifications.meta.description,
};

export default function PushNotificationsPage() {
  return <ProductPage config={pushNotifications} />;
}

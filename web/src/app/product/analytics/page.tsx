import type { Metadata } from "next";
import ProductPage from "@/components/product/ProductPage";
import { reportingAnalytics } from "@/content/products/reporting-analytics";

export const metadata: Metadata = {
  title: reportingAnalytics.meta.title,
  description: reportingAnalytics.meta.description,
};

export default function ReportingAnalyticsPage() {
  return <ProductPage config={reportingAnalytics} />;
}

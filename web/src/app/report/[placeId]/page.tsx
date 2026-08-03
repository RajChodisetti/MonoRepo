import type { Metadata } from "next";
import { Suspense } from "react";
import ReportClient from "@/components/report/ReportClient";
import ScanExperience from "@/components/report/ScanExperience";

type Props = { params: Promise<{ placeId: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { placeId } = await params;
  return {
    title: `AI report · ${decodeURIComponent(placeId)} | Tuvi`,
    description: "Personalized Google visibility report for your restaurant.",
  };
}

export default async function ReportPage({ params }: Props) {
  const { placeId } = await params;
  return (
    <Suspense fallback={<ScanExperience restaurantName="Your restaurant" fetchComplete={false} />}>
      <ReportClient placeId={decodeURIComponent(placeId)} />
    </Suspense>
  );
}

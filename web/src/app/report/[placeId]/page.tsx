import type { Metadata } from "next";
import { Suspense } from "react";
import ReportClient from "@/components/report/ReportClient";
import ScanExperience from "@/components/report/ScanExperience";
import { parsePreviewCoordinates } from "@/lib/report-preview";

type Props = {
  params: Promise<{ placeId: string }>;
  searchParams: Promise<Record<string, string | string[] | undefined>>;
};

function first(v: string | string[] | undefined): string {
  if (Array.isArray(v)) return (v[0] || "").trim();
  return (v || "").trim();
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { placeId } = await params;
  return {
    title: `Digital footprint · ${decodeURIComponent(placeId)} | Tuvi`,
    description: "Personalized digital-footprint and Google visibility report for your restaurant.",
  };
}

export default async function ReportPage({ params, searchParams }: Props) {
  const { placeId: rawId } = await params;
  const sp = await searchParams;
  const placeId = decodeURIComponent(rawId);
  const name = first(sp.name) || "Restaurant";
  const address = first(sp.address) || undefined;
  const { latitude, longitude } = parsePreviewCoordinates(first(sp.lat), first(sp.lng));

  return (
    <Suspense
      fallback={
        <ScanExperience
          restaurantName={name}
          address={address}
          placeId={placeId}
          latitude={latitude}
          longitude={longitude}
          fetchComplete={false}
        />
      }
    >
      <ReportClient placeId={placeId} />
    </Suspense>
  );
}

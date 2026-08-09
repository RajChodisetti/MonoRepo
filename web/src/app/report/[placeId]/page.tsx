import type { Metadata } from "next";
import { Suspense } from "react";
import ReportClient from "@/components/report/ReportClient";
import ScanExperience from "@/components/report/ScanExperience";

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
  const latRaw = Number(first(sp.lat));
  const lngRaw = Number(first(sp.lng));
  const latitude = Number.isFinite(latRaw) ? latRaw : undefined;
  const longitude = Number.isFinite(lngRaw) ? lngRaw : undefined;

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

import type { Metadata } from "next";
import { Suspense } from "react";
import ReportClient from "@/components/report/ReportClient";

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
    <Suspense
      fallback={
        <div className="mx-auto max-w-3xl px-6 py-24 text-center">
          <p className="font-display text-2xl font-semibold text-ink">Building your AI report…</p>
        </div>
      }
    >
      <ReportClient placeId={decodeURIComponent(placeId)} />
    </Suspense>
  );
}

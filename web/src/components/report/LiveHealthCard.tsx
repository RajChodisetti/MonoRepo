"use client";

import Image from "next/image";
import type { HealthMetric } from "@/lib/report";
import LockedBlur from "@/components/report/LockedBlur";

const MUTED = "#9b9690";
const FOREST = "#1f4d3a";

function MetricBar({ metric }: { metric: HealthMetric }) {
  const pct = Math.round(Math.min(Math.max(metric.value, 0), 1) * 100);
  return (
    <li>
      <div className="mb-1 flex items-center justify-between gap-2">
        <span className="truncate text-[12.5px] font-semibold text-[#111111]">{metric.label}</span>
        <span className="shrink-0 text-[11px] font-semibold" style={{ color: metric.statusColor }}>
          {metric.status}
        </span>
      </div>
      <div className="h-1.5 overflow-hidden rounded-full bg-[#ebe7e2]">
        <div
          className="h-full rounded-full transition-[width] duration-500"
          style={{ width: `${pct}%`, backgroundColor: metric.statusColor }}
        />
      </div>
    </li>
  );
}

function isWebsiteMetric(metric: HealthMetric) {
  const key = (metric.key || "").toLowerCase();
  const label = metric.label.toLowerCase();
  return key === "website" || label.includes("website design") || label.includes("website");
}

function readinessCopy(score: number, label: string) {
  if (score >= 75) return { badge: "Strong demand", hint: label };
  if (score >= 50) return { badge: "Building demand", hint: label };
  return { badge: "Leaking bookings", hint: label };
}

export default function LiveHealthCard({
  restaurantName,
  overallScore,
  overallLabel,
  overallColor,
  metrics,
  locked = false,
}: {
  restaurantName: string;
  overallScore: number;
  overallLabel: string;
  overallColor: string;
  metrics: HealthMetric[];
  locked?: boolean;
}) {
  const websiteIdx = metrics.findIndex(isWebsiteMetric);
  const splitAt = websiteIdx >= 0 ? websiteIdx + 1 : Math.min(3, metrics.length);
  const visibleMetrics = metrics.slice(0, splitAt);
  const gatedMetrics = metrics.slice(splitAt);
  const readiness = readinessCopy(overallScore, overallLabel);

  return (
    <div className="overflow-hidden rounded-[18px] bg-white shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <div className="relative h-[140px] w-full">
        <Image
          src="https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?auto=format&fit=crop&w=800&q=80"
          alt=""
          fill
          className="object-cover"
          sizes="320px"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent" />
        <div className="absolute inset-x-0 bottom-0 flex items-end justify-between gap-2 px-3.5 pb-3">
          <div className="min-w-0">
            <p className="text-[11px] font-medium uppercase tracking-[0.08em] text-white/70">
              Venue pulse
            </p>
            <p className="truncate text-[15px] font-semibold tracking-[-0.02em] text-white">
              {restaurantName}
            </p>
          </div>
          <span
            className="shrink-0 rounded-full px-2.5 py-1 text-[11px] font-semibold text-white ring-1 ring-white/30"
            style={{ backgroundColor: `${overallColor}cc` }}
          >
            {readiness.badge}
          </span>
        </div>
      </div>

      <div className="px-3.5 pb-4 pt-3.5">
        <div className="grid grid-cols-3 gap-2">
          <div className="rounded-xl bg-[#f5f3ef] px-2 py-2.5 text-center">
            <p className="text-[18px] font-bold leading-none tracking-[-0.03em] text-[#111111]">
              {overallScore}
            </p>
            <p className="mt-1 text-[10px] font-medium leading-tight" style={{ color: MUTED }}>
              Booking readiness
            </p>
          </div>
          <div className="rounded-xl bg-[#f5f3ef] px-2 py-2.5 text-center">
            <p className="text-[18px] font-bold leading-none tracking-[-0.03em]" style={{ color: FOREST }}>
              {Math.max(1, Math.round(overallScore / 8))}
            </p>
            <p className="mt-1 text-[10px] font-medium leading-tight" style={{ color: MUTED }}>
              Open gaps
            </p>
          </div>
          <div className="rounded-xl bg-[#f5f3ef] px-2 py-2.5 text-center">
            <p className="text-[13px] font-bold leading-tight tracking-[-0.02em] text-[#111111]">
              {readiness.hint}
            </p>
            <p className="mt-1 text-[10px] font-medium leading-tight" style={{ color: MUTED }}>
              Guest path
            </p>
          </div>
        </div>

        <p className="mt-4 text-[11px] font-semibold uppercase tracking-[0.07em]" style={{ color: MUTED }}>
          What guests hit first
        </p>
        <ul className="mt-2.5 space-y-2.5">
          {visibleMetrics.map((metric) => (
            <MetricBar key={metric.key || metric.label} metric={metric} />
          ))}
        </ul>

        {gatedMetrics.length > 0 ? (
          <LockedBlur
            locked={locked}
            label="Confirm email for the full venue pulse"
            className="mt-3 rounded-xl"
          >
            <ul className="space-y-2.5 pt-1">
              {gatedMetrics.map((metric) => (
                <MetricBar key={metric.key || metric.label} metric={metric} />
              ))}
            </ul>
          </LockedBlur>
        ) : null}
      </div>
    </div>
  );
}

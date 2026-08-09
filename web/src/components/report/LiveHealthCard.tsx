"use client";

import type { HealthMetric } from "@/lib/report";
import LockedBlur from "@/components/report/LockedBlur";

const MUTED = "#9b9690";
const FOREST = "#1f4d3a";

function MetricBar({ metric }: { metric: HealthMetric }) {
  const pct = Math.round(Math.min(Math.max(metric.value, 0), 1) * 100);
  const score = typeof metric.score === "number" || metric.score ? String(metric.score) : "—";
  const points = typeof metric.max === "number" && metric.max > 0 ? `${score}/${metric.max}` : score;
  return (
    <li>
      <div className="mb-1 flex items-start justify-between gap-2">
        <span className="truncate text-[12.5px] font-semibold text-[#111111]">{metric.label}</span>
        <span className="flex shrink-0 items-baseline gap-1.5">
          <span className="text-[11px] font-bold tabular-nums text-[#111111]" aria-label={`${points} weighted points`}>
            {points}
          </span>
          <span className="text-[10px] font-semibold" style={{ color: metric.statusColor }}>
            {metric.status}
          </span>
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

export default function LiveHealthCard({
  restaurantName,
  overallScore,
  overallLabel,
  overallColor,
  metrics,
  partial = false,
  locked = false,
}: {
  restaurantName: string;
  overallScore: number;
  overallLabel: string;
  overallColor: string;
  metrics: HealthMetric[];
  partial?: boolean;
  locked?: boolean;
}) {
  const websiteIdx = metrics.findIndex(isWebsiteMetric);
  const splitAt = websiteIdx >= 0 ? websiteIdx + 1 : Math.min(3, metrics.length);
  const visibleMetrics = metrics.slice(0, splitAt);
  const gatedMetrics = metrics.slice(splitAt);
  const openGaps = metrics.filter(
    (metric) => !metric.status.toLowerCase().includes("assessed") && metric.value < 0.65,
  ).length;
  const assessedWeight = metrics
    .filter((metric) => !metric.status.toLowerCase().includes("assessed"))
    .reduce((total, metric) => total + (metric.max || 0), 0);

  return (
    <div className="overflow-hidden rounded-[18px] bg-white shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <div className="relative h-[118px] w-full bg-[linear-gradient(135deg,#173f31_0%,#28674f_58%,#d8e9df_160%)]">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_85%_15%,rgba(255,255,255,0.2),transparent_36%)]" />
        <div className="absolute inset-x-0 bottom-0 flex items-end justify-between gap-2 px-3.5 pb-3">
          <div className="min-w-0">
            <p className="text-[11px] font-medium uppercase tracking-[0.08em] text-white/70">
              Weighted digital footprint
            </p>
            <p className="truncate text-[15px] font-semibold tracking-[-0.02em] text-white">
              {restaurantName}
            </p>
          </div>
          <span
            className="shrink-0 rounded-full px-2.5 py-1 text-[11px] font-semibold text-white ring-1 ring-white/30"
            style={{ backgroundColor: partial ? "#6b7280" : `${overallColor}cc` }}
          >
            {partial ? "Partial score" : overallLabel}
          </span>
        </div>
      </div>

      <div className="px-3.5 pb-4 pt-3.5">
        <div className="grid grid-cols-3 gap-2">
          <div className="rounded-xl bg-[#f5f3ef] px-2 py-2.5 text-center">
            <p className="text-[18px] font-bold leading-none tracking-[-0.03em] text-[#111111]">
              {overallScore}<span className="text-[11px] font-semibold text-[#77716b]">/100</span>
            </p>
            <p className="mt-1 text-[10px] font-medium leading-tight" style={{ color: MUTED }}>
              Overall score
            </p>
          </div>
          <div className="rounded-xl bg-[#f5f3ef] px-2 py-2.5 text-center">
            <p className="text-[18px] font-bold leading-none tracking-[-0.03em]" style={{ color: FOREST }}>
              {openGaps}
            </p>
            <p className="mt-1 text-[10px] font-medium leading-tight" style={{ color: MUTED }}>
              Assessed below 65%
            </p>
          </div>
          <div className="rounded-xl bg-[#f5f3ef] px-2 py-2.5 text-center">
            <p className="text-[13px] font-bold leading-tight tracking-[-0.02em] text-[#111111]">
              {assessedWeight}<span className="text-[10px] font-semibold text-[#77716b]">/100</span>
            </p>
            <p className="mt-1 text-[10px] font-medium leading-tight" style={{ color: MUTED }}>
              Fully assessed
            </p>
          </div>
        </div>

        <p className="mt-4 text-[11px] font-semibold uppercase tracking-[0.07em]" style={{ color: MUTED }}>
          Score contribution by signal
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

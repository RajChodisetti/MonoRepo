"use client";

import type { HealthMetric } from "@/lib/report";
import LockedBlur from "@/components/report/LockedBlur";

const MUTED = "#9b9690";

function ProgressRing({
  value,
  size,
  strokeWidth,
  color,
}: {
  value: number;
  size: number;
  strokeWidth: number;
  color: string;
}) {
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference * (1 - Math.min(Math.max(value, 0), 1));

  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="-rotate-90" aria-hidden="true">
      <circle cx={size / 2} cy={size / 2} r={radius} fill="none" stroke="#ebe7e2" strokeWidth={strokeWidth} />
      <circle
        cx={size / 2}
        cy={size / 2}
        r={radius}
        fill="none"
        stroke={color}
        strokeWidth={strokeWidth}
        strokeLinecap="round"
        strokeDasharray={circumference}
        strokeDashoffset={offset}
      />
    </svg>
  );
}

function MetricRow({ metric }: { metric: HealthMetric }) {
  return (
    <li className="flex items-center gap-2.5">
      <div className="relative flex h-6 w-6 shrink-0 items-center justify-center">
        <ProgressRing value={metric.value} size={24} strokeWidth={3} color={metric.statusColor} />
      </div>
      <div className="min-w-0 flex-1 text-left">
        <p className="truncate text-[12.5px] font-semibold leading-tight text-[#111111]">{metric.label}</p>
        <p className="text-[11.5px] font-semibold leading-tight" style={{ color: metric.statusColor }}>
          {metric.status}
        </p>
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

  return (
    <div className="rounded-[18px] bg-white px-4 pb-4 pt-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <div className="flex items-center gap-2.5">
        <div
          className="flex h-[30px] w-[30px] shrink-0 items-center justify-center rounded-full bg-[#e8f1eb] text-[12px] font-bold uppercase text-primary"
          aria-hidden="true"
        >
          {restaurantName.trim().charAt(0) || "R"}
        </div>
        <p className="truncate text-[15px] font-semibold tracking-[-0.02em] text-[#111111]">
          {restaurantName}
        </p>
      </div>

      <div className="relative mx-auto mt-4 flex h-[138px] w-[138px] items-center justify-center">
        <ProgressRing value={overallScore / 100} size={138} strokeWidth={11} color={overallColor} />
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-[40px] font-bold leading-none tracking-[-0.05em] text-[#111111]">
            {overallScore}
          </span>
          <span className="mt-0.5 text-[12px] font-medium" style={{ color: MUTED }}>
            SEO score
          </span>
        </div>
      </div>

      <div className="mt-2.5 text-center">
        <p className="text-[12px] font-medium" style={{ color: MUTED }}>
          Digital footprint score
        </p>
        <p className="mt-0.5 text-[20px] font-bold tracking-[-0.03em]" style={{ color: overallColor }}>
          {overallLabel}
        </p>
      </div>

      <div className="mt-4 border-t border-[#efebe6] pt-4">
        <p className="mb-3 text-[11px] font-semibold uppercase tracking-[0.06em]" style={{ color: MUTED }}>
          Score breakdown
        </p>
        <ul className="space-y-3">
          {visibleMetrics.map((metric) => (
            <MetricRow key={metric.key || metric.label} metric={metric} />
          ))}
        </ul>

        {gatedMetrics.length > 0 ? (
          <LockedBlur
            locked={locked}
            label="Verify email to unlock full scoring"
            className="mt-3 rounded-xl"
          >
            <ul className="space-y-3 pt-1">
              {gatedMetrics.map((metric) => (
                <MetricRow key={metric.key || metric.label} metric={metric} />
              ))}
            </ul>
          </LockedBlur>
        ) : null}
      </div>
    </div>
  );
}

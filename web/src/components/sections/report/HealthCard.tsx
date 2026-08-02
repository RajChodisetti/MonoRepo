import Image from "next/image";

const ACCENT = "#e86a2d";
const FAIR = "#e8a33a";
const MUTED = "#9b9690";

const metrics = [
  { label: "Search results", status: "Poor", statusColor: ACCENT, score: "12/40", value: 0.3 },
  { label: "Guest experience", status: "Fair", statusColor: FAIR, score: "35/40", value: 0.88 },
  { label: "Local listings", status: "Poor", statusColor: ACCENT, score: "4/20", value: 0.2 },
] as const;

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

export default function HealthCard() {
  return (
    <>
      <div className="flex items-center gap-2.5">
        <div className="relative h-[30px] w-[30px] overflow-hidden rounded-full">
          <Image
            src="https://images.unsplash.com/photo-1414235077428-338989a2e8c0?auto=format&fit=crop&w=120&q=80"
            alt=""
            fill
            className="object-cover"
            sizes="30px"
          />
        </div>
        <p className="text-[15px] font-semibold tracking-[-0.02em] text-[#111111]">Your restaurant</p>
      </div>

      <div className="rounded-[18px] bg-white px-3.5 pb-3.5 pt-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
        <div className="relative mx-auto flex h-[138px] w-[138px] items-center justify-center">
          <ProgressRing value={0.36} size={138} strokeWidth={11} color={ACCENT} />
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <span className="text-[40px] font-bold leading-none tracking-[-0.05em] text-[#111111]">36</span>
            <span className="mt-0.5 text-[12px] font-medium" style={{ color: MUTED }}>
              / 100
            </span>
          </div>
        </div>

        <div className="mt-2.5 text-center">
          <p className="text-[12px] font-medium" style={{ color: MUTED }}>
            Website health
          </p>
          <p className="mt-0.5 text-[20px] font-bold tracking-[-0.03em]" style={{ color: ACCENT }}>
            Poor
          </p>
        </div>

        <ul className="mt-4 space-y-3">
          {metrics.map((metric) => (
            <li key={metric.label} className="flex items-center gap-2.5">
              <div className="relative flex h-6 w-6 shrink-0 items-center justify-center">
                <ProgressRing value={metric.value} size={24} strokeWidth={3} color={metric.statusColor} />
              </div>
              <div className="min-w-0 flex-1 text-left">
                <p className="truncate text-[12.5px] font-semibold leading-tight text-[#111111]">{metric.label}</p>
                <p className="text-[11.5px] font-semibold leading-tight" style={{ color: metric.statusColor }}>
                  {metric.status}
                </p>
              </div>
              <span className="shrink-0 text-[11.5px] font-medium tabular-nums" style={{ color: MUTED }}>
                {metric.score}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </>
  );
}

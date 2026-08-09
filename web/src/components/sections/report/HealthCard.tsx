import Image from "next/image";

const FOREST = "#1f4d3a";
const AMBER = "#c45c1a";
const MUTED = "#8a8580";

const channels = [
  { label: "Your website", value: 28, tone: FOREST },
  { label: "Google / Maps", value: 41, tone: "#2f6b52" },
  { label: "Marketplaces", value: 31, tone: AMBER },
] as const;

const signals = [
  { label: "Online menu", detail: "Live · 86 dishes", ok: true },
  { label: "Table requests", detail: "12 this week", ok: true },
  { label: "Phone line", detail: "4 missed calls", ok: false },
] as const;

export default function HealthCard() {
  return (
    <div className="overflow-hidden rounded-[18px] bg-white shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <div className="relative h-[152px] w-full sm:h-[160px]">
        <Image
          src="https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?auto=format&fit=crop&w=800&q=80"
          alt=""
          fill
          className="object-cover"
          sizes="280px"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-black/75 via-black/25 to-transparent" />
        <div className="absolute inset-x-0 bottom-0 flex items-end justify-between gap-2 px-3.5 pb-3">
          <div>
            <p className="text-[11px] font-medium uppercase tracking-[0.08em] text-white/70">
              Venue pulse
            </p>
            <p className="text-[15px] font-semibold tracking-[-0.02em] text-white">Quillnest Kitchen</p>
          </div>
          <span className="rounded-full bg-white/15 px-2.5 py-1 text-[11px] font-semibold text-white ring-1 ring-white/25 backdrop-blur-sm">
            Gildford
          </span>
        </div>
      </div>

      <div className="px-3.5 pb-4 pt-3.5">
        <div className="grid grid-cols-3 gap-2">
          {[
            { value: "186", label: "Menu views" },
            { value: "24", label: "Direct orders" },
            { value: "9", label: "Bookings" },
          ].map((stat) => (
            <div key={stat.label} className="rounded-xl bg-[#f5f3ef] px-2 py-2.5 text-center">
              <p className="text-[18px] font-bold leading-none tracking-[-0.03em] text-[#111111]">
                {stat.value}
              </p>
              <p className="mt-1 text-[10px] font-medium leading-tight" style={{ color: MUTED }}>
                {stat.label}
              </p>
            </div>
          ))}
        </div>

        <p className="mt-4 text-[11px] font-semibold uppercase tracking-[0.07em]" style={{ color: MUTED }}>
          Where guests find you
        </p>
        <ul className="mt-2 space-y-2">
          {channels.map((channel) => (
            <li key={channel.label}>
              <div className="mb-1 flex items-center justify-between gap-2">
                <span className="text-[12px] font-semibold text-[#111111]">{channel.label}</span>
                <span className="text-[11px] font-semibold" style={{ color: channel.tone }}>
                  {channel.value}%
                </span>
              </div>
              <div className="h-1.5 overflow-hidden rounded-full bg-[#ebe7e2]">
                <div
                  className="h-full rounded-full"
                  style={{ width: `${channel.value}%`, backgroundColor: channel.tone }}
                />
              </div>
            </li>
          ))}
        </ul>

        <ul className="mt-4 space-y-2 border-t border-[#efebe6] pt-3">
          {signals.map((signal) => (
            <li key={signal.label} className="flex items-center gap-2.5">
              <span
                className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-bold text-white ${
                  signal.ok ? "bg-[#1f9a4a]" : "bg-[#e5483b]"
                }`}
                aria-hidden
              >
                {signal.ok ? "✓" : "!"}
              </span>
              <div className="min-w-0 flex-1">
                <p className="text-[12.5px] font-semibold text-[#111111]">{signal.label}</p>
                <p className="text-[11px]" style={{ color: MUTED }}>
                  {signal.detail}
                </p>
              </div>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

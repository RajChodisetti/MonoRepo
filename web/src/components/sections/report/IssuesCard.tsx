const issues = [
  {
    title: "Thin coverage in 4 local search pockets",
    description:
      "Diners typing suburb + cuisine phrases are landing on other menus before yours.",
  },
  {
    title: "Menu pages lack clear dish signals",
    description:
      "Google and AI answers need sharper dish names, prices, and dietary cues on your site.",
  },
  {
    title: "Guest proof is hard to find",
    description:
      "Recent reviews and photos aren’t surfaced where new guests decide where to book.",
  },
] as const;

function WarningIcon() {
  return (
    <span
      className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-[#e5483b] text-[13px] font-bold leading-none text-white"
      aria-hidden="true"
    >
      !
    </span>
  );
}

export default function IssuesCard() {
  const [first, ...rest] = issues;

  return (
    <div className="rounded-[18px] bg-white px-3.5 pb-5 pt-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <p className="text-[16px] font-bold leading-snug tracking-[-0.025em] text-[#111111]">
        Priority fixes for your venue
      </p>
      <p className="mt-1.5 text-[12.5px] leading-snug text-[#8a8580]">
        Roughly 30–40 booked tables a month may be walking to nearby venues first.
      </p>

      <ul className="mt-4 space-y-2.5">
        <li className="flex gap-2.5 rounded-2xl bg-[#f3f1ed] px-3 py-3">
          <WarningIcon />
          <div className="min-w-0">
            <p className="text-[13px] font-bold leading-snug text-[#111111]">{first.title}</p>
            <p className="mt-1 text-[12px] leading-snug text-[#8a8580]">{first.description}</p>
          </div>
        </li>
      </ul>

      <div className="relative mt-2.5 overflow-hidden rounded-2xl">
        <ul className="pointer-events-none select-none space-y-2.5 blur-[7px] saturate-50" aria-hidden="true">
          {rest.map((issue) => (
            <li key={issue.title} className="flex gap-2.5 rounded-2xl bg-[#f3f1ed] px-3 py-3">
              <WarningIcon />
              <div className="min-w-0">
                <p className="text-[13px] font-bold leading-snug text-[#111111]">{issue.title}</p>
                <p className="mt-1 text-[12px] leading-snug text-[#8a8580]">{issue.description}</p>
              </div>
            </li>
          ))}
        </ul>
        <div className="absolute inset-0 flex items-center justify-center bg-gradient-to-b from-white/25 via-white/55 to-white/70 px-3">
          <p className="rounded-full bg-white/95 px-3.5 py-1.5 text-center text-[11.5px] font-semibold text-[#111111] shadow-[0_4px_18px_rgba(0,0,0,0.08)]">
            Confirm email for the full Tuvi playbook
          </p>
        </div>
      </div>

      <p className="mt-5 text-center text-[13px] font-medium text-[#8a8580]">
        Close the gaps. Own more direct bookings.
      </p>

      <button
        type="button"
        className="mt-3 w-full rounded-full bg-[#111111] px-4 py-3.5 text-[14px] font-semibold text-white"
      >
        Unlock your Tuvi action list
      </button>
    </div>
  );
}

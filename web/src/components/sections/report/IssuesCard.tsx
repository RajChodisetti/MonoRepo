const issues = [
  {
    title: "Local discovery signals are incomplete",
    description: "Clear cuisine and suburb context helps guests and search engines understand the venue.",
  },
  {
    title: "Title missing primary keyword",
    description: 'A descriptive title such as "Pizza in Springfield" makes the page topic and location clearer.',
  },
  {
    title: "2 images missing alt tags",
    description: "Meaningful alt text improves accessibility and gives website images clearer context.",
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
        Here&apos;s where you can improve
      </p>
      <p className="mt-1.5 text-[12.5px] leading-snug text-[#8a8580]">
        Illustrative preview — the live review does not infer revenue loss without your sales data.
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
            Verify email to see all suggestions
          </p>
        </div>
      </div>

      <p className="mt-5 text-center text-[13px] font-medium text-[#8a8580]">
        Improve your score to drive more sales
      </p>

      <p className="mt-3 w-full rounded-full bg-[#111111] px-4 py-3.5 text-center text-[14px] font-semibold text-white">
        Select a restaurant above for the live review
      </p>
    </div>
  );
}

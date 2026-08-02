const issues = [
  {
    title: "Not ranking in 3 nearby areas",
    description: "Missing keywords to rank nearby for terms competitors are winning with.",
  },
  {
    title: "Title missing primary keyword",
    description: 'Including "Pizza in Springfield" will increase Google rankings.',
  },
  {
    title: "2 images missing alt tags",
    description: "Adding alt tags to all images will boost visibility on Google Maps and Google Images.",
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
  return (
    <div className="rounded-[18px] bg-white px-3.5 pb-5 pt-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <p className="text-[16px] font-bold leading-snug tracking-[-0.025em] text-[#111111]">
        You&apos;re losing $450 a month in sales until you fix these issues:
      </p>

      <ul className="mt-4 space-y-2.5">
        {issues.map((issue) => (
          <li key={issue.title} className="flex gap-2.5 rounded-2xl bg-[#f3f1ed] px-3 py-3">
            <WarningIcon />
            <div className="min-w-0">
              <p className="text-[13px] font-bold leading-snug text-[#111111]">{issue.title}</p>
              <p className="mt-1 text-[12px] leading-snug text-[#8a8580]">{issue.description}</p>
            </div>
          </li>
        ))}
      </ul>

      <p className="mt-5 text-center text-[13px] font-medium text-[#8a8580]">
        Improve your score to drive more sales
      </p>

      <button
        type="button"
        className="mt-3 w-full rounded-full bg-[#111111] px-4 py-3.5 text-[14px] font-semibold text-white"
      >
        Fix it now with AI
      </button>
    </div>
  );
}

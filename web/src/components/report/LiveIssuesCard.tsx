"use client";

import type { ReportIssue } from "@/lib/report";
import LockedBlur from "@/components/report/LockedBlur";

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

function IssueRow({ issue }: { issue: ReportIssue }) {
  return (
    <li className="flex gap-2.5 rounded-2xl bg-[#f3f1ed] px-3 py-3">
      <WarningIcon />
      <div className="min-w-0">
        <p className="text-[13px] font-bold leading-snug text-[#111111]">{issue.title}</p>
        <p className="mt-1 text-[12px] leading-snug text-[#8a8580]">{issue.description}</p>
      </div>
    </li>
  );
}

export default function LiveIssuesCard({
  issues,
  locked = false,
  onFix,
}: {
  issues: ReportIssue[];
  locked?: boolean;
  onFix?: () => void;
}) {
  const first = issues[0];
  const rest = issues.slice(1);

  return (
    <div className="rounded-[18px] bg-white px-3.5 pb-5 pt-4 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <p className="text-[16px] font-bold leading-snug tracking-[-0.025em] text-[#111111]">
        Here&apos;s where you can improve
      </p>
      <p className="mt-1.5 text-[12.5px] leading-snug text-[#8a8580]">
        These gaps can reduce discovery and direct conversions. Tuvi does not infer a dollar loss without your venue&apos;s sales data.
      </p>

      <ul className="mt-4 space-y-2.5">
        {first ? <IssueRow issue={first} /> : null}
      </ul>

      {locked && !first ? (
        <LockedBlur
          locked
          label="Verify email to see suggested fixes"
          className="mt-4 rounded-2xl"
          onUnlock={onFix}
        >
          <div className="space-y-2.5" aria-hidden="true">
            {[0, 1, 2].map((item) => (
              <div key={item} className="h-[70px] rounded-2xl bg-[#e7e2db]" />
            ))}
          </div>
        </LockedBlur>
      ) : null}

      {rest.length > 0 ? (
        <LockedBlur
          locked={locked}
          label="Verify email to see all suggestions"
          className="mt-2.5 rounded-2xl"
          onUnlock={onFix}
        >
          <ul className="space-y-2.5">
            {rest.map((issue) => (
              <IssueRow key={issue.title} issue={issue} />
            ))}
          </ul>
        </LockedBlur>
      ) : null}

      <p className="mt-5 text-center text-[13px] font-medium text-[#8a8580]">
        Improve your score to drive more sales
      </p>

      <button
        type="button"
        onClick={onFix}
        className="mt-3 min-h-11 w-full cursor-pointer rounded-full bg-[#111111] px-4 py-3.5 text-[14px] font-semibold text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      >
        {locked ? "Verify email — unlock suggested fixes" : "See how Tuvi can help"}
      </button>
    </div>
  );
}

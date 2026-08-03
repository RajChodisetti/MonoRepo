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
  estimatedMonthlyLoss,
  locked = false,
  onFix,
}: {
  issues: ReportIssue[];
  estimatedMonthlyLoss: number;
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
        You&apos;re losing ${estimatedMonthlyLoss.toLocaleString()} a month in sales until these are fixed.
      </p>

      <ul className="mt-4 space-y-2.5">
        {first ? <IssueRow issue={first} /> : null}
      </ul>

      {rest.length > 0 ? (
        <LockedBlur
          locked={locked}
          label="Verify email to see all suggestions"
          className="mt-2.5 rounded-2xl"
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
        className="mt-3 w-full cursor-pointer rounded-full bg-[#111111] px-4 py-3.5 text-[14px] font-semibold text-white"
      >
        {locked ? "Verify email — unlock free AI fixes" : "Fix it now with AI"}
      </button>
    </div>
  );
}

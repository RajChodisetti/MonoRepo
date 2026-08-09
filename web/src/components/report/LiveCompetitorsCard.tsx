"use client";

import type { CompetitorRow } from "@/lib/report";
import LockedBlur from "@/components/report/LockedBlur";

const MUTED = "#9b9690";

function Star() {
  return (
    <svg viewBox="0 0 12 12" className="h-3 w-3 fill-[#f5b942]" aria-hidden="true">
      <path d="M6 1.1 7.3 4.4l3.6.3-2.8 2.3.9 3.5L6 8.6 3 10.5l.9-3.5L1.1 4.7l3.6-.3L6 1.1Z" />
    </svg>
  );
}

/** Strip bucket fractions like 23/25 → show clean score label. */
function cleanScore(score: string): string {
  const raw = (score || "").trim();
  const m = raw.match(/^(\d+(?:\.\d+)?)\s*\/\s*\d+(?:\.\d+)?$/);
  if (m) return m[1];
  return raw.replace(/\/\d+(?:\.\d+)?$/, "").trim() || raw;
}

function strengthLabel(score: string, scoreColor: string): { text: string; color: string } {
  const n = Number.parseFloat(cleanScore(score));
  if (!Number.isFinite(n)) return { text: "Steady demand", color: scoreColor };
  if (n >= 35 || (n <= 25 && n >= 20)) return { text: "Winning search", color: scoreColor };
  if (n >= 15) return { text: "Steady demand", color: scoreColor };
  return { text: "Building up", color: scoreColor };
}

export default function LiveCompetitorsCard({
  rows,
  locked = false,
  onUnlock,
}: {
  rows: CompetitorRow[];
  locked?: boolean;
  onUnlock?: () => void;
}) {
  return (
    <div className="rounded-[18px] bg-white px-3.5 pb-4 pt-3.5 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <p className="text-[15px] font-bold leading-snug tracking-[-0.02em] text-[#111111]">
        Local search snapshot
      </p>
      <p className="mt-1 text-[12px] leading-snug" style={{ color: MUTED }}>
        How nearby venues show up when diners look for a table tonight.
      </p>

      <LockedBlur locked={locked} label="Confirm email for the full local map" className="mt-3 rounded-xl">
        <ul className="space-y-3">
          {rows.map((row, index) => {
            const strength = row.highlight
              ? { text: "Your venue", color: row.scoreColor }
              : strengthLabel(row.score, row.scoreColor);
            return (
              <li key={`${row.rank}-${row.name}`}>
                {index === rows.length - 1 && rows.length > 1 ? (
                  <div className="mb-2 flex flex-col items-start gap-0.5 pl-[18px]" aria-hidden="true">
                    <span className="h-1 w-1 rounded-full bg-[#cfcac4]" />
                    <span className="h-1 w-1 rounded-full bg-[#cfcac4]" />
                    <span className="h-1 w-1 rounded-full bg-[#cfcac4]" />
                  </div>
                ) : null}
                <div className="flex items-center gap-2.5">
                  <span
                    className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-[11px] font-semibold ${
                      row.highlight ? "bg-[#e8e4df] text-[#5a5550]" : "bg-[#f0eeea] text-[#7a7570]"
                    }`}
                  >
                    {row.rank}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-[13px] font-semibold text-[#111111]">{row.name}</p>
                    <p className="mt-0.5 flex items-center gap-1 text-[12px]" style={{ color: MUTED }}>
                      <Star />
                      {row.rating}
                    </p>
                  </div>
                  <span className="shrink-0 text-[11px] font-semibold" style={{ color: strength.color }}>
                    {strength.text}
                  </span>
                </div>
              </li>
            );
          })}
        </ul>
      </LockedBlur>

      {locked ? (
        <button
          type="button"
          onClick={onUnlock}
          className="mt-3.5 w-full cursor-pointer rounded-full bg-[#111111] px-4 py-2.5 text-[12.5px] font-semibold text-white"
        >
          Unlock your local search snapshot
        </button>
      ) : (
        <p className="mt-3.5 text-center text-[11.5px] font-medium" style={{ color: MUTED }}>
          Full local snapshot unlocked
        </p>
      )}
    </div>
  );
}

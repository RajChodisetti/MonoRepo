"use client";

import type { CompetitorRow, CompetitorScan } from "@/lib/report";

const MUTED = "#746f69";

function Star() {
  return (
    <svg viewBox="0 0 12 12" className="h-3 w-3 fill-[#f5b942]" aria-hidden="true">
      <path d="M6 1.1 7.3 4.4l3.6.3-2.8 2.3.9 3.5L6 8.6 3 10.5l.9-3.5L1.1 4.7l3.6-.3L6 1.1Z" />
    </svg>
  );
}

function displayScore(row: CompetitorRow): string {
  const score = String(row.score ?? "-").trim() || "-";
  if (score.includes("/")) return score;
  return row.scoreMax ? `${score}/${row.scoreMax}` : score;
}

function safeExternalHref(raw?: string): string | undefined {
  const value = (raw || "").trim();
  if (!value) return undefined;
  try {
    const parsed = new URL(value.startsWith("//") ? `https:${value}` : value);
    return parsed.protocol === "http:" || parsed.protocol === "https:"
      ? parsed.toString()
      : undefined;
  } catch {
    return undefined;
  }
}

function CompetitorResult({ row }: { row: CompetitorRow }) {
  return (
    <li className="rounded-xl border border-[#ece8e3] bg-[#fbfaf8] p-3">
      <div className="flex items-start gap-2.5">
        <span
          className={`flex min-h-8 min-w-8 shrink-0 items-center justify-center rounded-md px-1.5 text-[11px] font-semibold ${
            row.highlight ? "bg-[#dfece5] text-primary" : "bg-[#f0eeea] text-[#6f6a65]"
          }`}
        >
          {row.rank}
        </span>
        <div className="min-w-0 flex-1">
          <p className="truncate text-[13px] font-semibold text-[#111111]">{row.name}</p>
          <p className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11.5px]" style={{ color: MUTED }}>
            <span className="inline-flex items-center gap-1">
              <Star /> {row.rating || "-"}
            </span>
            {typeof row.userRatingCount === "number" ? (
              <span>{row.userRatingCount.toLocaleString()} reviews</span>
            ) : null}
            {typeof row.distanceKm === "number" ? <span>{row.distanceKm.toFixed(1)} km away</span> : null}
          </p>
        </div>
        <div className="shrink-0 text-right">
          <p className="text-[12px] font-bold text-primary">{displayScore(row)}</p>
          <p className="mt-0.5 text-[9px] font-semibold uppercase tracking-[0.05em]" style={{ color: MUTED }}>
            visibility
          </p>
        </div>
      </div>
      {row.reasons?.length ? (
        <p className="mt-2 border-t border-[#ece8e3] pt-2 text-[11px] leading-relaxed" style={{ color: MUTED }}>
          {row.reasons.slice(0, 2).join(" · ")}
        </p>
      ) : null}
      {row.attributions?.length ? (
        <div
          className="mt-2 flex flex-wrap gap-x-2 gap-y-1 border-t border-[#ece8e3] pt-2 text-[12px] leading-4"
          style={{ color: MUTED }}
          translate="no"
        >
          {row.attributions.map((attribution, index) => {
            const providerHref = safeExternalHref(attribution.providerUri);
            const provider = attribution.provider?.trim() || "Data source";
            return providerHref ? (
              <a
                key={`${provider}-${index}`}
                href={providerHref}
                target="_blank"
                rel="noreferrer"
                className="underline underline-offset-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
              >
                Source: {provider}
              </a>
            ) : (
              <span key={`${provider}-${index}`}>Source: {provider}</span>
            );
          })}
        </div>
      ) : null}
    </li>
  );
}

function LockedTeaser({ onUnlock }: { onUnlock?: () => void }) {
  return (
    <div className="mt-3">
      <div className="space-y-2" aria-hidden="true">
        {[0, 1, 2].map((index) => (
          <div key={index} className="flex items-center gap-3 rounded-xl border border-[#ece8e3] bg-[#fbfaf8] p-3 blur-[5px]">
            <span className="h-8 w-8 rounded-md bg-[#ddd8d1]" />
            <span className="h-3 flex-1 rounded-full bg-[#d7d2cc]" />
            <span className="h-3 w-12 rounded-full bg-[#ddd8d1]" />
          </div>
        ))}
      </div>
      <button
        type="button"
        onClick={onUnlock}
        className="mt-3.5 min-h-11 w-full cursor-pointer rounded-full bg-[#111111] px-4 py-2.5 text-[12.5px] font-semibold text-white focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      >
        Confirm email to reveal who is ahead
      </button>
    </div>
  );
}

export default function LiveCompetitorsCard({
  rows,
  scan,
  locked = false,
  onUnlock,
}: {
  rows: CompetitorRow[] | null | undefined;
  scan?: CompetitorScan;
  locked?: boolean;
  onUnlock?: () => void;
}) {
  const realRows = scan?.rows ?? rows ?? [];
  const radius = scan?.radiusKm || 10;
  const cuisine = scan?.cuisine?.trim() || "same-cuisine restaurants";

  return (
    <div className="rounded-[18px] bg-white px-3.5 pb-4 pt-3.5 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <p className="text-[15px] font-bold leading-snug tracking-[-0.02em] text-[#111111]">
        Nearby Google visibility
      </p>
      <p className="mt-1 text-[12px] leading-snug" style={{ color: MUTED }}>
        {cuisine} within {radius} km with stronger comparable public signals.
      </p>
      <p className="mt-1 text-[10.5px] leading-snug" style={{ color: MUTED }}>
        This is an observable visibility comparison, not a claim about a fixed Google Search position.
      </p>

      {!locked && typeof scan?.currentScore === "number" ? (
        <div className="mt-3 flex items-center justify-between rounded-xl bg-[#f7f4ef] px-3 py-2.5 text-[11.5px]">
          <span className="font-semibold text-ink">Your comparable visibility</span>
          <span className="font-bold text-primary">{scan.currentScore}/100</span>
        </div>
      ) : null}

      {locked ? (
        <LockedTeaser onUnlock={onUnlock} />
      ) : scan?.currentRestaurantLeading ? (
        <div className="mt-3 rounded-xl border border-[#b9ddc7] bg-[#edf8f1] p-4">
          <p className="text-[13px] font-bold text-[#176b3a]">You lead this observed comparison set</p>
          <p className="mt-1 text-[11.5px] leading-relaxed text-[#356348]">
            None of the eligible nearby listings scored higher on the comparable Google visibility signals checked now.
          </p>
        </div>
      ) : realRows.length ? (
        <ul className="mt-3 space-y-2">
          {realRows.map((row) => (
            <CompetitorResult key={`${row.placeId || row.name}-${row.rank}`} row={row} />
          ))}
        </ul>
      ) : (
        <div className="mt-3 rounded-xl border border-[#ece8e3] bg-[#fbfaf8] p-4">
          <p className="text-[13px] font-semibold text-ink">No honest comparison available</p>
          <p className="mt-1 text-[11.5px] leading-relaxed" style={{ color: MUTED }}>
            Tuvi did not receive enough eligible nearby same-cuisine listings to name a stronger competitor.
          </p>
        </div>
      )}

      {!locked && scan?.notice ? (
        <p className="mt-3 text-[10.5px] leading-relaxed" style={{ color: MUTED }}>
          {scan.notice}
        </p>
      ) : null}

      <a
        href="https://www.google.com/maps"
        target="_blank"
        rel="noreferrer"
        className="mt-3 inline-block text-[12px] font-normal text-[#5e5e5e] underline-offset-2 hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
        translate="no"
      >
        Google Maps
      </a>
    </div>
  );
}

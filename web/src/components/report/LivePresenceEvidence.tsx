"use client";

import LockedBlur from "@/components/report/LockedBlur";
import type { MenuEvidence, SocialPresence } from "@/lib/report";

function statusLabel(status?: string) {
  switch ((status || "").toLowerCase()) {
    case "present":
    case "confirmed":
      return "Confirmed";
    case "not_found":
    case "missing":
      return "Not found";
    case "partial":
      return "Partial";
    default:
      return "Not confirmed";
  }
}

function statusClass(status?: string) {
  const normalized = (status || "").toLowerCase();
  if (normalized === "present" || normalized === "confirmed") {
    return "bg-[#e8f6ee] text-[#176b3a]";
  }
  if (normalized === "partial") return "bg-[#fff2dc] text-[#8a5200]";
  return "bg-[#f1ede7] text-[#6f675e]";
}

function EvidencePanel({
  title,
  status,
  score,
  children,
}: {
  title: string;
  status?: string;
  score?: string;
  children: React.ReactNode;
}) {
  return (
    <article className="rounded-2xl border border-border bg-white p-4 sm:p-5">
      <div className="flex flex-wrap items-start justify-between gap-2">
        <div>
          <h3 className="text-[15px] font-bold tracking-[-0.02em] text-ink">{title}</h3>
          {score ? <p className="mt-0.5 text-[12px] font-semibold text-primary">{score}</p> : null}
        </div>
        <span className={`rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-[0.06em] ${statusClass(status)}`}>
          {statusLabel(status)}
        </span>
      </div>
      <div className="mt-3">{children}</div>
    </article>
  );
}

export default function LivePresenceEvidence({
  menu,
  social,
  locked,
  onUnlock,
}: {
  menu?: MenuEvidence;
  social?: SocialPresence;
  locked: boolean;
  onUnlock: () => void;
}) {
  const menuDetail = menu?.rationale || menu?.explanation;

  return (
    <div className="grid gap-3 sm:grid-cols-2">
      <EvidencePanel title="Menu assessment" status={menu?.status}>
        <div className="flex flex-wrap gap-1.5 text-[10.5px] font-semibold">
          <span className={`rounded-full px-2 py-1 ${menu?.hasWebsiteLink ? "bg-[#e8f6ee] text-[#176b3a]" : "bg-[#f1ede7] text-[#6f675e]"}`}>
            Public menu link {menu?.hasWebsiteLink ? "found" : "not found"}
          </span>
          <span className={`rounded-full px-2 py-1 ${menu?.hasStructuredData ? "bg-[#e8f6ee] text-[#176b3a]" : "bg-[#f1ede7] text-[#6f675e]"}`}>
            Structured menu {menu?.hasStructuredData ? "found" : "not found"}
          </span>
        </div>
        <LockedBlur
          locked={locked}
          label="Confirm email for menu findings"
          className="mt-3 min-h-[82px] rounded-xl"
          onUnlock={onUnlock}
        >
          <div className="rounded-xl bg-[#f7f4ef] p-3 text-[12px] leading-relaxed text-muted">
            <p>{menuDetail || "Tuvi checked the restaurant website for a verifiable menu link and structured menu evidence."}</p>
            {menu?.menuUrl ? (
              <a
                href={menu.menuUrl}
                target="_blank"
                rel="noreferrer"
                className="mt-2 inline-flex min-h-11 items-center font-semibold text-primary underline underline-offset-4"
              >
                Open confirmed menu source
              </a>
            ) : null}
          </div>
        </LockedBlur>
        <p className="mt-3 text-[10.5px] leading-relaxed text-muted">
          Generic listing photos are kept as photo evidence, not treated as proof of a menu.
        </p>
      </EvidencePanel>

      <EvidencePanel
        title="Social media presence"
        status={social?.status}
        score={typeof social?.score === "number" ? `${social.score}/${social.max || 0} weighted points` : undefined}
      >
        <LockedBlur
          locked={locked}
          label="Confirm email for social findings"
          className="min-h-[112px] rounded-xl"
          onUnlock={onUnlock}
        >
          <div className="rounded-xl bg-[#f7f4ef] p-3 text-[12px] leading-relaxed text-muted">
            {locked ? (
              <p>Verified social profile details unlock after email confirmation.</p>
            ) : social?.profiles?.length ? (
              <ul className="space-y-2">
                {social.profiles.map((profile) => (
                  <li key={`${profile.platform}-${profile.url}`}>
                    <a
                      href={profile.url}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex min-h-11 items-center font-semibold text-primary underline underline-offset-4"
                    >
                      {profile.platform}{profile.handle ? ` · ${profile.handle}` : ""}
                    </a>
                  </li>
                ))}
              </ul>
            ) : (
              <p>No public social profile was confirmed from links on the restaurant website.</p>
            )}
            {social?.rationale ? <p className="mt-2">{social.rationale}</p> : null}
          </div>
        </LockedBlur>
        <p className="mt-3 text-[10.5px] leading-relaxed text-muted">
          Profiles are confirmed from the venue&apos;s own public website; Tuvi does not guess handles from similar names.
        </p>
      </EvidencePanel>
    </div>
  );
}

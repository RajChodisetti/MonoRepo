const TEAL = "#0f7a5a";
const AMBER = "#c45c1a";
const MUTED = "#9b9690";

const rows = [
  {
    rank: "1st",
    name: "Driftpane Dining",
    rating: "4.7",
    label: "Winning search",
    scoreColor: TEAL,
    highlight: false,
  },
  {
    rank: "2nd",
    name: "Marrowbay Table",
    rating: "4.5",
    label: "Winning search",
    scoreColor: TEAL,
    highlight: false,
  },
  {
    rank: "3rd",
    name: "Pebblewick Room",
    rating: "4.4",
    label: "Steady demand",
    scoreColor: TEAL,
    highlight: false,
  },
  {
    rank: "8th",
    name: "Your venue",
    rating: "4.1",
    label: "Gap to close",
    scoreColor: AMBER,
    highlight: true,
  },
] as const;

function Star() {
  return (
    <svg viewBox="0 0 12 12" className="h-3 w-3 fill-[#f5b942]" aria-hidden="true">
      <path d="M6 1.1 7.3 4.4l3.6.3-2.8 2.3.9 3.5L6 8.6 3 10.5l.9-3.5L1.1 4.7l3.6-.3L6 1.1Z" />
    </svg>
  );
}

export default function CompetitorsCard() {
  return (
    <div className="rounded-[18px] bg-white px-3.5 pb-4 pt-3.5 shadow-[0_1px_3px_rgba(0,0,0,0.04)]">
      <p className="text-[15px] font-bold leading-snug tracking-[-0.02em] text-[#111111]">
        Local search snapshot
      </p>
      <p className="mt-1 text-[12px] leading-snug" style={{ color: MUTED }}>
        How nearby venues show up when diners look for a table tonight.
      </p>

      <ul className="mt-3.5 space-y-3">
        {rows.map((row, index) => (
          <li key={row.rank}>
            {index === 3 ? (
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
              <span className="shrink-0 text-[11px] font-semibold" style={{ color: row.scoreColor }}>
                {row.label}
              </span>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}

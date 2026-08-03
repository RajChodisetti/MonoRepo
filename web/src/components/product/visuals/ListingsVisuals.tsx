import Image from "next/image";

export function ListingMapCardVisual() {
  return (
    <div className="relative flex h-full min-h-[400px] w-full max-w-[340px] items-center justify-center">
      {/* Map panel */}
      <div className="absolute inset-x-2 inset-y-6 overflow-hidden rounded-[22px] bg-[#e8e4dc] sm:inset-x-4">
        <svg className="absolute inset-0 h-full w-full opacity-50" aria-hidden="true">
          {Array.from({ length: 8 }).map((_, i) => (
            <line
              key={`h-${i}`}
              x1="0"
              y1={i * 48}
              x2="400"
              y2={i * 48}
              stroke="rgba(0,0,0,0.08)"
              strokeWidth="1"
            />
          ))}
          {Array.from({ length: 6 }).map((_, i) => (
            <line
              key={`v-${i}`}
              x1={i * 60}
              y1="0"
              x2={i * 60}
              y2="400"
              stroke="rgba(0,0,0,0.08)"
              strokeWidth="1"
            />
          ))}
        </svg>
        <div className="absolute left-1/2 top-[42%] -translate-x-1/2 -translate-y-1/2">
          <svg viewBox="0 0 24 32" className="h-10 w-7 drop-shadow-md" aria-hidden="true">
            <path
              d="M12 0C6.5 0 2 4.4 2 9.8c0 7.4 10 22.2 10 22.2s10-14.8 10-22.2C22 4.4 17.5 0 12 0z"
              fill="#0f271f"
            />
            <circle cx="12" cy="10" r="3.2" fill="#fff" />
          </svg>
        </div>
      </div>

      {/* Floating listing card */}
      <div className="relative z-10 mt-8 w-[88%] max-w-[280px] rounded-2xl bg-white p-3.5 shadow-[0_14px_36px_rgba(0,0,0,0.12)]">
        <div className="flex gap-3">
          <div className="relative h-14 w-14 shrink-0 overflow-hidden rounded-xl">
            <Image src="/owners/pasta.jpg" alt="" fill className="object-cover" sizes="56px" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[15px] font-bold text-[#0f271f]">Ashley&apos;s Cafe</p>
            <div className="mt-1 flex items-center gap-1.5">
              <span className="text-[11px] tracking-tight text-[#2f6b54]">★★★★★</span>
              <span className="text-[12px] font-semibold text-[#0f271f]">4.4</span>
            </div>
            <div className="mt-2 space-y-1.5">
              <div className="h-1.5 w-[90%] rounded-full bg-[#dce6dd]" />
              <div className="h-1.5 w-[65%] rounded-full bg-[#dce6dd]" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

const SYNC_PLATFORMS = [
  { name: "Google", color: "#4285F4" },
  { name: "Facebook", color: "#1877F2" },
  { name: "Yelp", color: "#D32323" },
  { name: "TripAdvisor", color: "#34E0A1" },
  { name: "DoorDash", color: "#FF3008" },
  { name: "Uber Eats", color: "#06C167" },
] as const;

export function ListingsSyncedVisual() {
  return (
    <div className="w-full max-w-[300px] rounded-[22px] bg-white/95 p-4 shadow-[0_16px_40px_rgba(0,0,0,0.16)] backdrop-blur-sm sm:p-5">
      <div className="flex items-center gap-2">
        <span className="h-2.5 w-2.5 rounded-full bg-[#2f6b54]" />
        <p className="text-[14px] font-semibold text-[#0f271f]">All listings synced</p>
      </div>
      <ul className="mt-4 space-y-2.5">
        {SYNC_PLATFORMS.map((platform) => (
          <li
            key={platform.name}
            className="flex items-center justify-between gap-3 rounded-xl bg-[#f2ecdf] px-3 py-2.5"
          >
            <div className="flex items-center gap-2.5">
              <span
                className="flex h-7 w-7 items-center justify-center rounded-full text-[11px] font-bold text-white"
                style={{ backgroundColor: platform.color }}
              >
                {platform.name.charAt(0)}
              </span>
              <span className="text-[13px] font-semibold text-[#0f271f]">{platform.name}</span>
            </div>
            <span className="rounded-full bg-[#2f6b54]/15 px-2.5 py-1 text-[10px] font-bold tracking-wide text-[#174c3a]">
              SYNCED
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}

export function AddressFixCardsVisual() {
  return (
    <div className="flex w-full max-w-[300px] flex-col gap-3">
      <div className="rounded-2xl bg-white px-4 py-3.5 shadow-[0_10px_28px_rgba(0,0,0,0.08)]">
        <p className="text-[15px] font-bold text-[#0f271f]">The Pizza Spot</p>
        <p className="mt-1 text-[13px] text-[#555]">123 Main Street</p>
        <p className="text-[13px] text-[#555]">New York</p>
      </div>
      <div className="rounded-2xl bg-white px-4 py-3.5 shadow-[0_10px_28px_rgba(0,0,0,0.08)]">
        <p className="text-[15px] font-bold text-[#0f271f]">The Pizza Spot</p>
        <p className="mt-1 text-[13px] text-[#555]">
          123 Main <span className="font-semibold text-[#c45c5c]">St.</span>
        </p>
        <p className="text-[13px] text-[#555]">
          <span className="font-semibold text-[#c45c5c]">NY, USA</span>
        </p>
      </div>
    </div>
  );
}

export function ListingsExpertsPhotoVisual() {
  return (
    <div className="absolute inset-0">
      <Image
        src="/guides/interview.jpg"
        alt=""
        fill
        className="object-cover object-[center_20%]"
        sizes="(max-width: 1024px) 90vw, 560px"
        quality={90}
      />
      <div
        className="absolute inset-0"
        style={{
          background:
            "linear-gradient(180deg, rgba(10,10,10,0.55) 0%, rgba(10,10,10,0.2) 45%, rgba(10,10,10,0.45) 100%)",
        }}
        aria-hidden="true"
      />
    </div>
  );
}

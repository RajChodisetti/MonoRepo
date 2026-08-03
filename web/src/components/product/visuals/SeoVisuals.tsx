import Image from "next/image";

export function SeoScoreVisual() {
  const score = 94;
  const radius = 54;
  const circumference = 2 * Math.PI * radius;
  const progress = (score / 100) * circumference;

  return (
    <div className="flex w-full max-w-[300px] items-center justify-center py-4 sm:max-w-[320px]">
      <div className="w-full rounded-[28px] bg-white/90 p-6 shadow-[0_16px_40px_rgba(0,0,0,0.12)] backdrop-blur-md sm:rounded-[32px] sm:p-7">
        <div className="mx-auto flex w-[140px] flex-col items-center">
          <div className="relative h-[140px] w-[140px]">
            <svg viewBox="0 0 140 140" className="h-full w-full -rotate-90">
              <circle
                cx="70"
                cy="70"
                r={radius}
                fill="none"
                stroke="rgba(47,191,85,0.2)"
                strokeWidth="10"
              />
              <circle
                cx="70"
                cy="70"
                r={radius}
                fill="none"
                stroke="#2f6b54"
                strokeWidth="10"
                strokeLinecap="round"
                strokeDasharray={`${progress} ${circumference}`}
              />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <p className="text-[1.65rem] font-bold tracking-tight text-[#0f271f]">
                {score}
                <span className="text-[1rem] font-semibold text-[#888]"> / 100</span>
              </p>
            </div>
          </div>
          <p className="mt-3 text-[13px] font-medium text-[#6b6b6b]">Online health grade</p>
          <p className="mt-0.5 text-[1.35rem] font-bold text-[#0f271f]">Good</p>
        </div>

        <div className="mt-5 space-y-3 border-t border-black/[0.06] pt-4">
          {[
            { label: "Search results", score: "37/40" },
            { label: "Website Experience", score: "36/40" },
          ].map((row) => (
            <div key={row.label} className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2.5">
                <span className="flex h-6 w-6 items-center justify-center rounded-full bg-[#2f6b54]/15 text-[#174c3a]">
                  <svg viewBox="0 0 16 16" className="h-3.5 w-3.5" aria-hidden="true">
                    <path
                      d="M4 8.2 6.6 10.8 12 5.4"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.8"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </span>
                <div>
                  <p className="text-[13px] font-medium text-[#555]">{row.label}</p>
                  <p className="text-[14px] font-bold text-[#0f271f]">Good</p>
                </div>
              </div>
              <p className="text-[14px] font-semibold text-[#0f271f]">{row.score}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

/** Green-card search mock (SEO AI card) — same pattern as website AI search */
export function SeoAiSearchVisual() {
  return (
    <div className="w-full max-w-[320px]">
      <div className="flex items-center gap-2.5 rounded-2xl bg-white px-3.5 py-3.5 shadow-[0_8px_28px_rgba(0,0,0,0.12)]">
        <svg viewBox="0 0 20 20" className="h-4 w-4 shrink-0 text-[#666]" aria-hidden="true">
          <circle cx="9" cy="9" r="5.5" fill="none" stroke="currentColor" strokeWidth="1.6" />
          <path d="m13.5 13.5 3 3" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
        </svg>
        <span className="text-[14px] font-medium text-[#888]">Restaurants near me...</span>
      </div>

      <div className="relative mt-3.5 rounded-[18px] bg-white/20 p-3">
        <div className="flex items-center gap-3 rounded-2xl bg-white p-3.5 shadow-[0_10px_28px_rgba(0,0,0,0.14)]">
          <div className="relative h-12 w-12 shrink-0 overflow-hidden rounded-xl">
            <Image src="/owners/pizza.jpg" alt="" fill className="object-cover" sizes="48px" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[15px] font-semibold text-[#0f271f]">Your restaurant</p>
            <div className="mt-1 flex items-center gap-1">
              <span className="text-[12px] font-semibold text-[#0f271f]">4.5</span>
              <span className="text-[12px] text-[#f5a623]">★★★★★</span>
            </div>
            <div className="mt-2 h-1.5 w-[85%] rounded-full bg-[#dce6dd]" />
          </div>
        </div>
        <div className="mt-2.5 space-y-2.5 px-1 opacity-45">
          <div className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-lg bg-white/50" />
            <div className="h-2.5 flex-1 rounded-full bg-white/50" />
          </div>
          <div className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-lg bg-white/35" />
            <div className="h-2.5 w-[70%] rounded-full bg-white/35" />
          </div>
        </div>
      </div>
    </div>
  );
}

export function GoogleUpdateVisual() {
  return (
    <video
      className="absolute inset-0 h-full w-full object-cover object-center"
      autoPlay
      muted
      loop
      playsInline
      preload="metadata"
      aria-label="We track every Google update so your rankings never slip"
    >
      <source
        src="/hf_20260727_183617_e357c57c-6231-4751-be87-15b9235fbadc.mp4"
        type="video/mp4"
      />
    </video>
  );
}

const EXPERT_AVATARS = [
  "/people/james.jpg",
  "/people/maria.jpg",
  "/people/priya.jpg",
  "/people/david.jpg",
  "/people/lena.jpg",
  "/people/kevin.jpg",
] as const;

export function ExpertsAvatarsVisual() {
  return (
    <div className="relative flex w-full items-center justify-center pb-2 pt-6">
      <div className="pointer-events-none absolute inset-0 opacity-40" aria-hidden="true">
        <svg className="h-full w-full" viewBox="0 0 400 200" preserveAspectRatio="none">
          {[40, 70, 100, 130, 160].map((y) => (
            <path
              key={y}
              d={`M0 ${y} Q 100 ${y - 18} 200 ${y} T 400 ${y}`}
              fill="none"
              stroke="rgba(0,0,0,0.08)"
              strokeWidth="1"
            />
          ))}
        </svg>
      </div>
      <div className="relative flex -space-x-3">
        {EXPERT_AVATARS.map((src, i) => (
          <div
            key={src}
            className="relative h-14 w-14 overflow-hidden rounded-full border-[3px] border-white shadow-md sm:h-16 sm:w-16"
            style={{ zIndex: EXPERT_AVATARS.length - i }}
          >
            <Image src={src} alt="" fill className="object-cover" sizes="64px" />
          </div>
        ))}
      </div>
    </div>
  );
}

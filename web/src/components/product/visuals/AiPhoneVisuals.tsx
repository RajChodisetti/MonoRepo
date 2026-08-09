import Image from "next/image";

/** Feature-split left — AI phone ordering video (full panel) */
export function AiPhoneMockupVisual() {
  return (
    <video
      className="absolute inset-0 h-full w-full object-cover object-center"
      autoPlay
      muted
      loop
      playsInline
      preload="metadata"
      aria-label="AI phone ordering conversation"
    >
      <source
        src="/hf_20260727_184539_e440cc62-a384-404c-8a3d-71538b7f843c.mp4"
        type="video/mp4"
      />
    </video>
  );
}

export function AiPhoneConversationVisual() {
  return (
    <div className="relative w-full max-w-[300px]">
      <div className="rounded-[22px] bg-white p-4 shadow-[0_16px_40px_rgba(0,0,0,0.14)]">
        <div className="flex items-center gap-2.5">
          <span className="flex h-9 w-9 items-center justify-center rounded-full bg-[#174c3a]">
            <svg viewBox="0 0 20 20" className="h-4 w-4 text-white" aria-hidden="true">
              <path
                d="M4 10c1.5-2 3-3 4.5-3s3 1 4.5 3 3 3 4.5 3"
                fill="none"
                stroke="currentColor"
                strokeWidth="1.6"
                strokeLinecap="round"
              />
            </svg>
          </span>
          <div>
            <p className="text-[13px] font-bold text-[#0f271f]">AI Host - Timber Quince</p>
            <p className="text-[11px] text-[#888]">Call with John Smith</p>
          </div>
        </div>

        <div className="mt-4 space-y-3">
          <div>
            <p className="text-[11px] font-semibold text-[#0f271f]">
              John Smith <span className="font-normal text-[#999]">0:24</span>
            </p>
            <p className="mt-0.5 text-[13px] leading-snug text-[#333]">
              Actually, make the pizza a medium instead
            </p>
          </div>
          <div>
            <p className="text-[11px] font-semibold text-[#174c3a]">
              AI Host <span className="font-normal text-[#999]">0:29</span>
            </p>
            <p className="mt-0.5 text-[13px] leading-snug text-[#333]">
              No problem! switched to a medium pepperoni. Anything else?
            </p>
          </div>
          <div>
            <p className="text-[11px] font-semibold text-[#0f271f]">
              John Smith <span className="font-normal text-[#999]">1:12</span>
            </p>
            <p className="mt-0.5 text-[13px] leading-snug text-[#333]">No, thanks!</p>
          </div>
        </div>
      </div>

      <div className="absolute -bottom-3 right-2 flex items-center gap-2 rounded-full bg-white px-3 py-2 shadow-[0_10px_28px_rgba(0,0,0,0.14)]">
        <span className="flex h-5 w-5 items-center justify-center rounded-full bg-[#2f6b54] text-[10px] font-bold text-white">
          ↻
        </span>
        <div className="text-[11px] font-semibold text-[#0f271f]">
          Order updated{" "}
          <span className="text-[#888]">Large →</span>{" "}
          <span className="rounded-md bg-[#2f6b54] px-1.5 py-0.5 text-white">Medium</span>
        </div>
      </div>
    </div>
  );
}

export function AiPhoneLoyaltyPhotoVisual() {
  return (
    <div className="absolute inset-0">
      <Image
        src="/guides/interview.jpg"
        alt=""
        fill
        className="object-cover object-[center_22%]"
        sizes="(max-width: 1024px) 90vw, 560px"
        quality={90}
      />
      <div
        className="absolute inset-0"
        style={{
          background:
            "linear-gradient(180deg, rgba(10,10,10,0.55) 0%, rgba(10,10,10,0.2) 50%, rgba(10,10,10,0.4) 100%)",
        }}
        aria-hidden="true"
      />
    </div>
  );
}

const AI_PHONE_FOOD = [
  "/menu/pasta-alfredo.jpg",
  "/menu/caesar-salad.jpg",
  "/menu/pepperoni-pizza.jpg",
  "/menu/classic-burger.jpg",
  "/menu/birria-tacos.jpg",
  "/menu/garlic-bread.jpg",
] as const;

export function AiPhoneFoodTilesVisual() {
  return (
    <div className="relative flex h-[260px] w-full items-center justify-center sm:h-[300px]">
      <svg
        className="pointer-events-none absolute inset-0 h-full w-full opacity-35"
        viewBox="0 0 360 280"
        aria-hidden="true"
      >
        {[60, 100, 140].map((r) => (
          <circle
            key={r}
            cx="180"
            cy="150"
            r={r}
            fill="none"
            stroke="rgba(0,0,0,0.12)"
            strokeWidth="1"
          />
        ))}
      </svg>
      <div className="relative grid grid-cols-3 gap-2.5">
        {AI_PHONE_FOOD.map((src, i) => (
          <div
            key={src}
            className="relative h-[72px] w-[72px] overflow-hidden rounded-2xl shadow-md sm:h-20 sm:w-20"
            style={{ transform: `translateY(${(i % 3) * 6 - 6}px)` }}
          >
            <Image src={src} alt="" fill className="object-cover" sizes="80px" />
            <span className="absolute bottom-1 right-1 flex h-5 w-5 items-center justify-center rounded-full bg-white text-[12px] font-bold text-[#0f271f] shadow">
              +
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

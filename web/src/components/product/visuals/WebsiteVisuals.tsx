import Image from "next/image";

export function GyroPreviewVisual() {
  return (
    <div className="flex h-[520px] w-full max-w-[300px] flex-col overflow-hidden rounded-[22px] bg-white sm:h-[560px] sm:max-w-[320px] sm:rounded-[26px]">
      <div className="flex items-center justify-between gap-2 px-3.5 pb-2 pt-3.5 sm:px-4 sm:pt-4">
        <div className="flex min-w-0 items-center gap-2">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#4a90d9] text-[10px] font-bold text-white">
            QK
          </span>
          <span className="min-w-0 leading-tight">
            <span className="block truncate text-[11px] font-bold tracking-[-0.02em] text-[#3b7fc4] sm:text-[12px]">
              QUILLNEST
            </span>
            <span className="block text-[9px] font-medium text-[#7aa8d8]">kitchen · melbourne</span>
          </span>
        </div>
        <div className="flex shrink-0 items-center gap-1.5">
          <span className="rounded-md border border-[#ddd] bg-white px-2.5 py-1 text-[10px] font-semibold text-[#333]">
            Menu
          </span>
          <span className="flex h-7 w-7 flex-col items-center justify-center gap-[3px] rounded-md">
            <span className="block h-[1.5px] w-3.5 rounded bg-[#222]" />
            <span className="block h-[1.5px] w-3.5 rounded bg-[#222]" />
            <span className="block h-[1.5px] w-3.5 rounded bg-[#222]" />
          </span>
        </div>
      </div>

      <div className="relative mx-3 mt-1 h-[210px] shrink-0 overflow-hidden rounded-2xl sm:mx-3.5 sm:h-[240px]">
        <Image src="/resources/resource-website-hero.png" alt="" fill className="object-cover" sizes="420px" />
        <div
          className="absolute inset-x-0 bottom-0 h-[55%]"
          style={{
            background:
              "linear-gradient(to top, rgba(20,20,20,0.92) 0%, rgba(20,20,20,0.72) 45%, transparent 100%)",
          }}
          aria-hidden="true"
        />
        <div className="absolute inset-x-0 bottom-0 p-3.5 sm:p-4">
          <p className="text-[11px] font-medium text-white/90 sm:text-[12px]">
            Neighbourhood favourites, ordered direct
          </p>
          <p className="mt-1 max-w-[18ch] text-[15px] font-bold leading-[1.2] tracking-[-0.02em] text-white sm:text-[17px]">
            Your brand. Your menu. Your guests.
          </p>
        </div>
      </div>

      <div className="mt-3 shrink-0 bg-[#eef5fb] px-3 py-3 sm:px-3.5 sm:py-3.5">
        <div className="grid grid-cols-2 gap-2">
          <span className="rounded-xl bg-[#4a90d9] py-2.5 text-center text-[12px] font-semibold text-white">
            Order pickup
          </span>
          <span className="rounded-xl bg-[#4a90d9] py-2.5 text-center text-[12px] font-semibold text-white">
            Order delivery
          </span>
        </div>
        <p className="mt-2.5 text-center text-[11px] font-medium text-[#4a90d9] underline underline-offset-2">
          Start a catering order ›
        </p>
      </div>

      <div className="mt-3 shrink-0 px-3 sm:px-3.5">
        <div className="flex items-center justify-between">
          <p className="text-[13px] font-bold text-[#0f271f]">Featured</p>
          <span className="rounded-md border border-[#ddd] bg-white px-2 py-1 text-[10px] font-semibold text-[#333]">
            View menu ›
          </span>
        </div>
        <div className="mt-2.5 flex gap-2">
          {["/menu/pepperoni-pizza.jpg", "/menu/pasta-alfredo.jpg", "/menu/classic-burger.jpg"].map((src) => (
            <div
              key={src}
              className="relative h-[100px] w-[90px] shrink-0 overflow-hidden rounded-2xl bg-[#f0f0f0]"
            >
              <Image src={src} alt="" fill className="object-cover" sizes="90px" />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export function AiSearchMockVisual() {
  return (
    <div className="w-full max-w-[320px]">
      <div className="flex items-center gap-2.5 rounded-2xl bg-white px-3.5 py-3.5 shadow-[0_8px_28px_rgba(0,0,0,0.12)]">
        <svg viewBox="0 0 20 20" className="h-4 w-4 shrink-0 text-[#666]" aria-hidden="true">
          <circle cx="9" cy="9" r="5.5" fill="none" stroke="currentColor" strokeWidth="1.6" />
          <path d="m13.5 13.5 3 3" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
        </svg>
        <span className="text-[14px] font-medium text-[#888]">Restaurants near me...</span>
      </div>

      <div className="relative mt-3.5 rounded-[18px] bg-white/25 p-3 backdrop-blur-[1px]">
        <div className="flex items-center gap-3 rounded-2xl bg-white p-3.5 shadow-[0_10px_28px_rgba(0,0,0,0.14)]">
          <div className="relative h-12 w-12 shrink-0 overflow-hidden rounded-xl">
            <Image src="/menu/pepperoni-pizza.jpg" alt="" fill className="object-cover" sizes="48px" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[15px] font-semibold text-[#0f271f]">Your restaurant</p>
            <div className="mt-1 flex items-center gap-1">
              <span className="text-[12px] font-semibold text-[#0f271f]">4.5</span>
              <span className="text-[12px] text-[#f5a623]">★★★★★</span>
            </div>
          </div>
        </div>
        <div className="mt-2.5 space-y-2.5 px-1 opacity-50">
          <div className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-lg bg-white/40" />
            <div className="h-2.5 flex-1 rounded-full bg-white/40" />
          </div>
          <div className="flex items-center gap-2.5">
            <div className="h-9 w-9 rounded-lg bg-white/30" />
            <div className="h-2.5 w-[70%] rounded-full bg-white/30" />
          </div>
        </div>
      </div>
    </div>
  );
}

export function OrderingRavioliVisual() {
  return (
    <div className="relative w-full max-w-[320px]">
      <div className="absolute inset-x-3 top-3 h-[96px] rounded-2xl bg-[#cfc5b8]/70" />
      <div className="absolute inset-x-1.5 top-1.5 h-[96px] rounded-2xl bg-[#ddd4c7]/85" />
      <div className="relative flex items-center gap-3.5 rounded-2xl bg-white p-4 shadow-[0_12px_32px_rgba(0,0,0,0.1)]">
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-2">
            <p className="text-[16px] font-bold text-[#0f271f]">Mushroom Ravioli</p>
            <span className="text-[#c45c5c]" aria-hidden="true">
              ♥
            </span>
          </div>
          <div className="mt-2.5 space-y-1.5">
            <div className="h-1.5 w-[90%] rounded-full bg-[#dce6dd]" />
            <div className="h-1.5 w-[60%] rounded-full bg-[#dce6dd]" />
          </div>
          <p className="mt-2.5 text-[14px] font-semibold text-[#0f271f]">$13.00</p>
        </div>
        <div className="relative h-[72px] w-[72px] shrink-0 overflow-hidden rounded-full">
          <Image src="/menu/pasta-alfredo.jpg" alt="" fill className="object-cover" sizes="72px" />
        </div>
      </div>
    </div>
  );
}

export function PhoneImprovingVisual() {
  return (
    <div className="absolute inset-0">
      <div
        className="absolute inset-x-0 bottom-0 h-[58%]"
        style={{
          background: "linear-gradient(160deg, #3d8f6e 0%, #2f6b54 40%, #174c3a 100%)",
        }}
        aria-hidden="true"
      >
        <svg
          className="absolute inset-0 h-full w-full opacity-25"
          viewBox="0 0 400 220"
          preserveAspectRatio="none"
        >
          {[80, 130, 180, 240].map((r) => (
            <circle
              key={r}
              cx="320"
              cy="220"
              r={r}
              fill="none"
              stroke="rgba(255,255,255,0.35)"
              strokeWidth="1"
            />
          ))}
        </svg>
      </div>

      <div className="absolute bottom-[-300px] left-1/2 w-[440px] -translate-x-1/2 rotate-[10deg] sm:bottom-[-340px] sm:w-[520px] md:bottom-[-380px] md:w-[580px]">
        <div className="relative aspect-[3/4] w-full drop-shadow-[0_24px_50px_rgba(0,0,0,0.32)]">
          <Image
            src="/mockup1.png"
            alt="Restaurant website on a phone"
            fill
            className="object-contain object-top"
            sizes="580px"
          />
        </div>
      </div>
    </div>
  );
}

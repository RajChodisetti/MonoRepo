import Image from "next/image";

export function CateringMenuStackVisual() {
  const items = [
    {
      name: "Lasagna - 12 Servings",
      price: "$89.99",
      image: "/menu/pasta-alfredo.jpg",
    },
    {
      name: "50 Breadsticks",
      price: "$45.99",
      image: "/menu/garlic-bread-v2.jpg",
    },
    {
      name: "Large Greek Salad",
      price: "$39.99",
      image: "/menu/caesar-salad.jpg",
    },
  ] as const;

  return (
    <div className="flex w-full max-w-[300px] flex-col gap-3 pt-1 sm:max-w-[320px]">
      {items.map((item) => (
        <div
          key={item.name}
          className="flex items-center gap-3.5 rounded-2xl bg-white p-3 shadow-[0_8px_24px_rgba(0,0,0,0.08)]"
        >
          <div className="relative h-[72px] w-[72px] shrink-0 overflow-hidden rounded-xl">
            <Image src={item.image} alt="" fill className="object-cover" sizes="72px" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[14px] font-bold leading-snug text-[#0f271f]">{item.name}</p>
            <div className="mt-2 space-y-1.5">
              <div className="h-1.5 w-[90%] rounded-full bg-[#dce6dd]" />
              <div className="h-1.5 w-[60%] rounded-full bg-[#dce6dd]" />
            </div>
            <p className="mt-2 text-[13px] font-semibold text-[#0f271f]">{item.price}</p>
          </div>
        </div>
      ))}
    </div>
  );
}

export function CateringSearchVisual() {
  return (
    <div className="w-full max-w-[320px]">
      <div className="flex items-center gap-2.5 rounded-2xl bg-white px-3.5 py-3.5 shadow-[0_8px_28px_rgba(0,0,0,0.12)]">
        <svg viewBox="0 0 20 20" className="h-4 w-4 shrink-0 text-[#666]" aria-hidden="true">
          <circle cx="9" cy="9" r="5.5" fill="none" stroke="currentColor" strokeWidth="1.6" />
          <path d="m13.5 13.5 3 3" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" />
        </svg>
        <span className="text-[14px] font-medium text-[#888]">Catering near me</span>
      </div>

      <div className="relative mt-3.5 rounded-[18px] bg-white/25 p-3 backdrop-blur-[1px]">
        <div className="flex items-center gap-3 rounded-2xl bg-white p-3.5 shadow-[0_10px_28px_rgba(0,0,0,0.14)]">
          <div className="relative h-12 w-12 shrink-0 overflow-hidden rounded-xl">
            <Image src="/menu/classic-burger.jpg" alt="" fill className="object-cover" sizes="48px" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-[15px] font-semibold text-[#0f271f]">Your restaurant</p>
            <div className="mt-1 flex items-center gap-1">
              <span className="text-[12px] font-semibold text-[#0f271f]">4.8</span>
              <span className="text-[12px] text-[#f5a623]">★★★★★</span>
            </div>
            <div className="mt-2 h-1.5 w-[80%] rounded-full bg-[#dce6dd]" />
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

const FOOD_TILES = [
  "/menu/pasta-alfredo.jpg",
  "/menu/caesar-salad.jpg",
  "/menu/classic-burger.jpg",
  "/menu/pepperoni-pizza.jpg",
] as const;

export function CateringFoodCollageVisual() {
  return (
    <div className="relative flex h-[260px] w-full items-center justify-center sm:h-[300px]">
      <svg
        className="pointer-events-none absolute inset-0 h-full w-full opacity-40"
        viewBox="0 0 400 260"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        {[70, 110, 150, 190].map((y) => (
          <path
            key={y}
            d={`M0 ${y} Q 100 ${y - 16} 200 ${y} T 400 ${y}`}
            fill="none"
            stroke="rgba(0,0,0,0.1)"
            strokeWidth="1"
          />
        ))}
      </svg>
      <div className="relative grid grid-cols-2 gap-3">
        {FOOD_TILES.map((src, i) => (
          <div
            key={src}
            className="relative h-24 w-24 overflow-hidden rounded-2xl shadow-md sm:h-28 sm:w-28"
            style={{ transform: `translateY(${(i % 2) * 12}px)` }}
          >
            <Image src={src} alt="" fill className="object-cover" sizes="112px" />
            <span className="absolute bottom-1.5 right-1.5 flex h-6 w-6 items-center justify-center rounded-full bg-white text-[14px] font-bold text-[#0f271f] shadow">
              +
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

/** Bottom-right card — uses `mockup-2.png` phone */
export function CateringPhoneMockupVisual() {
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

      <div className="absolute bottom-[-260px] left-1/2 w-[380px] -translate-x-1/2 rotate-[8deg] sm:bottom-[-300px] sm:w-[440px] md:bottom-[-340px] md:w-[500px]">
        <div className="relative aspect-[402/621] w-full drop-shadow-[0_24px_50px_rgba(0,0,0,0.32)]">
          <Image
            src="/mockup-2.png"
            alt="Catering ordering on a phone"
            fill
            className="object-contain object-top"
            sizes="500px"
            priority
            quality={92}
          />
        </div>
      </div>
    </div>
  );
}

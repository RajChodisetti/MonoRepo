import Image from "next/image";

export function UpsellFlowVisual() {
  return (
    <div className="relative flex w-full max-w-[320px] flex-col items-center gap-3 py-2">
      {/* Primary item */}
      <div className="flex w-full items-center gap-3 rounded-2xl bg-white p-3 shadow-[0_10px_28px_rgba(0,0,0,0.1)]">
        <div className="relative h-14 w-14 shrink-0 overflow-hidden rounded-xl">
          <Image src="/menu/birria-tacos.jpg" alt="" fill className="object-cover" sizes="56px" />
          <span className="absolute left-1 top-1 rounded-full bg-[#2f6b54] px-1.5 py-0.5 text-[8px] font-bold text-white">
            New
          </span>
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-[14px] font-bold text-[#0f271f]">Birria Tacos</p>
          <p className="text-[13px] text-[#666]">$13.99</p>
        </div>
        <span className="inline-flex items-center gap-1 rounded-full bg-[#2f6b54] px-2.5 py-1.5 text-[11px] font-bold text-white">
          ✓ Added
        </span>
      </div>

      {/* Connector */}
      <div className="flex flex-col items-center">
        <div className="h-4 w-[2px] bg-[#2f6b54]" />
        <span className="rounded-full bg-[#2f6b54] px-3 py-1 text-[11px] font-bold text-white">
          ✓ Most likely upsell
        </span>
        <div className="h-3 w-[2px] bg-[#2f6b54]" />
      </div>

      {/* Chosen upsell */}
      <div className="flex flex-col items-center">
        <div className="relative h-20 w-20 overflow-hidden rounded-2xl shadow-md">
          <Image src="/menu/tacos.jpg" alt="" fill className="object-cover" sizes="80px" />
        </div>
        <p className="mt-1.5 text-[13px] font-semibold text-[#0f271f]">Quesadilla</p>
      </div>

      {/* Rejected options */}
      <div className="mt-1 flex w-full justify-between gap-4 px-2">
        {[
          { name: "Churros", src: "/menu/garlic-bread.jpg" },
          { name: "Lemonade", src: "/menu/caesar-salad.jpg" },
        ].map((item) => (
          <div key={item.name} className="relative flex flex-col items-center opacity-70">
            <div className="relative h-14 w-14 overflow-hidden rounded-xl">
              <Image src={item.src} alt="" fill className="object-cover" sizes="56px" />
            </div>
            <span className="absolute -right-1 -top-1 flex h-5 w-5 items-center justify-center rounded-full bg-[#e85a5a] text-[11px] font-bold text-white">
              ×
            </span>
            <p className="mt-1 text-[11px] font-medium text-[#555]">{item.name}</p>
          </div>
        ))}
      </div>
    </div>
  );
}

export function UpsellCheckoutVisual() {
  return (
    <div className="w-full max-w-[280px] rounded-[22px] bg-white p-4 shadow-[0_16px_40px_rgba(0,0,0,0.14)]">
      <div className="flex items-center gap-3">
        <div className="relative h-12 w-12 shrink-0 overflow-hidden rounded-xl">
          <Image src="/menu/birria-tacos.jpg" alt="" fill className="object-cover" sizes="48px" />
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-[14px] font-bold text-[#0f271f]">Birria Tacos</p>
          <p className="text-[12px] text-[#666]">$13.99</p>
        </div>
        <span className="rounded-full bg-[#2f6b54] px-2.5 py-1 text-[10px] font-bold text-white">
          ✓ Added
        </span>
      </div>

      <p className="mt-4 text-[12px] font-medium text-[#888]">Goes well with</p>

      <div className="mt-2 space-y-2.5">
        {[
          { name: "Blue Lemonade", price: "$4.99", src: "/menu/caesar-salad.jpg" },
          { name: "Cinnamon Churros", price: "$7.99", src: "/menu/garlic-bread.jpg" },
        ].map((item) => (
          <div
            key={item.name}
            className="flex items-center gap-3 rounded-xl bg-[#f2ecdf] px-2.5 py-2"
          >
            <div className="relative h-11 w-11 shrink-0 overflow-hidden rounded-lg">
              <Image src={item.src} alt="" fill className="object-cover" sizes="44px" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-[13px] font-semibold text-[#0f271f]">{item.name}</p>
              <p className="text-[12px] text-[#666]">{item.price}</p>
            </div>
            <span className="flex h-7 w-7 items-center justify-center rounded-full bg-white text-[16px] font-bold text-[#0f271f] shadow-sm">
              +
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

const UPSELL_AVATARS = [
  "/people/james.jpg",
  "/people/maria.jpg",
  "/people/priya.jpg",
  "/people/david.jpg",
  "/people/lena.jpg",
  "/people/kevin.jpg",
] as const;

export function UpsellDataAvatarsVisual() {
  return (
    <div className="relative flex h-[240px] w-full items-center justify-center sm:h-[280px]">
      <svg
        className="pointer-events-none absolute inset-0 h-full w-full"
        viewBox="0 0 360 240"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        {[90, 130, 170].map((y) => (
          <path
            key={y}
            d={`M30 ${y} Q 120 ${y - 20} 180 ${y} T 330 ${y}`}
            fill="none"
            stroke="rgba(0,0,0,0.1)"
            strokeWidth="1.2"
            strokeDasharray="4 6"
          />
        ))}
      </svg>
      <div className="relative grid grid-cols-3 gap-x-4 gap-y-5">
        {UPSELL_AVATARS.map((src, i) => (
          <div
            key={src}
            className={`relative h-14 w-14 overflow-hidden rounded-full border-[3px] shadow-md sm:h-16 sm:w-16 ${
              i % 2 === 0 ? "border-[#2f6b54]" : "border-white"
            }`}
          >
            <Image src={src} alt="" fill className="object-cover" sizes="64px" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function UpsellImprovingPhotoVisual() {
  return (
    <div className="absolute inset-0">
      <Image
        src="/guides/interview.jpg"
        alt=""
        fill
        className="object-cover object-[center_25%]"
        sizes="(max-width: 1024px) 90vw, 560px"
        quality={90}
      />
      <div
        className="absolute inset-0"
        style={{
          background:
            "linear-gradient(180deg, rgba(10,10,10,0.55) 0%, rgba(10,10,10,0.2) 45%, rgba(10,10,10,0.4) 100%)",
        }}
        aria-hidden="true"
      />
    </div>
  );
}

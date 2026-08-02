import Image from "next/image";

/** Feature-split + green card phone — uses `map.png` */
export function DeliveryMapPhoneVisual() {
  return (
    <div className="relative mx-auto w-full max-w-[240px] sm:max-w-[270px]">
      <div className="relative aspect-[1024/1536] w-full drop-shadow-[0_22px_44px_rgba(0,0,0,0.2)]">
        <Image
          src="/map.png"
          alt="Delivery tracking on a phone"
          fill
          className="object-contain object-top"
          sizes="270px"
          priority
          quality={92}
        />
      </div>
    </div>
  );
}

/** Compact tracking card for green feature card (also map.png) */
export function DeliveryTrackingCardVisual() {
  return (
    <div className="relative mx-auto w-full max-w-[260px] overflow-hidden rounded-[22px] bg-white shadow-[0_16px_40px_rgba(0,0,0,0.16)]">
      <div className="relative h-[160px] w-full sm:h-[180px]">
        <Image
          src="/map.png"
          alt=""
          fill
          className="object-cover object-top"
          sizes="260px"
          quality={90}
        />
      </div>
      <div className="px-4 py-3.5">
        <div className="flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-[#2f6b54]" />
          <p className="text-[13px] font-semibold text-[#0f271f]">
            Arrives between 2:00 – 2:04 PM
          </p>
        </div>
        <p className="mt-2 text-[12px] leading-snug text-[#666]">
          1x Pesto salad, 1x Coca Cola, 2x Spinach Soup, 2x Calamari
        </p>
        <div className="mt-3 rounded-full border border-[#ddd] py-2 text-center text-[13px] font-semibold text-[#333]">
          View order
        </div>
      </div>
    </div>
  );
}

export function DeliveryControlMapVisual() {
  return (
    <div className="absolute inset-0">
      <div
        className="absolute inset-0 opacity-40"
        style={{
          backgroundImage:
            "linear-gradient(rgba(255,255,255,0.65), rgba(255,255,255,0.65)), url(/app/wood-bg.jpg)",
          backgroundSize: "cover",
          backgroundPosition: "center",
        }}
        aria-hidden="true"
      />
      {/* Simple street grid overlay */}
      <svg
        className="absolute inset-0 h-full w-full opacity-30"
        viewBox="0 0 400 400"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        {Array.from({ length: 10 }).map((_, i) => (
          <line
            key={`h-${i}`}
            x1="0"
            y1={i * 40}
            x2="400"
            y2={i * 40}
            stroke="#999"
            strokeWidth="1"
          />
        ))}
        {Array.from({ length: 8 }).map((_, i) => (
          <line
            key={`v-${i}`}
            x1={i * 50}
            y1="0"
            x2={i * 50}
            y2="400"
            stroke="#999"
            strokeWidth="1"
          />
        ))}
      </svg>
      <div className="absolute inset-0 flex items-center justify-center pt-16">
        <span className="flex h-16 w-16 items-center justify-center rounded-full bg-[#2f6b54] shadow-lg">
          <svg viewBox="0 0 24 24" className="h-8 w-8 text-white" aria-hidden="true">
            <path
              d="M4 10.5V19h16v-8.5M4 10.5 12 4l8 6.5M9 19v-5h6v5"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.7"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </span>
      </div>
    </div>
  );
}

export function DeliveryGuestPhotoVisual() {
  return (
    <div className="absolute inset-0">
      <Image
        src="/guides/interview.jpg"
        alt=""
        fill
        className="object-cover object-[center_30%]"
        sizes="(max-width: 1024px) 90vw, 560px"
        quality={90}
      />
      <div
        className="absolute inset-0"
        style={{
          background:
            "linear-gradient(180deg, rgba(10,10,10,0.55) 0%, rgba(10,10,10,0.25) 50%, rgba(10,10,10,0.4) 100%)",
        }}
        aria-hidden="true"
      />
      {/* Floating mini ordering UI */}
      <div className="absolute bottom-8 left-1/2 w-[200px] -translate-x-1/2 overflow-hidden rounded-2xl bg-white shadow-[0_16px_36px_rgba(0,0,0,0.3)] sm:w-[220px]">
        <div className="px-3 pb-3 pt-2.5">
          <p className="text-[12px] font-bold text-[#0f271f]">Manhattan Bistro</p>
          <div className="mt-2 rounded-full bg-[#efefef] px-3 py-1.5 text-center text-[11px] font-semibold text-[#444]">
            Pickup at 11:35am ▾
          </div>
          <div className="mt-2 flex gap-1.5 overflow-hidden">
            {["Offers", "Pasta", "Pizza"].map((label) => (
              <span
                key={label}
                className="shrink-0 rounded-full bg-[#f2ecdf] px-2 py-1 text-[9px] font-semibold text-[#555]"
              >
                {label}
              </span>
            ))}
          </div>
          <div className="relative mt-2 h-16 w-full overflow-hidden rounded-xl">
            <Image src="/menu/classic-burger.jpg" alt="" fill className="object-cover" sizes="220px" />
          </div>
        </div>
      </div>
    </div>
  );
}

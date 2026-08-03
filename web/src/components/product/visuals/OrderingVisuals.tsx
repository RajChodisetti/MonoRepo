import Image from "next/image";

/** Feature-split left — Las Vegas ordering UI (`mockup-2.png`) */
export function OrderingPhonePreviewVisual() {
  return (
    <div className="relative mx-auto w-full max-w-[260px] sm:max-w-[280px]">
      <div className="relative aspect-[402/621] w-full drop-shadow-[0_20px_40px_rgba(0,0,0,0.18)]">
        <Image
          src="/mockup-2.png"
          alt="Online ordering app preview"
          fill
          className="object-contain object-top"
          sizes="280px"
          priority
          quality={92}
        />
      </div>
    </div>
  );
}

/** Full feature card — Pioneer Pizza dark UI (`mockup3.png`) */
export function OrderingAppShowcaseVisual() {
  return (
    <div className="relative mx-auto w-full max-w-[240px] sm:max-w-[280px]">
      <div className="relative aspect-[1024/1536] w-full drop-shadow-[0_24px_48px_rgba(0,0,0,0.28)]">
        <Image
          src="/mockup3.png"
          alt="Restaurant online ordering on a phone"
          fill
          className="object-contain object-top"
          sizes="280px"
          priority
          quality={92}
        />
      </div>
    </div>
  );
}

const ORDERING_AVATARS = [
  "/people/james.jpg",
  "/people/maria.jpg",
  "/people/priya.jpg",
  "/people/david.jpg",
] as const;

export function OrderingCustomerListVisual() {
  return (
    <div className="relative flex h-[240px] w-full items-end justify-center pb-4 sm:h-[280px]">
      <svg
        className="pointer-events-none absolute inset-0 h-full w-full"
        viewBox="0 0 360 240"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        {[80, 120, 160, 200].map((y) => (
          <path
            key={y}
            d={`M20 ${y} Q 110 ${y + 24} 180 ${y} T 340 ${y}`}
            fill="none"
            stroke="rgba(0,0,0,0.1)"
            strokeWidth="1.2"
            strokeDasharray="4 6"
          />
        ))}
      </svg>
      <div className="relative flex -space-x-2">
        {ORDERING_AVATARS.map((src, i) => (
          <div
            key={src}
            className="relative h-14 w-14 overflow-hidden rounded-full border-[3px] border-white shadow-md sm:h-16 sm:w-16"
            style={{
              transform: `translateY(${i * 10}px)`,
              zIndex: ORDERING_AVATARS.length - i,
            }}
          >
            <Image src={src} alt="" fill className="object-cover" sizes="64px" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function FeeSavingsToastVisual() {
  return (
    <div className="flex w-full items-center justify-center px-4 pb-6 pt-8">
      <div className="flex max-w-[300px] items-start gap-3 rounded-2xl bg-white px-4 py-3.5 shadow-[0_14px_36px_rgba(0,0,0,0.14)]">
        <span className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#0f271f] text-white">
          <svg viewBox="0 0 16 16" className="h-3.5 w-3.5" aria-hidden="true">
            <path
              d="M8 2.2 9.6 6h3.8l-3 2.3 1.1 3.7L8 9.9l-3.5 2.1 1.1-3.7-3-2.3H6.4L8 2.2Z"
              fill="currentColor"
            />
          </svg>
        </span>
        <p className="text-[13px] font-semibold leading-snug text-[#0f271f]">
          You saved $4.51 by ordering directly from us instead of delivery apps.
        </p>
      </div>
    </div>
  );
}

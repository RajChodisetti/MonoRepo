import Image from "next/image";

export function ReviewsOwnerPhotoVisual() {
  return (
    <div className="absolute inset-0">
      <Image
        src="/product/reviews-owner.jpg"
        alt="Restaurant owner in the kitchen"
        fill
        className="object-cover object-[center_18%]"
        sizes="(max-width: 1024px) 90vw, 480px"
        quality={90}
      />
    </div>
  );
}

export function GoogleReviewsStackVisual() {
  const reviews = [
    {
      name: "Jenny G",
      text: "Great ordering experience and delivery!",
    },
    {
      name: "Tom B",
      text: "Best wings I've had delivered, still hot!",
    },
    {
      name: "Alex J",
      text: "So easy to order, will definitely be back.",
    },
  ] as const;

  return (
    <div className="flex w-full max-w-[320px] flex-col gap-3">
      {reviews.map((review) => (
        <div
          key={review.name}
          className="rounded-2xl bg-white px-4 py-3.5 shadow-[0_10px_28px_rgba(0,0,0,0.12)]"
        >
          <div className="flex items-center justify-between gap-2">
            <p className="text-[14px] font-bold text-[#0f271f]">{review.name}</p>
            <span className="text-[12px] tracking-tight text-[#f5a623]">★★★★★</span>
          </div>
          <p className="mt-1.5 text-[13px] leading-snug text-[#555]">{review.text}</p>
        </div>
      ))}
    </div>
  );
}

const FLOW_AVATARS = [
  "/people/james.jpg",
  "/people/maria.jpg",
  "/people/priya.jpg",
  "/people/david.jpg",
  "/people/lena.jpg",
  "/people/kevin.jpg",
  "/people/sofia.jpg",
  "/people/anil.jpg",
] as const;

export function CustomersFlowVisual() {
  return (
    <div className="relative flex h-[260px] w-full items-center justify-center sm:h-[300px]">
      <svg
        className="pointer-events-none absolute inset-0 h-full w-full"
        viewBox="0 0 400 260"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        {[70, 110, 150, 190].map((y) => (
          <path
            key={y}
            d={`M20 ${y} Q 120 ${y - 28} 200 ${y} T 380 ${y}`}
            fill="none"
            stroke="rgba(0,0,0,0.12)"
            strokeWidth="1.25"
            strokeDasharray="5 6"
          />
        ))}
      </svg>
      <div className="relative grid w-full max-w-[320px] grid-cols-4 gap-x-3 gap-y-5 px-2">
        {FLOW_AVATARS.map((src, i) => (
          <div
            key={src}
            className="relative mx-auto h-12 w-12 overflow-hidden rounded-full border-[3px] border-white shadow-md sm:h-14 sm:w-14"
            style={{
              transform: `translateY(${(i % 3) * 8 - 8}px)`,
            }}
          >
            <Image src={src} alt="" fill className="object-cover" sizes="56px" />
          </div>
        ))}
      </div>
    </div>
  );
}

/** Better ratings card — uses mockup1.png phone (same as website page) */
export function ReviewsPhoneMockupVisual() {
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
            priority
          />
        </div>
      </div>
    </div>
  );
}

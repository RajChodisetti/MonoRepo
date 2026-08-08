import Image from "next/image";
import Link from "next/link";

type GrowOnlineCtaProps = {
  /** Landing: split + phone. Product pages: centered banner. */
  variant?: "split" | "centered";
};

export default function GrowOnlineCta({ variant = "split" }: GrowOnlineCtaProps) {
  if (variant === "centered") {
    return (
      <section className="bg-bg pt-6 sm:pt-8">
        <div className="tuvi-forest-panel relative w-full overflow-hidden">
          <svg
            className="pointer-events-none absolute inset-0 h-full w-full opacity-90"
            viewBox="0 0 1440 520"
            preserveAspectRatio="xMidYMid slice"
            aria-hidden="true"
          >
            {[160, 260, 380, 520, 680].map((r, i) => (
              <rect
                key={r}
                x={720 - r / 2}
                y={280 - r / 2}
                width={r}
                height={r}
                rx={r * 0.28}
                fill="none"
                stroke="rgba(255,255,255,0.12)"
                strokeWidth={1.1 - i * 0.06}
                transform={`rotate(${i * 8} 720 280)`}
              />
            ))}
          </svg>

          <div className="relative z-10 mx-auto flex max-w-[820px] flex-col items-center px-5 pb-24 pt-16 text-center sm:px-8 sm:pb-28 sm:pt-20 md:pb-32 md:pt-24">
            <h2 className="max-w-[16ch] font-display text-[clamp(1.85rem,4.2vw,3.15rem)] font-semibold leading-[1.12] tracking-[-0.03em] text-bg">
              The easiest way to grow your restaurant online
            </h2>

            <Link
              href="/book"
              className="mt-7 inline-flex items-center justify-center rounded-full bg-bg px-7 py-3 text-[14px] font-semibold text-ink transition-colors hover:bg-sage sm:mt-8 sm:px-8 sm:py-3.5 sm:text-[15px]"
            >
              Get a free demo
            </Link>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className="bg-bg pt-8 sm:pt-10">
      <div className="tuvi-forest-panel relative w-full overflow-hidden rounded-t-[28px] sm:rounded-t-[36px] md:rounded-t-[44px]">
        <svg
          className="pointer-events-none absolute inset-0 h-full w-full"
          viewBox="0 0 1440 480"
          preserveAspectRatio="none"
          aria-hidden="true"
        >
          {[180, 280, 390, 520, 680].map((r, i) => (
            <circle
              key={r}
              cx="180"
              cy="500"
              r={r}
              fill="none"
              stroke="rgba(255,255,255,0.1)"
              strokeWidth={1.15 - i * 0.05}
            />
          ))}
        </svg>

        <div className="relative z-10 mx-auto grid max-w-[1100px] items-center gap-4 px-4 pt-8 sm:gap-6 sm:px-8 sm:pt-9 md:px-12 lg:grid-cols-2 lg:gap-4 lg:pt-6 lg:pb-0">
          <div className="pb-6 lg:pb-10">
            <h2 className="max-w-[14ch] font-display text-[clamp(1.7rem,3.4vw,2.65rem)] font-semibold leading-[1.12] tracking-[-0.03em] text-bg">
              The easiest way to grow your restaurant online.
            </h2>

            <div className="mt-5 flex flex-wrap items-center gap-2.5 sm:mt-6 sm:gap-3">
              <Link
                href="/book"
                className="inline-flex items-center justify-center rounded-full bg-bg px-5 py-2.5 text-[13px] font-semibold text-ink transition-colors hover:bg-sage sm:px-6 sm:text-[14px]"
              >
                Get a free demo
              </Link>
              <Link
                href="/how-it-works"
                className="inline-flex items-center justify-center rounded-full bg-white/10 px-5 py-2.5 text-[13px] font-semibold text-bg ring-1 ring-white/20 backdrop-blur-sm transition-colors hover:bg-white/20 sm:px-6 sm:text-[14px]"
              >
                See how it works
              </Link>
            </div>
          </div>

          <div className="relative mx-auto flex w-full justify-center lg:justify-end lg:self-end">
            <div className="relative -mb-2 h-[420px] w-[300px] sm:h-[500px] sm:w-[360px] md:h-[540px] md:w-[390px] lg:h-[560px] lg:w-[420px]">
              <Image
                src="/image.png"
                alt="Tuvi restaurant mobile app on a phone"
                fill
                priority
                className="object-contain object-bottom drop-shadow-[0_20px_40px_rgba(0,0,0,0.25)]"
                sizes="(max-width: 640px) 300px, (max-width: 1024px) 390px, 420px"
              />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

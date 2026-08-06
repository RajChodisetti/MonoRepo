export default function BrandTech() {
  return (
    <section className="bg-bg px-4 pb-16 pt-14 sm:px-8 sm:pb-20 sm:pt-16 md:px-12 md:pt-20">
      <div className="mx-auto grid max-w-[1100px] items-center gap-8 lg:grid-cols-[minmax(0,0.3fr)_minmax(0,0.7fr)] lg:gap-10">
        <div>
          <h2 className="max-w-[14ch] font-display text-[clamp(1.85rem,3.4vw,2.85rem)] font-semibold leading-[1.1] tracking-[-0.03em] text-ink">
            Scan. Order. Enjoy — under your brand
          </h2>
          <p className="mt-4 max-w-[34ch] text-[15px] leading-relaxed text-muted sm:text-[16px]">
            Guests already order on their phone. That&apos;s why Tuvi gives your restaurant its
            own mobile app — so every order stays yours.
          </p>
        </div>

        <div className="relative overflow-hidden rounded-[28px] sm:rounded-[36px] md:rounded-[40px]">
          <div className="relative aspect-[16/10] w-full min-h-[260px] sm:min-h-[360px] lg:min-h-[420px]">
            <video
              className="absolute inset-0 h-full w-full object-cover object-center"
              autoPlay
              muted
              loop
              playsInline
              preload="metadata"
              aria-label="Guests ordering at a restaurant with Tuvi"
            >
              <source
                src="/hf_20260727_055931_a989648e-ba15-4a67-919d-e2e758e351fe.mp4"
                type="video/mp4"
              />
            </video>
          </div>
        </div>
      </div>
    </section>
  );
}

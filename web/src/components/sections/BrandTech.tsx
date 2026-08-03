export default function BrandTech() {
  return (
    <section className="bg-bg px-4 pb-16 pt-14 sm:px-8 sm:pb-20 sm:pt-16 md:px-12 md:pt-20">
      <div className="mx-auto max-w-[1100px]">
        <h2 className="max-w-[16ch] font-display text-[clamp(2.15rem,4.5vw,3.75rem)] font-semibold leading-[1.08] tracking-[-0.03em] text-ink">
          Give your restaurant the same tech as the big brands
        </h2>

        <div className="relative mt-8 overflow-hidden rounded-[28px] sm:mt-10 sm:rounded-[36px] md:rounded-[40px]">
          <div className="relative aspect-[16/10] w-full min-h-[320px] sm:min-h-[420px] md:min-h-[520px]">
            <video
              className="absolute inset-0 h-full w-full object-cover object-center"
              autoPlay
              muted
              loop
              playsInline
              preload="metadata"
              aria-label="Restaurant branded mobile app on a phone"
            >
              <source
                src="/hf_20260727_055931_a989648e-ba15-4a67-919d-e2e758e351fe.mp4"
                type="video/mp4"
              />
            </video>

            {/* Soft fade so caption stays readable */}
            <div
              className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/55 via-black/15 to-transparent"
              aria-hidden="true"
            />

            <p className="absolute bottom-6 left-6 z-10 max-w-[28ch] text-[clamp(1.05rem,2vw,1.45rem)] font-semibold leading-snug tracking-[-0.02em] text-white sm:bottom-8 sm:left-8 md:bottom-10 md:left-10">
              Your customers are used to ordering on their phone. That&apos;s why we
              give your restaurant its own mobile app.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}

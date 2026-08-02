import type { ProductFeatureSplitConfig } from "@/content/products/types";
import { ProductIcon } from "@/components/product/ProductIcons";
import { ProductVisual } from "@/components/product/visuals/registry";

type ProductFeatureSplitProps = {
  config: ProductFeatureSplitConfig;
};

const PANEL_BG = {
  peach: "#f2ecdf",
  green: "#2f6b54",
} as const;

export default function ProductFeatureSplit({ config }: ProductFeatureSplitProps) {
  const panel = config.visualPanel ?? "peach";
  const headingTone = config.headingTone ?? "muted";
  const isFlushPhoto = panel === "none";
  const isVideoPanel = config.visual === "ai-phone-mockup";

  return (
    <section className="bg-bg px-4 pb-16 pt-5 sm:px-8 sm:pb-20 sm:pt-7 md:px-12">
      <div className="mx-auto max-w-[1040px]">
        <h2
          className={`mx-auto max-w-[22ch] text-center font-display text-[clamp(1.45rem,3vw,2.25rem)] font-semibold leading-[1.18] tracking-[-0.03em] ${
            headingTone === "dark" ? "text-ink" : "text-muted"
          }`}
        >
          {config.heading.split("\n").map((line, i) => (
            <span key={line} className={i > 0 ? "mt-1 block" : "block"}>
              {line}
            </span>
          ))}
        </h2>

        <div className="mt-10 grid items-stretch gap-8 lg:mt-12 lg:grid-cols-2 lg:gap-12">
          {isFlushPhoto || isVideoPanel ? (
            <div className="relative h-[440px] overflow-hidden rounded-[28px] sm:h-[480px] sm:rounded-[32px]">
              <ProductVisual id={config.visual} />
            </div>
          ) : (
            <div
              className="relative flex h-[440px] justify-center overflow-hidden rounded-[28px] px-6 pt-5 sm:h-[480px] sm:rounded-[32px] sm:px-10 sm:pt-6"
              style={{ backgroundColor: PANEL_BG[panel] }}
            >
              <div className="flex w-full max-w-[340px] items-start justify-center">
                <ProductVisual id={config.visual} />
              </div>
            </div>
          )}

          <div className="flex flex-col gap-8 sm:gap-9">
            {config.features.map((feature) => (
              <div key={feature.title} className="flex gap-4">
                <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-sage text-ink sm:h-12 sm:w-12">
                  <ProductIcon id={feature.icon} />
                </span>
                <div className="min-w-0 pt-0.5">
                  <h3 className="text-[16px] font-bold tracking-[-0.02em] text-ink sm:text-[17px]">
                    {feature.title}
                  </h3>
                  <p className="mt-1.5 text-[14px] leading-relaxed text-muted sm:text-[15px]">
                    {feature.body}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

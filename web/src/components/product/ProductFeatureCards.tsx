import type {
  FeatureCardTheme,
  ProductFeatureCard,
  ProductFeatureCardsConfig,
} from "@/content/products/types";
import { ProductVisual } from "@/components/product/visuals/registry";

const THEME_STYLES: Record<
  FeatureCardTheme,
  { background: string; labelClass: string; titleClass: string }
> = {
  blue: {
    background:
      "radial-gradient(36rem 24rem at 8% 0%, rgba(102, 138, 116, 0.4), transparent 58%), #0f271f",
    labelClass: "text-bg/70",
    titleClass: "text-bg",
  },
  beige: {
    background: "#f2ecdf",
    labelClass: "text-muted",
    titleClass: "text-ink",
  },
  "white-green": {
    background: "#fffef8",
    labelClass: "text-muted",
    titleClass: "text-ink",
  },
  green: {
    background:
      "radial-gradient(ellipse 90% 80% at 30% 40%, #3d8f6e 0%, #2f6b54 35%, #174c3a 70%, #0f271f 100%)",
    labelClass: "text-bg/75",
    titleClass: "text-bg",
  },
  sky: {
    background: "linear-gradient(160deg, #7b927f 0%, #2f6b54 45%, #174c3a 100%)",
    labelClass: "text-bg/80",
    titleClass: "text-bg",
  },
  light: {
    background: "#f5f0e6",
    labelClass: "text-muted",
    titleClass: "text-ink",
  },
  "cream-blue": {
    background: "linear-gradient(180deg, #f2ecdf 0%, #f5f0e6 48%, #dce6dd 78%, #7b927f 100%)",
    labelClass: "text-muted",
    titleClass: "text-ink",
  },
  dark: {
    background: "#0f271f",
    labelClass: "text-bg/70",
    titleClass: "text-bg",
  },
  white: {
    background: "#fffef8",
    labelClass: "text-muted",
    titleClass: "text-ink",
  },
  indigo: {
    background:
      "radial-gradient(30rem 20rem at 96% 100%, rgba(220, 230, 221, 0.16), transparent 58%), #0f271f",
    labelClass: "text-bg/75",
    titleClass: "text-bg",
  },
};

function FeatureCard({ card }: { card: ProductFeatureCard }) {
  const theme = THEME_STYLES[card.theme];
  const isFull = card.layout === "full";
  const isWhiteGreen = card.theme === "white-green";
  const isDark = card.theme === "dark";
  const isPhoneMockup =
    card.visual === "reviews-phone-mockup" ||
    card.visual === "phone-improving" ||
    card.visual === "catering-phone-mockup";
  const isVideoCard = card.visual === "google-update";
  const isFillBackground =
    isDark ||
    isVideoCard ||
    card.visual === "delivery-control-map" ||
    card.visual === "delivery-guest-photo" ||
    card.visual === "listings-experts-photo" ||
    card.visual === "upsell-improving-photo" ||
    card.visual === "ai-phone-loyalty-photo" ||
    card.visual === "app-photo-fill" ||
    card.visual === "owner-photo-fill" ||
    card.visual === "kitchen-photo-fill";
  const visualOnRight = card.visualSide === "right";
  const isIvoryCard = isWhiteGreen || card.theme === "white";

  if (isFull) {
    return (
      <div
        className="grid min-h-[420px] overflow-hidden rounded-[28px] sm:min-h-[480px] sm:rounded-[32px] lg:grid-cols-2"
        style={{ background: theme.background }}
      >
        <div
          className={`flex items-center justify-center px-6 py-12 sm:px-10 sm:py-14 ${
            visualOnRight ? "order-2" : "order-1"
          }`}
        >
          <ProductVisual id={card.visual} />
        </div>
        <div
          className={`flex flex-col justify-center px-6 pb-12 pt-2 sm:px-10 sm:pb-14 lg:pt-14 ${
            visualOnRight ? "order-1" : "order-2"
          }`}
        >
          <p className={`text-[14px] font-medium ${theme.labelClass}`}>{card.label}</p>
          <h3
            className={`mt-2.5 max-w-[16ch] font-display text-[clamp(1.6rem,2.8vw,2.35rem)] font-semibold leading-[1.15] tracking-[-0.03em] ${theme.titleClass}`}
          >
            {card.title}
          </h3>
        </div>
      </div>
    );
  }

  return (
    <div
      className={`relative flex min-h-[520px] flex-col overflow-hidden rounded-[28px] sm:min-h-[580px] sm:rounded-[32px] ${
        isIvoryCard ? "bg-bg" : ""
      }`}
      style={isIvoryCard || isVideoCard ? undefined : { background: theme.background }}
    >
      {/* Video cards bake in label/title — skip HTML copy to avoid doubling */}
      {!isVideoCard ? (
        <div className="relative z-10 px-7 pt-8 sm:px-9 sm:pt-9">
          <p className={`text-[15px] font-medium ${theme.labelClass}`}>{card.label}</p>
          <h3
            className={`mt-2.5 max-w-[18ch] font-display text-[clamp(1.55rem,2.8vw,2.15rem)] font-semibold leading-[1.15] tracking-[-0.03em] ${theme.titleClass}`}
          >
            {card.title}
          </h3>
        </div>
      ) : (
        <h3 className="sr-only">
          {card.label}. {card.title}
        </h3>
      )}
      {isFillBackground ? (
        <div className="absolute inset-0 z-0">
          <ProductVisual id={card.visual} />
        </div>
      ) : (
        <div
          className={`relative mt-auto flex flex-1 overflow-hidden ${
            isWhiteGreen || isPhoneMockup
              ? "min-h-[320px]"
              : "items-end justify-center px-7 pb-9 sm:px-9 sm:pb-11"
          }`}
        >
          <ProductVisual id={card.visual} />
        </div>
      )}
    </div>
  );
}

type ProductFeatureCardsProps = {
  config: ProductFeatureCardsConfig;
};

export default function ProductFeatureCards({ config }: ProductFeatureCardsProps) {
  const fullCards = config.cards.filter((c) => c.layout === "full");
  const halfCards = config.cards.filter((c) => c.layout === "half");
  const sectionBg = config.sectionBg ?? "white";

  return (
    <section
      className={`px-4 py-14 sm:px-8 sm:py-16 md:px-12 md:py-18 ${
        sectionBg === "beige" ? "bg-parchment" : "bg-bg"
      }`}
    >
      <div className="mx-auto max-w-[1180px]">
        {config.sectionHeading ? (
          <h2 className="mx-auto max-w-[18ch] text-center font-display text-[clamp(1.55rem,3.2vw,2.4rem)] font-semibold leading-[1.15] tracking-[-0.03em] text-ink">
            {config.sectionHeading}
          </h2>
        ) : null}

        <div
          className={`flex flex-col gap-4 sm:gap-5 ${
            config.sectionHeading ? "mt-9 sm:mt-11" : ""
          }`}
        >
          {fullCards.map((card) => (
            <FeatureCard key={card.title} card={card} />
          ))}
          {halfCards.length > 0 ? (
            <div className="grid gap-4 sm:gap-5 lg:grid-cols-2">
              {halfCards.map((card) => (
                <FeatureCard key={card.title} card={card} />
              ))}
            </div>
          ) : null}
        </div>
      </div>
    </section>
  );
}

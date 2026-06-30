import type { RestaurantContent } from "@/data/types/restaurant";
import { mapAuroraContent } from "./lib/mapContent";
import { buildAuroraJsonLd } from "./seo";
import AuroraBackground from "./components/AuroraBackground";
import CursorGlow from "./components/ui/CursorGlow";
import ScrollProgress from "./components/ui/ScrollProgress";
import AuroraNav from "./components/AuroraNav";
import AuroraHero from "./components/AuroraHero";
import FeaturesInteractive from "./components/FeaturesInteractive";
import ProductShowcase from "./components/ProductShowcase";
import WorkflowTimeline from "./components/WorkflowTimeline";
import StatsSection from "./components/StatsSection";
import FeatureCards from "./components/FeatureCards";
import ShowcaseGrid from "./components/ShowcaseGrid";
import AuroraMenu from "./components/AuroraMenu";
import Testimonials from "./components/Testimonials";
import PricingCards from "./components/PricingCards";
import FAQ from "./components/FAQ";
import AuroraLocation from "./components/AuroraLocation";
import CTABanner from "./components/CTABanner";
import AuroraFooter from "./components/AuroraFooter";
import AuroraMobileBar from "./components/AuroraMobileBar";

export default function AuroraTemplate({
  restaurant,
}: {
  restaurant: RestaurantContent;
}) {
  const content = mapAuroraContent(restaurant);
  const jsonLd = buildAuroraJsonLd(restaurant);

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <AuroraBackground />
      <CursorGlow />
      <ScrollProgress />
      <AuroraNav restaurant={restaurant} />
      <main className="relative">
        <AuroraHero content={content.hero} />
        <FeaturesInteractive features={content.features} />
        <StatsSection stats={content.stats} />
        <ProductShowcase dishes={restaurant.signatureDishes} />
        <WorkflowTimeline steps={restaurant.storySteps} />
        <FeatureCards restaurant={restaurant} />
        <AuroraMenu restaurant={restaurant} />
        <ShowcaseGrid images={restaurant.galleryImages} />
        <Testimonials reviews={restaurant.reviews} rating={restaurant.rating} />
        <PricingCards tiers={content.pricingTiers} ctaHref={restaurant.secondaryCTA.href} />
        <FAQ items={content.faq} />
        <AuroraLocation restaurant={restaurant} />
        <CTABanner restaurant={restaurant} />
      </main>
      <AuroraFooter restaurant={restaurant} />
      <AuroraMobileBar restaurant={restaurant} />
    </>
  );
}

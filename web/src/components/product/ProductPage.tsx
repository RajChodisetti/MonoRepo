import ProductFaq from "@/components/product/ProductFaq";
import ProductFeatureCards from "@/components/product/ProductFeatureCards";
import ProductFeatureSplit from "@/components/product/ProductFeatureSplit";
import ProductHero from "@/components/product/ProductHero";
import SiteFooter from "@/components/layout/SiteFooter";
import GrowOnlineCta from "@/components/sections/GrowOnlineCta";
import type { ProductPageConfig } from "@/content/products/types";

type ProductPageProps = {
  config: ProductPageConfig;
};

export default function ProductPage({ config }: ProductPageProps) {
  return (
    <div className="bg-bg">
      <ProductHero hero={config.hero} />
      <ProductFeatureSplit config={config.featureSplit} />
      <ProductFeatureCards config={config.featureCards} />
      <ProductFaq items={config.faq.items} />
      <GrowOnlineCta variant="centered" />
      <SiteFooter />
    </div>
  );
}

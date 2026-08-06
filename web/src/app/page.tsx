import BrandTech from "@/components/sections/BrandTech";
import Hero from "@/components/sections/Hero";
import OwnerStories from "@/components/sections/OwnerStories";
import RatingsProof from "@/components/sections/RatingsProof";
import ReportMockup from "@/components/sections/ReportMockup";
import ValueFeatures from "@/components/sections/ValueFeatures";
import ValueHeading from "@/components/sections/ValueHeading";
import SiteFooter from "@/components/layout/SiteFooter";

export default function HomePage() {
  return (
    <>
      <Hero />
      <ReportMockup />
      <OwnerStories />
      <ValueHeading />
      <ValueFeatures />
      <RatingsProof />
      <BrandTech />
      <SiteFooter />
    </>
  );
}

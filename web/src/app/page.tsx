import BrandTech from "@/components/sections/BrandTech";
import DualFeatureCards from "@/components/sections/DualFeatureCards";
import GrowOnlineCta from "@/components/sections/GrowOnlineCta";
import GuidesResources from "@/components/sections/GuidesResources";
import Hero from "@/components/sections/Hero";
import OwnerStories from "@/components/sections/OwnerStories";
import RatingsProof from "@/components/sections/RatingsProof";
import ReportMockup from "@/components/sections/ReportMockup";
import TrustedByOwners from "@/components/sections/TrustedByOwners";
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
      <DualFeatureCards />
      <TrustedByOwners />
      <GuidesResources />
      <GrowOnlineCta />
      <SiteFooter />
    </>
  );
}

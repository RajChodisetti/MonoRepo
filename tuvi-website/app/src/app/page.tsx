import type { Metadata } from "next";
import Nav from "@/components/layout/Nav";
import Footer from "@/components/layout/Footer";
import Hero from "@/components/sections/Hero";
import GoogleWorkspaceApp from "@/components/sections/GoogleWorkspaceApp";
import WhatWeBuild from "@/components/sections/WhatWeBuild";
import HowWeWork from "@/components/sections/HowWeWork";
import StatsStrip from "@/components/sections/StatsStrip";
import AboutSection from "@/components/sections/AboutSection";
import TestimonialsMarquee from "@/components/sections/TestimonialsMarquee";
import GuaranteeSection from "@/components/sections/GuaranteeSection";
import ContactCTA from "@/components/sections/ContactCTA";
import VoiceAssistantWidget from "@/components/VoiceAssistantWidget";
import { siteContent } from "@/content/site";

export const metadata: Metadata = {
  title: "Tuvi Solutions | Software Studio & tuvi Google Workspace App",
  description: `${siteContent.hero.subcopy} ${siteContent.oauthApp.metadataDescription}`,
  alternates: { canonical: "/" },
  openGraph: {
    title: "Tuvi Solutions | Software Studio & tuvi Google Workspace App",
    description: siteContent.oauthApp.metadataDescription,
    type: "website",
    url: "/",
  },
};

export default function HomePage() {
  return (
    <>
      <Nav />
      <main id="main-content" tabIndex={-1}>
        <Hero />
        <GoogleWorkspaceApp />
        <WhatWeBuild />
        <HowWeWork />
        <StatsStrip />
        <AboutSection />
        <TestimonialsMarquee />
        <GuaranteeSection />
        <ContactCTA />
      </main>
      <Footer />
      <VoiceAssistantWidget />
    </>
  );
}

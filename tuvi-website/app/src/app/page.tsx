import type { Metadata } from "next";
import Nav from "@/components/layout/Nav";
import Footer from "@/components/layout/Footer";
import Hero from "@/components/sections/Hero";
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
  title: "Tuvi Solutions | AI Systems, Websites & Apps",
  description: siteContent.hero.subcopy,
  alternates: { canonical: "/" },
  openGraph: {
    title: "Tuvi Solutions | AI Systems, Websites & Apps",
    description: siteContent.hero.subcopy,
    type: "website",
    url: "/",
  },
};

export default function HomePage() {
  return (
    <div data-theme="dark" className="min-h-screen bg-bg text-text">
      <Nav />
      <main id="main-content" tabIndex={-1}>
        <Hero />
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
    </div>
  );
}

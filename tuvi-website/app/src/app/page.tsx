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

export default function HomePage() {
  return (
    <>
      <Nav />
      <main>
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
    </>
  );
}

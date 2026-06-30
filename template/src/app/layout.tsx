import type { Metadata } from "next";
import { Cormorant_Garamond, Outfit, Inter, Space_Grotesk } from "next/font/google";
import SmoothScrollProvider from "@/components/SmoothScrollProvider";
import TemplateShell from "@/components/TemplateShell";
import { getActiveTemplate } from "@/lib/templateConfig";
import "./globals.css";

const cinematicDisplay = Cormorant_Garamond({
  subsets: ["latin"],
  variable: "--font-cinematic-display",
  weight: ["400", "600", "700"],
});

const cinematicBody = Outfit({
  subsets: ["latin"],
  variable: "--font-cinematic-body",
  weight: ["300", "400", "500", "600"],
});

const auroraDisplay = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-aurora-display",
  weight: ["400", "500", "600", "700"],
});

const auroraBody = Inter({
  subsets: ["latin"],
  variable: "--font-aurora-body",
  weight: ["300", "400", "500", "600"],
});

export const metadata: Metadata = {
  title: "Restaurant Demo | Tuvi",
  description: "Premium restaurant website templates",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const templateId = getActiveTemplate();

  return (
    <html
      lang="en"
      data-template={templateId}
      className={`${cinematicDisplay.variable} ${cinematicBody.variable} ${auroraDisplay.variable} ${auroraBody.variable}`}
    >
      <body>
        <TemplateShell templateId={templateId}>
          <SmoothScrollProvider>{children}</SmoothScrollProvider>
        </TemplateShell>
      </body>
    </html>
  );
}

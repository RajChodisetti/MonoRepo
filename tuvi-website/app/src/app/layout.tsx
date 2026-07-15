import type { Metadata } from "next";
import { Instrument_Sans, Newsreader } from "next/font/google";
import SmoothScrollProvider from "@/components/SmoothScrollProvider";
import { siteContent } from "@/content/site";
import "./globals.css";

const newsreader = Newsreader({
  subsets: ["latin"],
  variable: "--font-newsreader",
  weight: ["500", "600", "700"],
});

const instrument = Instrument_Sans({
  subsets: ["latin"],
  variable: "--font-instrument",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  metadataBase: new URL("https://tuvisolutions.com"),
  title: `${siteContent.brand.name} | ${siteContent.brand.tagline}`,
  description: siteContent.hero.subcopy,
  icons: {
    icon: [{ url: "/brand/tuvi-solutions-icon.svg", type: "image/svg+xml" }],
  },
  openGraph: {
    title: siteContent.brand.name,
    description: siteContent.hero.subcopy,
    type: "website",
    images: ["/brand/tuvi-solutions-logo.png"],
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${newsreader.variable} ${instrument.variable}`}>
      <body className="font-body antialiased">
        <a href="#main-content" className="skip-link">
          Skip to content
        </a>
        <SmoothScrollProvider>{children}</SmoothScrollProvider>
      </body>
    </html>
  );
}

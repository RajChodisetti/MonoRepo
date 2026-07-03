import type { Metadata } from "next";
import { DM_Sans, Syne } from "next/font/google";
import SmoothScrollProvider from "@/components/SmoothScrollProvider";
import { siteContent } from "@/content/site";
import "./globals.css";

const syne = Syne({
  subsets: ["latin"],
  variable: "--font-syne",
  weight: ["500", "600", "700", "800"],
});

const dmSans = DM_Sans({
  subsets: ["latin"],
  variable: "--font-dm-sans",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: `${siteContent.brand.name} | ${siteContent.brand.tagline}`,
  description: siteContent.hero.subcopy,
  openGraph: {
    title: siteContent.brand.name,
    description: siteContent.hero.subcopy,
    type: "website",
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${syne.variable} ${dmSans.variable}`}>
      <body className="font-body antialiased">
        <SmoothScrollProvider>{children}</SmoothScrollProvider>
      </body>
    </html>
  );
}

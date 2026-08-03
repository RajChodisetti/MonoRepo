import type { Metadata } from "next";
import { Instrument_Sans, Newsreader } from "next/font/google";
import Navbar from "@/components/layout/Navbar";
import VoiceAssistantWidget from "@/components/voice/VoiceAssistantWidget";
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
  title: "Tuvi — Grow restaurant sales online",
  description:
    "Tuvi is the AI platform restaurants use to grow discovery, first-party orders, and repeat guests — under your brand.",
  icons: {
    icon: [{ url: "/brand/tuvi-solutions-icon.svg", type: "image/svg+xml" }],
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${newsreader.variable} ${instrument.variable} h-full antialiased`}
    >
      <body className="flex min-h-full flex-col bg-bg font-sans text-ink">
        <Navbar />
        <main className="flex-1">{children}</main>
        <VoiceAssistantWidget />
      </body>
    </html>
  );
}

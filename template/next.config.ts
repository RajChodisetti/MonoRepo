import path from "node:path";
import type { NextConfig } from "next";

const configuredMediaBase = process.env.NEXT_PUBLIC_MEDIA_BASE_URL?.trim();
let configuredMediaPattern:
  | { protocol: "https"; hostname: string; port: string; pathname: string }
  | undefined;
if (configuredMediaBase) {
  try {
    const parsed = new URL(configuredMediaBase);
    if (parsed.protocol === "https:") {
      configuredMediaPattern = {
        protocol: "https",
        hostname: parsed.hostname,
        port: parsed.port,
        pathname: `${parsed.pathname.replace(/\/$/, "") || ""}/**`,
      };
    }
  } catch {
    // Backend configuration validation reports the actionable URL error.
  }
}

const nextConfig: NextConfig = {
  // Restaurant fixtures are intentionally shared from the monorepo-level
  // `data/` directory. Next 16 Turbopack will not resolve outside its root.
  turbopack: {
    root: path.resolve(process.cwd(), ".."),
  },
  env: {
    TEMPLATE: process.env.TEMPLATE ?? "1",
  },
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "lh3.googleusercontent.com",
      },
      {
        protocol: "https",
        hostname: "streetviewpixels-pa.googleapis.com",
      },
      {
        protocol: "https",
        hostname: "maps.googleapis.com",
      },
      ...(configuredMediaPattern ? [configuredMediaPattern] : []),
    ],
  },
};

export default nextConfig;

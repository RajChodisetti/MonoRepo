import type { NextConfig } from "next";

const nextConfig: NextConfig = {
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
    ],
  },
};

export default nextConfig;

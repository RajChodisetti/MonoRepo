import type { NextConfig } from "next";

const immutableMediaHeaders = [
  {
    key: "Cache-Control",
    value: "public, max-age=31536000, immutable",
  },
];

const nextConfig: NextConfig = {
  async headers() {
    return [
      {
        source: "/services/restaurants/qr-ordering-kitchen-v3-web.mp4",
        headers: immutableMediaHeaders,
      },
      {
        source: "/services/restaurants/qr-ordering-kitchen-v3-web-poster.jpg",
        headers: immutableMediaHeaders,
      },
      {
        source: "/services/restaurants/rewards-reception-v4-web.mp4",
        headers: immutableMediaHeaders,
      },
      {
        source: "/services/restaurants/rewards-reception-v4-web-poster.jpg",
        headers: immutableMediaHeaders,
      },
    ];
  },
};

export default nextConfig;

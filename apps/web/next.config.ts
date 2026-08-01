import type { NextConfig } from "next";

// Deployed at https://api.tuvisolutions.com/admin (same host as the API,
// path-routed by Caddy) rather than a new subdomain. Locally this stays ""
// so `npm run dev` keeps working at the site root.
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || "";

const nextConfig: NextConfig = {
  basePath: basePath || undefined,
  experimental: {
    // The admin proxy forwards restaurant image uploads. Keep this slightly
    // above the backend's 15 MB hard limit without allowing unbounded bodies.
    middlewareClientMaxBodySize: "16mb",
  },
};

export default nextConfig;

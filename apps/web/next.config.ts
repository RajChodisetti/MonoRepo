import type { NextConfig } from "next";

// Deployed at https://api.tuvisolutions.com/admin (same host as the API,
// path-routed by Caddy) rather than a new subdomain. Locally this stays ""
// so `npm run dev` keeps working at the site root.
const basePath = process.env.NEXT_PUBLIC_BASE_PATH || "";

const nextConfig: NextConfig = {
  basePath: basePath || undefined,
};

export default nextConfig;

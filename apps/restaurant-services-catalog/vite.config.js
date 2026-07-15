import { defineConfig } from "vite";

export default defineConfig({
  server: {
    allowedHosts: [".trycloudflare.com"]
  },
  preview: {
    allowedHosts: [".trycloudflare.com"]
  }
});

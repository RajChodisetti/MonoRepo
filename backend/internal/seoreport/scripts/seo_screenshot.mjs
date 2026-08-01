/**
 * Capture a desktop viewport screenshot of a restaurant website.
 * Usage: node seo_screenshot.mjs <url>
 * Writes PNG bytes to stdout.
 *
 * Requires: npm i playwright && npx playwright install chromium
 * (run from MonoRepo/backend/internal/seoreport or with NODE_PATH set)
 */
import { chromium } from "playwright";

const url = process.argv[2];
if (!url) {
  console.error("usage: node seo_screenshot.mjs <url>");
  process.exit(2);
}

const browser = await chromium.launch({ headless: true });
try {
  const page = await browser.newPage({
    viewport: { width: 1280, height: 800 },
    userAgent:
      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
  });
  await page.goto(url, { waitUntil: "domcontentloaded", timeout: 25000 });
  await page.waitForTimeout(1500);
  const buf = await page.screenshot({ type: "png", fullPage: false });
  process.stdout.write(buf);
} finally {
  await browser.close();
}

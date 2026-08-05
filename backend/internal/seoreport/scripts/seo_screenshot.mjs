/**
 * Capture a restaurant website screenshot with light anti-bot hardening.
 * Usage: node seo_screenshot.mjs <url> [desktop|mobile]
 * Writes PNG bytes to stdout.
 * Exits 3 when the page is a bot/WAF block (Cloudflare, etc.).
 *
 * Requires: npm i playwright && npx playwright install chromium
 */
import { chromium, devices } from "playwright";

const url = process.argv[2];
const mode = (process.argv[3] || "desktop").toLowerCase();
if (!url) {
  console.error("usage: node seo_screenshot.mjs <url> [desktop|mobile]");
  process.exit(2);
}

const BLOCK_RE =
  /sorry,\s*you have been blocked|you have been blocked|access denied|attention required|checking your browser|cf-browser-verification|cf-challenge|just a moment|enable javascript and cookies|why have i been blocked|ray id\s*[:#]|unusual traffic|automated requests|bot detection|request blocked|security service to protect|captcha/i;

function isBlocked(text = "", html = "") {
  const blob = `${text}\n${html}`.slice(0, 80_000);
  return BLOCK_RE.test(blob);
}

const isMobile = mode === "mobile" || mode === "m";

const browser = await chromium.launch({
  headless: true,
  args: [
    "--disable-blink-features=AutomationControlled",
    "--no-sandbox",
    "--disable-dev-shm-usage",
    "--disable-infobars",
  ],
});

try {
  const iPhone = devices["iPhone 13"];
  const contextOptions = isMobile
    ? {
        ...iPhone,
        viewport: { width: 390, height: 844 },
        deviceScaleFactor: 2,
        isMobile: true,
        hasTouch: true,
        locale: "en-AU",
        timezoneId: "Australia/Sydney",
        extraHTTPHeaders: {
          "Accept-Language": "en-AU,en;q=0.9",
          "Upgrade-Insecure-Requests": "1",
        },
      }
    : {
        viewport: { width: 1280, height: 800 },
        userAgent:
          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
        locale: "en-AU",
        timezoneId: "Australia/Sydney",
        extraHTTPHeaders: {
          "Accept-Language": "en-AU,en;q=0.9",
          "Upgrade-Insecure-Requests": "1",
        },
      };

  const context = await browser.newContext(contextOptions);
  await context.addInitScript(() => {
    Object.defineProperty(navigator, "webdriver", { get: () => undefined });
    // eslint-disable-next-line no-extend-native
    window.chrome = { runtime: {} };
  });

  const page = await context.newPage();
  page.setDefaultTimeout(28000);

  const response = await page.goto(url, {
    waitUntil: "domcontentloaded",
    timeout: 28000,
  });
  // Give JS challenges a brief moment, then settle.
  await page.waitForTimeout(isMobile ? 2500 : 2000);
  try {
    await page.waitForLoadState("networkidle", { timeout: 4000 });
  } catch {
    /* ignore */
  }

  const status = response?.status() ?? 0;
  const title = await page.title().catch(() => "");
  const text = await page.locator("body").innerText({ timeout: 3000 }).catch(() => "");
  const html = await page.content().catch(() => "");

  if (status === 403 || status === 503 || isBlocked(`${title}\n${text}`, html)) {
    console.error("blocked_by_waf");
    process.exit(3);
  }

  const buf = await page.screenshot({ type: "png", fullPage: false });
  process.stdout.write(buf);
  await context.close();
} finally {
  await browser.close();
}

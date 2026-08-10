import assert from "node:assert/strict";
import test from "node:test";
import { seoForwardingHeaders } from "./seo-forwarding.ts";

test("forwards the Caddy client chain to the private Go hop", () => {
  const request = new Request("https://tuvisolutions.com/api/restaurants/search", {
    headers: {
      "x-forwarded-for": "192.0.2.10, 203.0.113.20",
      "x-real-ip": "203.0.113.20",
    },
  });
  assert.deepEqual(seoForwardingHeaders(request), {
    "X-Forwarded-For": "192.0.2.10, 203.0.113.20",
    "X-Real-IP": "203.0.113.20",
  });
});

test("drops malformed or oversized forwarding values", () => {
  const malformed = new Request("https://tuvisolutions.com/api/leads", {
    headers: {
      "x-forwarded-for": "203.0.113.20;evil",
      "x-real-ip": "not-an-ip",
    },
  });
  assert.deepEqual(seoForwardingHeaders(malformed), {});

  const oversized = new Request("https://tuvisolutions.com/api/leads", {
    headers: { "x-forwarded-for": "1".repeat(513) },
  });
  assert.deepEqual(seoForwardingHeaders(oversized), {});
});

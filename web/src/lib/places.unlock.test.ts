import assert from "node:assert/strict";
import test from "node:test";
import { getSeoReport } from "./places.ts";

test("report unlock bearer is sent only in Authorization, never the URL", async () => {
  const originalFetch = globalThis.fetch;
  const token = "a1".repeat(24);
  let requestedURL = "";
  let requestedAuthorization = "";
  globalThis.fetch = (async (input: string | URL | Request, init?: RequestInit) => {
    requestedURL = String(input);
    requestedAuthorization = new Headers(init?.headers).get("authorization") || "";
    return new Response(
      JSON.stringify({
        place: { placeId: "place-1", name: "Test", address: "", source: "places" },
        report: { fullReportLocked: false },
      }),
      { status: 200, headers: { "content-type": "application/json" } },
    );
  }) as typeof fetch;

  try {
    await getSeoReport("place-1", token, { "X-Forwarded-For": "203.0.113.10" });
  } finally {
    globalThis.fetch = originalFetch;
  }

  assert.equal(requestedAuthorization, `Bearer ${token}`);
  assert.equal(requestedURL.includes(token), false);
  assert.equal(requestedURL.includes("unlock="), false);
});

test("report API preserves busy status and retry timing for the BFF", async () => {
  const originalFetch = globalThis.fetch;
  globalThis.fetch = (async () =>
    new Response(JSON.stringify({ error: "report busy" }), {
      status: 503,
      headers: { "content-type": "application/json", "retry-after": "3" },
    })) as typeof fetch;

  try {
    await assert.rejects(
      () => getSeoReport("place-1"),
      (error: unknown) => {
        const typed = error as Error & { status?: number; retryAfter?: string };
        assert.equal(typed.status, 503);
        assert.equal(typed.retryAfter, "3");
        return true;
      },
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});

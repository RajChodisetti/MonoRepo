import assert from "node:assert/strict";
import test from "node:test";
import { searchSeoRestaurants } from "./places.ts";

test("search results retain every supplied Place data attribution", async (t) => {
  const originalFetch = globalThis.fetch;
  t.after(() => {
    globalThis.fetch = originalFetch;
  });

  globalThis.fetch = (async () =>
    new Response(
      JSON.stringify({
        results: [
          {
            placeId: "place-1",
            name: "Attributed Restaurant",
            address: "1 Main St",
            source: "places",
            attributions: [
              {
                provider: "Local data partner",
                providerUri: "https://provider.example/place-1",
              },
              { providerUri: "https://provider.example/unnamed" },
              {},
            ],
          },
        ],
        meta: { placesEnabled: true, inventoryEnabled: false },
      }),
      { status: 200, headers: { "Content-Type": "application/json" } },
    )) as typeof fetch;

  const payload = await searchSeoRestaurants("attributed", "Sydney");
  assert.deepEqual(payload.results[0]?.attributions, [
    {
      provider: "Local data partner",
      providerUri: "https://provider.example/place-1",
    },
    { provider: "", providerUri: "https://provider.example/unnamed" },
  ]);
});

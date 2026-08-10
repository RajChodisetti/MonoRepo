import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readFile } from "node:fs/promises";
import test from "node:test";

const webRoot = new URL("../../", import.meta.url);
const portraitDigest = "f8265cba7daf13383cf0254f615e10e3377523f3bd652466a5a17b68cf88de5a";

const assetUrls = [
  new URL("public/menu/garlic-bread.jpg", webRoot),
  new URL("public/menu/garlic-bread-v2.jpg", webRoot),
  new URL("public/menu/churros.jpg", webRoot),
];

test("menu food assets cannot regress to the portrait placeholder", async () => {
  for (const assetUrl of assetUrls) {
    const bytes = await readFile(assetUrl);
    const digest = createHash("sha256").update(bytes).digest("hex");

    assert.notEqual(digest, portraitDigest, assetUrl.pathname);
    assert.deepEqual([...bytes.subarray(0, 3)], [0xff, 0xd8, 0xff], assetUrl.pathname);
  }
});

test("menu cards use cache-busted, dish-specific assets", async () => {
  const mappings = [
    ["src/components/sections/features/OnlineSalesPanel.tsx", "/menu/garlic-bread-v2.jpg"],
    ["src/components/product/visuals/CateringVisuals.tsx", "/menu/garlic-bread-v2.jpg"],
    ["src/components/product/visuals/AiPhoneVisuals.tsx", "/menu/garlic-bread-v2.jpg"],
    ["src/components/product/visuals/MenuVisuals.tsx", "/menu/churros.jpg"],
    ["src/components/product/visuals/UpsellsVisuals.tsx", "/menu/churros.jpg"],
  ] as const;

  for (const [sourcePath, expectedAsset] of mappings) {
    const source = await readFile(new URL(sourcePath, webRoot), "utf8");
    assert.ok(source.includes(expectedAsset), `${sourcePath} must reference ${expectedAsset}`);
    assert.ok(!source.includes("/menu/garlic-bread.jpg"), `${sourcePath} uses the stale asset URL`);
  }
});

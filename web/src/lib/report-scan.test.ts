import assert from "node:assert/strict";
import test from "node:test";
import {
  normalizeScanPhotos,
  normalizeScanReviews,
  reportMapEmbedUrl,
  websiteCaptureEvidence,
} from "./report-scan.ts";

test("scan photos omit empty and duplicate sources without capping available evidence", () => {
  const photos = Array.from({ length: 15 }, (_, index) => ({
    src: `/photo-${index}.jpg`,
    label: `Photo ${index}`,
  }));
  const normalized = normalizeScanPhotos(
    " /hero.jpg ",
    "Restaurant",
    [{ src: "" }, { src: "/hero.jpg", label: "duplicate" }, ...photos],
  );

  assert.equal(normalized.length, 16);
  assert.deepEqual(normalized[0], { src: "/hero.jpg", label: "Restaurant" });
  assert.equal(normalized.at(-1)?.src, "/photo-14.jpg");
});

test("scan reviews retain every genuine review and discard empty payload shells", () => {
  const reviews = Array.from({ length: 15 }, (_, index) => ({
    author: ` Reviewer ${index} `,
    text: ` Review ${index} `,
    rating: 4,
  }));
  const normalized = normalizeScanReviews([
    {},
    { author: " ", text: "", rating: 0 },
    ...reviews,
  ]);

  assert.equal(normalized.length, 15);
  assert.deepEqual(normalized[0], {
    author: "Reviewer 0",
    text: "Review 0",
    rating: 4,
  });
});

test("website captures render only distinct available desktop and mobile evidence", () => {
  assert.deepEqual(websiteCaptureEvidence("", undefined), []);
  assert.deepEqual(websiteCaptureEvidence("desktop.jpg", "mobile.jpg"), [
    { kind: "desktop", src: "desktop.jpg" },
    { kind: "mobile", src: "mobile.jpg" },
  ]);
  assert.deepEqual(websiteCaptureEvidence("same.jpg", "same.jpg"), [
    { kind: "desktop", src: "same.jpg" },
  ]);
});

test("map embeds preserve exact coordinates and do not invent zero coordinates", () => {
  const exact = reportMapEmbedUrl({
    restaurantName: "Exact Place",
    latitude: -37.85,
    longitude: 144.99,
  });
  assert.match(exact || "", /q=-37\.85%2C144\.99/);
  assert.equal(reportMapEmbedUrl({ restaurantName: "Your restaurant" }), null);
});

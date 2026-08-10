import assert from "node:assert/strict";
import test from "node:test";
import {
  buildScanPhotoSlots,
  buildScanReviewStream,
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

test("scan photo board fills six real cards and assigns overflow photos to flips", () => {
  const photos = Array.from({ length: 10 }, (_, index) => ({
    src: `/photo-${index}.jpg`,
    label: `Photo ${index}`,
  }));
  const slots = buildScanPhotoSlots(photos);

  assert.equal(slots.length, 6);
  assert.deepEqual(slots.map((slot) => slot.map((photo) => photo.src)), [
    ["/photo-0.jpg", "/photo-6.jpg"],
    ["/photo-1.jpg", "/photo-7.jpg"],
    ["/photo-2.jpg", "/photo-8.jpg"],
    ["/photo-3.jpg", "/photo-9.jpg"],
    ["/photo-4.jpg"],
    ["/photo-5.jpg"],
  ]);
  assert.equal(buildScanPhotoSlots(photos.slice(0, 3)).length, 3);
  assert.deepEqual(buildScanPhotoSlots([]), []);
});

test("scan review board creates an honest ten-card stream from available reviews", () => {
  const reviews = Array.from({ length: 5 }, (_, index) => ({
    author: `Reviewer ${index}`,
    text: `Review ${index}`,
  }));
  const stream = buildScanReviewStream(reviews);

  assert.equal(stream.length, 10);
  assert.deepEqual(stream.map((item) => item.sourceIndex), [0, 1, 2, 3, 4, 0, 1, 2, 3, 4]);
  assert.deepEqual(stream.map((item) => item.repeated), [
    false, false, false, false, false, true, true, true, true, true,
  ]);
  assert.deepEqual(buildScanReviewStream([]), []);
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

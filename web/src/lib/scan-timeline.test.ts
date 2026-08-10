import assert from "node:assert/strict";
import test from "node:test";
import {
  EVIDENCE_BATCH_SIZE,
  EVIDENCE_CARD_ENTRY_MS,
  EVIDENCE_CARD_STAGGER_MS,
  EVIDENCE_DWELL_MS,
  EVIDENCE_LAST_CARD_HOLD_MS,
  MAX_SCAN_MS,
  LISTING_CARD_COUNT,
  MIN_SCAN_MS,
  PHOTO_FLIP_HOLD_MS,
  PHOTO_FLIP_MS,
  REVIEW_PAGE_SIZE,
  REVIEW_WALL_HOLD_MS,
  REVIEW_WALL_LIMIT,
  SATELLITE_ARC_CLEARANCE_DEG,
  dealPhotosToCards,
  evidenceBatchPresentationMs,
  evidenceCardEntryDelayMs,
  isScanCompletionReady,
  nextEvidenceBatchStart,
  nextPhotoFaceIndex,
  photoFlipCycleMs,
  satellitePositions,
} from "./scan-timeline.ts";

const EPSILON = 1e-9;
function assertClose(actual: number, expected: number, message: string) {
  assert.ok(Math.abs(actual - expected) < EPSILON, `${message}: ${actual} !~ ${expected}`);
}

test("scan always presents for at least twenty seconds", () => {
  assert.equal(
    isScanCompletionReady({
      elapsedMs: MIN_SCAN_MS - 1,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: null,
    }),
    false,
  );
  assert.equal(
    isScanCompletionReady({
      elapsedMs: MIN_SCAN_MS,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: null,
    }),
    true,
  );
});

test("a six-card step enters one-by-one, a second apart, and holds the last completed card", () => {
  assert.equal(EVIDENCE_BATCH_SIZE, 6);
  assert.equal(EVIDENCE_CARD_STAGGER_MS, 1_000);
  assert.equal(EVIDENCE_CARD_ENTRY_MS, 900);
  assert.equal(EVIDENCE_LAST_CARD_HOLD_MS, 3_000);
  assert.deepEqual(
    Array.from({ length: EVIDENCE_BATCH_SIZE }, (_, index) =>
      evidenceCardEntryDelayMs(index),
    ),
    [0, 1_000, 2_000, 3_000, 4_000, 5_000],
  );

  const lastCardFullyVisibleAt =
    evidenceCardEntryDelayMs(EVIDENCE_BATCH_SIZE - 1) + EVIDENCE_CARD_ENTRY_MS;
  assert.equal(evidenceBatchPresentationMs(EVIDENCE_BATCH_SIZE), 8_900);
  assert.equal(EVIDENCE_DWELL_MS - lastCardFullyVisibleAt, 3_000);
});

test("short remainder batches keep the same final-card hold", () => {
  assert.equal(evidenceBatchPresentationMs(0), 0);
  assert.equal(evidenceBatchPresentationMs(1), 3_900);
  assert.equal(evidenceBatchPresentationMs(2), 4_900);
  assert.equal(evidenceBatchPresentationMs(99), EVIDENCE_DWELL_MS);
  assert.equal(evidenceCardEntryDelayMs(-2), 0);
});

test("late real evidence stays open for one complete paced batch", () => {
  const receivedAt = MIN_SCAN_MS - 1_000;
  assert.equal(
    isScanCompletionReady({
      elapsedMs: MIN_SCAN_MS,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: receivedAt,
    }),
    false,
  );
  assert.equal(
    isScanCompletionReady({
      elapsedMs: receivedAt + EVIDENCE_DWELL_MS,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: receivedAt,
    }),
    true,
  );
});

test("scan requires fetch, preserves the evidence dwell, and retains its cap", () => {
  assert.equal(
    isScanCompletionReady({
      elapsedMs: MAX_SCAN_MS,
      fetchComplete: false,
      evidenceReadyAtElapsedMs: MAX_SCAN_MS - 100,
    }),
    false,
  );
  // Evidence that arrived before the minimum still gets its full dwell.
  const readyAt = MIN_SCAN_MS - 5_000;
  assert.ok(readyAt + EVIDENCE_DWELL_MS < MAX_SCAN_MS);
  assert.equal(
    isScanCompletionReady({
      elapsedMs: readyAt + EVIDENCE_DWELL_MS - 1,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: readyAt,
    }),
    false,
  );
  // The cap still wins over a dwell that would outrun it.
  assert.equal(
    isScanCompletionReady({
      elapsedMs: MAX_SCAN_MS,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: MIN_SCAN_MS,
    }),
    true,
  );
  assert.equal(
    isScanCompletionReady({
      elapsedMs: MAX_SCAN_MS,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: MAX_SCAN_MS - 100,
    }),
    true,
  );
});

test("a listing photo rests two seconds before the card turns", () => {
  assert.equal(PHOTO_FLIP_HOLD_MS, 2_000);
  assert.equal(PHOTO_FLIP_MS, 900);
  assert.equal(photoFlipCycleMs(), 2_900);
});

test("the hidden face is always loaded one photo ahead and wraps", () => {
  assert.equal(nextPhotoFaceIndex(0, 4), 1);
  assert.equal(nextPhotoFaceIndex(1, 4), 2);
  assert.equal(nextPhotoFaceIndex(3, 4), 0);
  assert.equal(nextPhotoFaceIndex(7, 4), 0);
  // A single photo never turns, so the card keeps its only face.
  assert.equal(nextPhotoFaceIndex(0, 1), 0);
  assert.equal(nextPhotoFaceIndex(3, 1), 0);
  assert.equal(nextPhotoFaceIndex(0, 0), 0);
  assert.equal(nextPhotoFaceIndex(-4, 5), 1);
  assert.equal(nextPhotoFaceIndex(Number.NaN, 5), 1);
});

test("listing photos spread across six cards before any card repeats", () => {
  assert.equal(LISTING_CARD_COUNT, 6);
  // Fewer photos than cards never invents an empty card.
  assert.deepEqual(dealPhotosToCards(["a", "b", "c"], 6), [["a"], ["b"], ["c"]]);
  // Every card gets a first photo before any card gets a second.
  assert.deepEqual(
    dealPhotosToCards(["a", "b", "c", "d", "e", "f", "g", "h", "i"], 6),
    [["a", "g"], ["b", "h"], ["c", "i"], ["d"], ["e"], ["f"]],
  );
  assert.deepEqual(dealPhotosToCards([], 6), []);
  assert.deepEqual(dealPhotosToCards(["a"], 0), []);
});

test("the review wall shows two reviews a page and wraps at the sample it was given", () => {
  assert.equal(REVIEW_WALL_LIMIT, 10);
  assert.equal(REVIEW_PAGE_SIZE, 2);
  assert.equal(REVIEW_WALL_HOLD_MS, 3_000);
  // Reuses the same advance-and-wrap rule as the evidence collage — a page is
  // just a batch of size REVIEW_PAGE_SIZE.
  assert.equal(nextEvidenceBatchStart(0, 7, REVIEW_PAGE_SIZE), 2);
  assert.equal(nextEvidenceBatchStart(2, 7, REVIEW_PAGE_SIZE), 4);
  assert.equal(nextEvidenceBatchStart(4, 7, REVIEW_PAGE_SIZE), 6);
  // Seven reviews page as [0,1] [2,3] [4,5] [6] — the last page is a lone
  // review, then wraps back to the start.
  assert.equal(nextEvidenceBatchStart(6, 7, REVIEW_PAGE_SIZE), 0);
  // A single review never pages.
  assert.equal(nextEvidenceBatchStart(0, 1, REVIEW_PAGE_SIZE), 0);
});

test("satellites skip the search-chip wedge and space out around the hero", () => {
  assert.deepEqual(satellitePositions(0), []);

  // A lone satellite centres in the open arc, straight down from the hero.
  const [solo] = satellitePositions(1);
  assertClose(solo.leftPercent, 50, "solo satellite left");
  assertClose(solo.topPercent, 83, "solo satellite top");
  assert.equal(solo.rotateDeg, -4);

  // Six cards — the busiest step (photo quality) — stay spread and in bounds.
  const six = satellitePositions(6);
  assert.equal(six.length, 6);
  for (const point of six) {
    assert.ok(point.leftPercent > 0 && point.leftPercent < 100, "left in bounds");
    assert.ok(point.topPercent > 0 && point.topPercent < 100, "top in bounds");
  }
  // No two cards land near enough to plausibly overlap.
  const MIN_SEPARATION_PERCENT = 20;
  for (let i = 0; i < six.length; i += 1) {
    for (let j = i + 1; j < six.length; j += 1) {
      const dx = six[i].leftPercent - six[j].leftPercent;
      const dy = six[i].topPercent - six[j].topPercent;
      const distance = Math.sqrt(dx * dx + dy * dy);
      assert.ok(
        distance >= MIN_SEPARATION_PERCENT,
        `satellites ${i} and ${j} are only ${distance.toFixed(1)}% apart`,
      );
    }
  }
  // Rotation alternates sign so the scatter doesn't lean one way.
  assert.deepEqual(
    six.map((p) => p.rotateDeg),
    [-4, 6, -8, 4, -6, 8],
  );

  // Every satellite, at any count, stays clear of the wedge reserved for the
  // search chip directly above the hero.
  for (const count of [1, 2, 3, 4, 5, 6, 7]) {
    for (const point of satellitePositions(count)) {
      const dxFromCenter = point.leftPercent - 50;
      const dyFromCenter = point.topPercent - 50;
      const degFromUp = (Math.atan2(dxFromCenter, -dyFromCenter) * 180) / Math.PI;
      const normalized = (degFromUp + 360) % 360;
      assert.ok(
        normalized >= SATELLITE_ARC_CLEARANCE_DEG &&
          normalized <= 360 - SATELLITE_ARC_CLEARANCE_DEG,
        `satellite at ${normalized.toFixed(1)}deg sits under the search chip`,
      );
    }
  }
});

test("evidence rotation advances whole batches and wraps after the remainder", () => {
  assert.equal(nextEvidenceBatchStart(0, 0, 4), 0);
  assert.equal(nextEvidenceBatchStart(0, 4, 4), 0);
  assert.equal(nextEvidenceBatchStart(0, 16, 4), 4);
  assert.equal(nextEvidenceBatchStart(4, 16, 4), 8);
  assert.equal(nextEvidenceBatchStart(8, 16, 4), 12);
  assert.equal(nextEvidenceBatchStart(12, 16, 4), 0);
  assert.equal(nextEvidenceBatchStart(0, 10, 4), 4);
  assert.equal(nextEvidenceBatchStart(4, 10, 4), 8);
  assert.equal(nextEvidenceBatchStart(8, 10, 4), 0);
  assert.equal(nextEvidenceBatchStart(0, 10, 0), 0);
});

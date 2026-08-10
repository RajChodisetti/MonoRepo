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
  REVIEW_WALL_HOLD_MS,
  REVIEW_WALL_LIMIT,
  dealPhotosToCards,
  evidenceBatchPresentationMs,
  evidenceCardEntryDelayMs,
  isScanCompletionReady,
  nextEvidenceBatchStart,
  nextPhotoFaceIndex,
  nextReviewIndex,
  photoFlipCycleMs,
} from "./scan-timeline.ts";

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

test("eight-card batches enter one-by-one and hold the last completed card", () => {
  assert.equal(EVIDENCE_BATCH_SIZE, 8);
  assert.equal(EVIDENCE_CARD_STAGGER_MS, 750);
  assert.equal(EVIDENCE_CARD_ENTRY_MS, 900);
  assert.equal(EVIDENCE_LAST_CARD_HOLD_MS, 3_000);
  assert.deepEqual(
    Array.from({ length: EVIDENCE_BATCH_SIZE }, (_, index) =>
      evidenceCardEntryDelayMs(index),
    ),
    [0, 750, 1_500, 2_250, 3_000, 3_750, 4_500, 5_250],
  );

  const lastCardFullyVisibleAt =
    evidenceCardEntryDelayMs(EVIDENCE_BATCH_SIZE - 1) + EVIDENCE_CARD_ENTRY_MS;
  assert.equal(evidenceBatchPresentationMs(EVIDENCE_BATCH_SIZE), 9_150);
  assert.equal(EVIDENCE_DWELL_MS - lastCardFullyVisibleAt, 3_000);
});

test("short remainder batches keep the same final-card hold", () => {
  assert.equal(evidenceBatchPresentationMs(0), 0);
  assert.equal(evidenceBatchPresentationMs(1), 3_900);
  assert.equal(evidenceBatchPresentationMs(2), 4_650);
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

test("a listing photo rests three seconds before the card turns", () => {
  assert.equal(PHOTO_FLIP_HOLD_MS, 3_000);
  assert.equal(PHOTO_FLIP_MS, 900);
  assert.equal(photoFlipCycleMs(), 3_900);
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

test("listing photos spread across five cards before any card repeats", () => {
  assert.equal(LISTING_CARD_COUNT, 5);
  // Fewer photos than cards never invents an empty card.
  assert.deepEqual(dealPhotosToCards(["a", "b", "c"], 5), [["a"], ["b"], ["c"]]);
  // Every card gets a first photo before any card gets a second.
  assert.deepEqual(dealPhotosToCards(["a", "b", "c", "d", "e", "f", "g"], 5), [
    ["a", "f"],
    ["b", "g"],
    ["c"],
    ["d"],
    ["e"],
  ]);
  assert.deepEqual(dealPhotosToCards([], 5), []);
  assert.deepEqual(dealPhotosToCards(["a"], 0), []);
});

test("the review wall holds each review and wraps at the sample it was given", () => {
  assert.equal(REVIEW_WALL_LIMIT, 10);
  assert.equal(REVIEW_WALL_HOLD_MS, 3_000);
  assert.equal(nextReviewIndex(0, 5), 1);
  assert.equal(nextReviewIndex(4, 5), 0);
  assert.equal(nextReviewIndex(0, 1), 0);
  assert.equal(nextReviewIndex(0, 0), 0);
  assert.equal(nextReviewIndex(Number.NaN, 3), 1);
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

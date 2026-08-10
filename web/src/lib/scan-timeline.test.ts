import assert from "node:assert/strict";
import test from "node:test";
import {
  EVIDENCE_BATCH_SIZE,
  EVIDENCE_CARD_ENTRY_MS,
  EVIDENCE_CARD_STAGGER_MS,
  EVIDENCE_DWELL_MS,
  EVIDENCE_LAST_CARD_HOLD_MS,
  MAX_SCAN_MS,
  MIN_SCAN_MS,
  evidenceBatchPresentationMs,
  evidenceCardEntryDelayMs,
  isScanCompletionReady,
  nextEvidenceBatchStart,
} from "./scan-timeline.ts";

test("scan always presents for at least fifteen seconds", () => {
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

test("four-card batches enter one-by-one and hold the last completed card", () => {
  assert.equal(EVIDENCE_BATCH_SIZE, 4);
  assert.equal(EVIDENCE_CARD_STAGGER_MS, 750);
  assert.equal(EVIDENCE_CARD_ENTRY_MS, 900);
  assert.equal(EVIDENCE_LAST_CARD_HOLD_MS, 3_000);
  assert.deepEqual(
    Array.from({ length: EVIDENCE_BATCH_SIZE }, (_, index) =>
      evidenceCardEntryDelayMs(index),
    ),
    [0, 750, 1_500, 2_250],
  );

  const lastCardFullyVisibleAt =
    evidenceCardEntryDelayMs(EVIDENCE_BATCH_SIZE - 1) + EVIDENCE_CARD_ENTRY_MS;
  assert.equal(evidenceBatchPresentationMs(EVIDENCE_BATCH_SIZE), 6_150);
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
  assert.equal(
    isScanCompletionReady({
      elapsedMs: MIN_SCAN_MS + EVIDENCE_DWELL_MS - 1,
      fetchComplete: true,
      evidenceReadyAtElapsedMs: MIN_SCAN_MS,
    }),
    false,
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

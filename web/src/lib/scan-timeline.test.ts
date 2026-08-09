import assert from "node:assert/strict";
import test from "node:test";
import {
  EVIDENCE_DWELL_MS,
  MAX_SCAN_MS,
  MIN_SCAN_MS,
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

test("late real evidence stays open for a full four-second frame", () => {
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

test("scan never completes before fetch and caps late-evidence waiting", () => {
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

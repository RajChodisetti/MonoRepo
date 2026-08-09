export const MIN_SCAN_MS = 15_000;
export const EVIDENCE_DWELL_MS = 4_000;
export const MAX_SCAN_MS = 23_000;
export const TARGET_SCAN_SECONDS = 15;

export function isScanCompletionReady({
  elapsedMs,
  fetchComplete,
  evidenceReadyAtElapsedMs,
}: {
  elapsedMs: number;
  fetchComplete: boolean;
  /** Milliseconds since this scan mounted, in the same time coordinate as elapsedMs. */
  evidenceReadyAtElapsedMs: number | null;
}): boolean {
  if (!fetchComplete || elapsedMs < MIN_SCAN_MS) return false;
  if (elapsedMs >= MAX_SCAN_MS) return true;
  return (
    evidenceReadyAtElapsedMs === null ||
    elapsedMs - evidenceReadyAtElapsedMs >= EVIDENCE_DWELL_MS
  );
}

export function nextEvidenceBatchStart(
  currentIndex: number,
  evidenceCount: number,
  batchSize: number,
): number {
  if (evidenceCount <= 0 || batchSize <= 0 || evidenceCount <= batchSize) return 0;
  const next = Math.max(0, currentIndex) + batchSize;
  return next >= evidenceCount ? 0 : next;
}

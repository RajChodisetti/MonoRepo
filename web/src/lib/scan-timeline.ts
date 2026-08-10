export const MIN_SCAN_MS = 15_000;
export const MAX_SCAN_MS = 23_000;
export const TARGET_SCAN_SECONDS = 15;
export const EVIDENCE_BATCH_SIZE = 4;
export const EVIDENCE_CARD_STAGGER_MS = 750;
export const EVIDENCE_CARD_ENTRY_MS = 900;
export const EVIDENCE_LAST_CARD_HOLD_MS = 3_000;

export function evidenceCardEntryDelayMs(cardIndex: number): number {
  return Math.max(0, Math.floor(cardIndex)) * EVIDENCE_CARD_STAGGER_MS;
}

/**
 * Time from mounting a batch until it is safe to replace it. The final card's
 * entry must finish before its hold begins, so a four-card batch takes longer
 * than the hold by the three stagger gaps and the entry animation itself.
 */
export function evidenceBatchPresentationMs(cardCount: number): number {
  const safeCardCount = Math.min(
    EVIDENCE_BATCH_SIZE,
    Math.max(0, Math.floor(cardCount)),
  );
  if (safeCardCount === 0) return 0;
  return (
    evidenceCardEntryDelayMs(safeCardCount - 1) +
    EVIDENCE_CARD_ENTRY_MS +
    EVIDENCE_LAST_CARD_HOLD_MS
  );
}

export const EVIDENCE_DWELL_MS = evidenceBatchPresentationMs(EVIDENCE_BATCH_SIZE);

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

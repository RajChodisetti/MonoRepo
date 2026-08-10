export const MIN_SCAN_MS = 20_000;
export const MAX_SCAN_MS = 28_000;
export const TARGET_SCAN_SECONDS = 20;
/** 6 listing + 2 website + 1 competitor: every card mounts at once, no rotation. */
export const EVIDENCE_BATCH_SIZE = 9;
/** Listing photos spread across this many cards before any card repeats. */
export const LISTING_CARD_COUNT = 6;
/** Gap between each card's entrance so the collage doesn't land all at once. */
export const EVIDENCE_CARD_STAGGER_MS = 1_000;
export const EVIDENCE_CARD_ENTRY_MS = 900;
export const EVIDENCE_LAST_CARD_HOLD_MS = 3_000;
/** A listing photo rests this long before the card turns to the next one. */
export const PHOTO_FLIP_HOLD_MS = 2_000;
/** The turn itself, kept slow enough to read as a deliberate flip. */
export const PHOTO_FLIP_MS = 900;
/** Google Maps returns a relevance-ordered sample; the wall never shows more. */
export const REVIEW_WALL_LIMIT = 10;
/** Each review holds for the same beat as a photo before the wall advances. */
export const REVIEW_WALL_HOLD_MS = 3_000;

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

/**
 * Deal decoded photos across the listing cards so every card owns distinct
 * photos, and a card only carries a second photo once each card has a first.
 */
export function dealPhotosToCards<T>(photos: readonly T[], cardCount: number): T[][] {
  if (photos.length === 0 || cardCount <= 0) return [];
  const cards = Math.min(Math.floor(cardCount), photos.length);
  const dealt: T[][] = Array.from({ length: cards }, () => []);
  photos.forEach((photo, index) => dealt[index % cards].push(photo));
  return dealt;
}

/** Rest plus turn: how long one photo owns the card before the next appears. */
export function photoFlipCycleMs(): number {
  return PHOTO_FLIP_HOLD_MS + PHOTO_FLIP_MS;
}

/**
 * The face hidden behind the card is loaded one photo ahead, so advancing wraps
 * the index rather than running off the end of a short gallery.
 */
export function nextPhotoFaceIndex(current: number, photoCount: number): number {
  if (photoCount <= 1) return 0;
  const safe = Number.isFinite(current) && current > 0 ? Math.floor(current) : 0;
  return (safe + 1) % photoCount;
}

/** Advance through the review wall, wrapping at the relevance-ordered sample. */
export function nextReviewIndex(current: number, reviewCount: number): number {
  if (reviewCount <= 1) return 0;
  const safe = Number.isFinite(current) && current > 0 ? Math.floor(current) : 0;
  return (safe + 1) % reviewCount;
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

export const MIN_SCAN_MS = 20_000;
export const MAX_SCAN_MS = 28_000;
export const TARGET_SCAN_SECONDS = 20;
/**
 * The richest single step (photo quality, up to 6 cards) sets how long the
 * scan waits after evidence becomes ready before it's allowed to finish. Not
 * a visual batch size — each step shows only its own relevant cards, but this
 * is still the right "how much is there to see" figure for that wait.
 */
export const EVIDENCE_BATCH_SIZE = 6;
/** Listing photos spread across this many cards before any card repeats. */
export const LISTING_CARD_COUNT = 6;
/** Gap between each card's entrance so a step's cards don't land all at once. */
export const EVIDENCE_CARD_STAGGER_MS = 1_000;
export const EVIDENCE_CARD_ENTRY_MS = 900;
export const EVIDENCE_LAST_CARD_HOLD_MS = 3_000;
/** A listing photo rests this long before the card turns to the next one. */
export const PHOTO_FLIP_HOLD_MS = 2_000;
/** The turn itself, kept slow enough to read as a deliberate flip. */
export const PHOTO_FLIP_MS = 900;
/** Google Maps returns a relevance-ordered sample; the wall never shows more. */
export const REVIEW_WALL_LIMIT = 10;
/** Two reviews share the wall at once before it slides to the next pair. */
export const REVIEW_PAGE_SIZE = 2;
/** Each page holds for the same beat as a photo before the wall advances. */
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

export function nextEvidenceBatchStart(
  currentIndex: number,
  evidenceCount: number,
  batchSize: number,
): number {
  if (evidenceCount <= 0 || batchSize <= 0 || evidenceCount <= batchSize) return 0;
  const next = Math.max(0, currentIndex) + batchSize;
  return next >= evidenceCount ? 0 : next;
}

/**
 * The wedge left clear at the top of the ring, in degrees either side of
 * straight up, for the search chip that always sits there.
 */
export const SATELLITE_ARC_CLEARANCE_DEG = 50;
export const SATELLITE_ARC_SWEEP_DEG = 360 - 2 * SATELLITE_ARC_CLEARANCE_DEG;
export const SATELLITE_RADIUS_X_PERCENT = 40;
export const SATELLITE_RADIUS_Y_PERCENT = 33;

export type SatellitePosition = {
  leftPercent: number;
  topPercent: number;
  rotateDeg: number;
};

/**
 * Positions for the cards orbiting a centred hero card, spaced by construction
 * rather than picked by hand — so however many a step has to show (a lone
 * competitor card, six listing photos), none of them can land on top of each
 * other. Angles run clockwise from straight up and skip the wedge reserved
 * for the search chip; each satellite sits at the centre of its own equal
 * slice of the remaining arc, which keeps small counts away from that
 * cleared wedge too rather than pinning them to its edges.
 */
export function satellitePositions(satelliteCount: number): SatellitePosition[] {
  const count = Math.max(0, Math.floor(satelliteCount));
  if (count === 0) return [];
  return Array.from({ length: count }, (_, i) => {
    const sliceCenter = (i + 0.5) / count;
    const deg = SATELLITE_ARC_CLEARANCE_DEG + SATELLITE_ARC_SWEEP_DEG * sliceCenter;
    const rad = (deg * Math.PI) / 180;
    return {
      leftPercent: 50 + SATELLITE_RADIUS_X_PERCENT * Math.sin(rad),
      topPercent: 50 - SATELLITE_RADIUS_Y_PERCENT * Math.cos(rad),
      // Deterministic alternating wobble — no randomness, so server and
      // client render the same markup on the first paint.
      rotateDeg: i % 2 === 0 ? -(4 + (i % 3) * 2) : 4 + (i % 3) * 2,
    };
  });
}

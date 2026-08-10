export type ScanPhotoEvidence = {
  src: string;
  label?: string;
};

export type ScanReviewEvidence = {
  author?: string;
  text?: string;
  rating?: number;
  relativeTime?: string;
  sentiment?: string;
};

export type WebsiteCaptureEvidence = {
  kind: "desktop" | "mobile";
  src: string;
};

export type ScanReviewStreamItem = {
  review: ScanReviewEvidence;
  sourceIndex: number;
  repeated: boolean;
};

/** Keep every distinct photo with a usable source. Empty cards are never data. */
export function normalizeScanPhotos(
  primarySrc: string | undefined,
  restaurantName: string,
  photos: ScanPhotoEvidence[],
): ScanPhotoEvidence[] {
  const seen = new Set<string>();
  const normalized: ScanPhotoEvidence[] = [];

  const append = (src: string | undefined, label?: string) => {
    const cleanSrc = src?.trim();
    if (!cleanSrc || seen.has(cleanSrc)) return;
    seen.add(cleanSrc);
    const cleanLabel = label?.trim();
    normalized.push({
      src: cleanSrc,
      ...(cleanLabel ? { label: cleanLabel } : {}),
    });
  };

  append(primarySrc, restaurantName);
  for (const photo of photos) append(photo.src, photo.label);
  return normalized;
}

/** Remove payload shells that contain no review evidence; never synthesize copy. */
export function normalizeScanReviews(reviews: ScanReviewEvidence[]): ScanReviewEvidence[] {
  return reviews.flatMap((review) => {
    const author = review.author?.trim();
    const text = review.text?.trim();
    const relativeTime = review.relativeTime?.trim();
    const sentiment = review.sentiment?.trim();
    const rating =
      typeof review.rating === "number" &&
      Number.isFinite(review.rating) &&
      review.rating > 0 &&
      review.rating <= 5
        ? review.rating
        : undefined;

    if (!author && !text && !relativeTime && rating === undefined) return [];
    return [{
      ...(author ? { author } : {}),
      ...(text ? { text } : {}),
      ...(rating !== undefined ? { rating } : {}),
      ...(relativeTime ? { relativeTime } : {}),
      ...(sentiment ? { sentiment } : {}),
    }];
  });
}

/**
 * Fill at most six board positions, then distribute overflow photos across the
 * occupied positions so each card can flip without ever rendering an empty
 * placeholder.
 */
export function buildScanPhotoSlots(
  photos: ScanPhotoEvidence[],
  maximumSlots = 6,
): ScanPhotoEvidence[][] {
  const slotLimit = Math.max(0, Math.floor(maximumSlots));
  const slotCount = Math.min(slotLimit, photos.length);
  if (slotCount === 0) return [];

  const slots = Array.from({ length: slotCount }, () => [] as ScanPhotoEvidence[]);
  photos.forEach((photo, index) => {
    slots[index % slotCount].push(photo);
  });
  return slots;
}

/**
 * Build a ten-card visual stream from genuine review evidence. Google Places
 * currently returns at most five review texts, so repeats are flagged for
 * assistive technology instead of being represented as additional reviews.
 */
export function buildScanReviewStream(
  reviews: ScanReviewEvidence[],
  targetCards = 10,
): ScanReviewStreamItem[] {
  const cardCount = Math.max(0, Math.floor(targetCards));
  if (reviews.length === 0 || cardCount === 0) return [];

  return Array.from({ length: cardCount }, (_, index) => {
    const sourceIndex = index % reviews.length;
    return {
      review: reviews[sourceIndex],
      sourceIndex,
      repeated: index >= reviews.length,
    };
  });
}

/** Return only real, distinct viewport captures in desktop-first order. */
export function websiteCaptureEvidence(
  desktopSrc?: string,
  mobileSrc?: string,
): WebsiteCaptureEvidence[] {
  const desktop = desktopSrc?.trim();
  const mobile = mobileSrc?.trim();
  const captures: WebsiteCaptureEvidence[] = [];
  if (desktop) captures.push({ kind: "desktop", src: desktop });
  if (mobile && mobile !== desktop) captures.push({ kind: "mobile", src: mobile });
  return captures;
}

export function reportMapEmbedUrl(options: {
  restaurantName: string;
  address?: string;
  placeId?: string;
  latitude?: number;
  longitude?: number;
}): string | null {
  const { latitude, longitude } = options;
  if (
    typeof latitude === "number" &&
    typeof longitude === "number" &&
    Number.isFinite(latitude) &&
    Number.isFinite(longitude) &&
    Math.abs(latitude) <= 90 &&
    Math.abs(longitude) <= 180
  ) {
    const params = new URLSearchParams({
      q: `${latitude},${longitude}`,
      ll: `${latitude},${longitude}`,
      z: "17",
      hl: "en",
      output: "embed",
    });
    return `https://www.google.com/maps?${params.toString()}`;
  }

  if (options.placeId) {
    const label = [options.restaurantName, options.address].filter(Boolean).join(", ");
    const params = new URLSearchParams({
      q: label || `place_id:${options.placeId}`,
      query_place_id: options.placeId,
      z: "17",
      hl: "en",
      output: "embed",
    });
    return `https://www.google.com/maps?${params.toString()}`;
  }

  const query = [options.restaurantName, options.address].filter(Boolean).join(" ");
  if (!query || query === "Your restaurant") return null;
  const params = new URLSearchParams({ q: query, z: "16", hl: "en", output: "embed" });
  return `https://www.google.com/maps?${params.toString()}`;
}

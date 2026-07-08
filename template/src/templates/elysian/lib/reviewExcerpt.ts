import type { RestaurantContent } from "@/data/types/restaurant";

const MAX_REVIEWS = 5;
const MAX_CHARS = 150;

/** One or two short sentences — keeps the carousel readable. */
export function excerptReview(text: string, maxChars = MAX_CHARS): string {
  const trimmed = text.trim().replace(/\s+/g, " ");
  if (!trimmed) return "";

  const sentences = trimmed.match(/[^.!?]+[.!?]+/g);
  if (sentences) {
    let block = "";
    for (const sentence of sentences.slice(0, 2)) {
      const next = (block + sentence).trim();
      if (next.length > maxChars && block) break;
      block = next;
      if (block.length >= 80) break;
    }
    if (block.length <= maxChars) return block.trim();
  }

  if (trimmed.length <= maxChars) return trimmed;
  const slice = trimmed.slice(0, maxChars);
  const lastSpace = slice.lastIndexOf(" ");
  return `${(lastSpace > 60 ? slice.slice(0, lastSpace) : slice).trim()}…`;
}

export function reviewsForCarousel(
  reviews: RestaurantContent["reviews"],
  max = MAX_REVIEWS,
): RestaurantContent["reviews"] {
  if (reviews.length <= max) return reviews;
  return [...reviews]
    .sort((a, b) => a.review.length - b.review.length)
    .slice(0, max);
}

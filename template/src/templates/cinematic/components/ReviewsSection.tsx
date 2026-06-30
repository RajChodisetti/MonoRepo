"use client";

import { useState } from "react";
import type { Review } from "@/data/types/reviews";

const PREVIEW_CHARS = 200;

function stars(n: number) {
  const full = Math.min(5, Math.round(n));
  return "★".repeat(full) + "☆".repeat(5 - full);
}

function ReviewCard({ review }: { review: Review }) {
  const [expanded, setExpanded] = useState(false);
  const isLong = review.review.length > PREVIEW_CHARS;
  const displayText =
    expanded || !isLong
      ? review.review
      : `${review.review.slice(0, PREVIEW_CHARS).trim()}…`;

  return (
    <blockquote className="flex h-full flex-col rounded-xl border border-[#e8e0d4] bg-white p-6 shadow-[0_4px_24px_rgba(26,22,20,0.08)]">
      <p className="text-sm tracking-wide text-[#b88a44]">{stars(review.stars)}</p>
      <p className="mt-4 flex-1 font-display text-lg leading-relaxed text-[#1a1614]">
        &ldquo;{displayText}&rdquo;
      </p>
      {isLong && (
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="mt-3 self-start text-xs font-semibold uppercase tracking-[0.12em] text-[#8a7340] transition hover:text-[#b88a44]"
        >
          {expanded ? "View less" : "View more"}
        </button>
      )}
      <footer className="mt-4 border-t border-[#e8e0d4] pt-4 text-sm font-medium text-[#5c534c]">
        {review.reviewer}
        {review.date && (
          <span className="mt-1 block text-xs font-normal text-[#a89f96]">{review.date}</span>
        )}
      </footer>
    </blockquote>
  );
}

export default function ReviewsSection({
  reviews,
  rating,
  reviewsCount,
}: {
  reviews: Review[];
  rating?: number;
  reviewsCount?: number;
}) {
  if (!reviews.length) return null;

  return (
    <section id="reviews" className="bg-[#F7F0E6] py-24">
      <div className="mx-auto max-w-6xl px-6">
        <div className="text-center">
          <p className="text-xs uppercase tracking-[0.2em] text-[#8a7340]">Guest Voices</p>
          <h2 className="font-display mt-3 text-4xl text-[#1a1614] md:text-5xl">What People Say</h2>
          {rating && (
            <p className="mt-3 text-[#5c534c]">
              {rating} average from {reviewsCount || reviews.length} reviews
            </p>
          )}
        </div>

        <div className="mt-12 grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {reviews.map((rev, i) => (
            <ReviewCard key={`${rev.reviewer}-${i}`} review={rev} />
          ))}
        </div>
      </div>
    </section>
  );
}

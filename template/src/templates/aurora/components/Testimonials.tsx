"use client";

import { useState } from "react";
import GlassCard from "./ui/GlassCard";
import BlurReveal from "./ui/BlurReveal";
import type { Review } from "@/data/types/reviews";

const PREVIEW_CHARS = 200;

function stars(n: number) {
  const full = Math.min(5, Math.round(n));
  return "★".repeat(full) + "☆".repeat(5 - full);
}

function TestimonialCard({ review }: { review: Review }) {
  const [expanded, setExpanded] = useState(false);
  const isLong = review.review.length > PREVIEW_CHARS;
  const displayText =
    expanded || !isLong
      ? review.review
      : `${review.review.slice(0, PREVIEW_CHARS).trim()}…`;

  return (
    <GlassCard hover={false} className="flex h-full flex-col p-6 md:p-8">
      <p className="text-sm tracking-wide text-purple-400">{stars(review.stars)}</p>
      <blockquote className="aurora-heading mt-4 flex-1 text-lg leading-relaxed text-white/90 md:text-xl">
        &ldquo;{displayText}&rdquo;
      </blockquote>
      {isLong && (
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="mt-3 self-start text-xs font-semibold uppercase tracking-[0.12em] text-cyan-300 transition hover:text-cyan-200"
        >
          {expanded ? "View less" : "View more"}
        </button>
      )}
      <footer className="mt-4 border-t border-white/10 pt-4 text-sm text-white/50">
        {review.reviewer}
        {review.date && (
          <span className="mt-1 block text-xs text-white/30">{review.date}</span>
        )}
      </footer>
    </GlassCard>
  );
}

export default function Testimonials({
  reviews,
  rating,
}: {
  reviews: Review[];
  rating?: number;
}) {
  if (!reviews.length) return null;

  return (
    <section id="reviews" className="aurora-section">
      <div className="aurora-container">
        <BlurReveal className="text-center">
          <p className="text-xs uppercase tracking-[0.2em] text-blue-400">Social Proof</p>
          <h2 className="aurora-heading mt-3 text-4xl font-bold text-white">
            Customer Testimonials
          </h2>
          {rating && (
            <p className="mt-3 text-purple-400">{rating} average rating</p>
          )}
        </BlurReveal>

        <div className="mt-12 grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {reviews.map((rev, i) => (
            <BlurReveal key={`${rev.reviewer}-${i}`} delay={i * 0.08}>
              <TestimonialCard review={rev} />
            </BlurReveal>
          ))}
        </div>
      </div>
    </section>
  );
}

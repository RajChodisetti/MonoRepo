"use client";

export default function ElysianScrollTop({ visible }: { visible: boolean }) {
  return (
    <button
      type="button"
      className={`scroll-top${visible ? " show" : ""}`}
      id="scrollTop"
      aria-label="Scroll to top"
      onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
    >
      <svg viewBox="0 0 24 24">
        <path
          d="M12 19V5M5 12l7-7 7 7"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </button>
  );
}

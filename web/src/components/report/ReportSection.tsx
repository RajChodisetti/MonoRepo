import type { ReactNode } from "react";

/** Shared bordered panel for systematic report layout. */
export default function ReportSection({
  eyebrow,
  title,
  children,
  className = "",
  bodyClassName = "",
}: {
  eyebrow?: string;
  title?: string;
  children: ReactNode;
  className?: string;
  bodyClassName?: string;
}) {
  return (
    <section
      className={`overflow-hidden rounded-[22px] border border-border bg-bg/90 shadow-[0_10px_36px_rgba(15,39,31,0.06)] ${className}`}
    >
      {eyebrow || title ? (
        <header className="border-b border-border px-5 py-4 sm:px-6">
          {eyebrow ? (
            <p className="text-[12px] font-semibold uppercase tracking-[0.08em] text-muted">{eyebrow}</p>
          ) : null}
          {title ? (
            <h2
              className={`font-display text-[1.15rem] font-semibold tracking-[-0.02em] text-ink sm:text-[1.25rem] ${
                eyebrow ? "mt-1.5" : ""
              }`}
            >
              {title}
            </h2>
          ) : null}
        </header>
      ) : null}
      <div className={bodyClassName || "px-5 py-5 sm:px-6 sm:py-6"}>{children}</div>
    </section>
  );
}

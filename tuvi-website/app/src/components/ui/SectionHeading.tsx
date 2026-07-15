export default function SectionHeading({
  eyebrow,
  title,
  description,
  align = "left",
}: {
  eyebrow: string;
  title: string;
  description?: string;
  align?: "left" | "center";
}) {
  const centered = align === "center";

  return (
    <div className={`mb-12 max-w-2xl ${centered ? "mx-auto text-center" : ""}`}>
      <span
        className={`flex items-center gap-2.5 text-xs font-semibold uppercase tracking-[0.18em] text-muted ${
          centered ? "justify-center" : ""
        }`}
      >
        <span className="h-2 w-2 rounded-[2px] bg-primary" aria-hidden />
        {eyebrow}
      </span>
      <h2 className="mt-4 font-display text-4xl font-semibold leading-[1.02] tracking-[-0.03em] text-ink md:text-5xl lg:text-6xl">
        {title}
      </h2>
      {description && (
        <p className="mt-4 text-base leading-relaxed text-muted md:text-lg">{description}</p>
      )}
    </div>
  );
}

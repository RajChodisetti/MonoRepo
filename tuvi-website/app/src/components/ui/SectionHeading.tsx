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
  const alignClass = align === "center" ? "text-center mx-auto" : "";

  return (
    <div className={`mb-12 max-w-2xl ${alignClass}`}>
      <p className="text-[11px] font-bold uppercase tracking-[0.2em] text-cyan">{eyebrow}</p>
      <h2 className="mt-3 font-display text-3xl font-bold leading-tight text-text md:text-4xl lg:text-5xl">
        {title}
      </h2>
      {description && (
        <p className="mt-4 text-base leading-relaxed text-muted md:text-lg">{description}</p>
      )}
    </div>
  );
}

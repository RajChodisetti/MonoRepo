import Image from "next/image";
import Link from "next/link";

type BrandLogoProps = {
  href?: string;
  size?: "nav" | "footer" | "hero";
  showName?: boolean;
  priority?: boolean;
  className?: string;
};

const frameClasses = {
  nav: "h-11 w-11",
  footer: "h-20 w-20",
  hero: "h-full w-full",
} as const;

export default function BrandLogo({
  href = "/",
  size = "nav",
  showName = true,
  priority = false,
  className = "",
}: BrandLogoProps) {
  const content = (
    <>
      <span
        className={`relative shrink-0 overflow-hidden rounded-full bg-[#fffef8] ${frameClasses[size]}`}
      >
        <Image
          src="/brand/tuvi-solutions-logo.png"
          alt=""
          fill
          sizes={size === "hero" ? "(max-width: 768px) 82vw, 520px" : size === "footer" ? "80px" : "44px"}
          className="scale-[1.72] object-contain"
          priority={priority}
        />
      </span>
      {showName ? (
        <span className={size === "nav" ? "hidden leading-none sm:block" : "leading-none"}>
          <span className="block font-display text-[1.05rem] font-semibold tracking-[-0.02em] text-ink md:text-lg">
            Tuvi Solutions
          </span>
          {size !== "nav" ? (
            <span className="mt-1.5 block text-[0.62rem] font-semibold uppercase tracking-[0.22em] text-muted">
              Software built with strength
            </span>
          ) : null}
        </span>
      ) : null}
    </>
  );

  return (
    <Link
      href={href}
      aria-label="Tuvi Solutions home"
      className={`inline-flex items-center gap-2.5 ${className}`}
    >
      {content}
    </Link>
  );
}

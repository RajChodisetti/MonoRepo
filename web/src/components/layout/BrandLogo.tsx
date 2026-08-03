import Image from "next/image";
import Link from "next/link";

type BrandLogoProps = {
  href?: string;
  onNavigate?: () => void;
  priority?: boolean;
};

export default function BrandLogo({
  href = "/",
  onNavigate,
  priority = true,
}: BrandLogoProps) {
  return (
    <Link
      href={href}
      onClick={onNavigate}
      className="inline-flex items-center gap-2.5 text-ink"
      aria-label="Tuvi home"
    >
      <span className="relative h-9 w-9 shrink-0 overflow-hidden rounded-full bg-surface ring-1 ring-border">
        <Image
          src="/brand/tuvi-solutions-logo.png"
          alt=""
          fill
          sizes="36px"
          className="scale-[1.72] object-contain"
          priority={priority}
        />
      </span>
      <span className="font-display text-[1.4rem] font-semibold tracking-[-0.03em]">Tuvi</span>
    </Link>
  );
}

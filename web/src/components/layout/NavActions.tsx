import Link from "next/link";

type NavActionsProps = {
  onNavigate?: () => void;
  className?: string;
};

export default function NavActions({ onNavigate, className = "" }: NavActionsProps) {
  return (
    <div className={`items-center gap-2.5 sm:gap-3 ${className}`}>
      <Link
        href="/how-it-works"
        onClick={onNavigate}
        className="inline-flex shrink-0 items-center justify-center whitespace-nowrap rounded-full bg-surface px-5 py-2.5 text-[14px] font-semibold text-ink transition-colors hover:bg-parchment sm:px-6 sm:text-[15px]"
      >
        See how it works
      </Link>
      <Link
        href="/book"
        onClick={onNavigate}
        className="inline-flex shrink-0 items-center justify-center whitespace-nowrap rounded-full bg-primary px-5 py-2.5 text-[14px] font-semibold text-bg transition-colors hover:bg-primary-dim sm:px-6 sm:text-[15px]"
      >
        Get a free demo
      </Link>
    </div>
  );
}

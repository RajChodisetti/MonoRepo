import Link from "next/link";

type NavActionsProps = {
  onNavigate?: () => void;
  className?: string;
};

export default function NavActions({ onNavigate, className = "" }: NavActionsProps) {
  return (
    <div className={`items-center gap-5 ${className}`}>
      <Link
        href="/login"
        onClick={onNavigate}
        className="text-[15px] font-semibold text-ink transition-colors hover:text-primary"
      >
        Login
      </Link>
      <Link
        href="/demo"
        onClick={onNavigate}
        className="inline-flex items-center justify-center rounded-full bg-primary px-5 py-2.5 text-[15px] font-semibold text-bg transition-colors hover:bg-primary-dim"
      >
        Get a free demo
      </Link>
    </div>
  );
}

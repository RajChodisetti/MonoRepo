import Link from "next/link";

type ButtonProps = {
  href: string;
  children: React.ReactNode;
  variant?: "primary" | "ghost";
  external?: boolean;
  className?: string;
};

export default function Button({
  href,
  children,
  variant = "primary",
  external = false,
  className = "",
}: ButtonProps) {
  const base =
    "inline-flex cursor-pointer items-center justify-center gap-2 rounded-full px-6 py-3 text-sm font-semibold transition-colors duration-200";
  const styles =
    variant === "primary"
      ? "bg-ink text-bg shadow-[0_12px_32px_-14px_rgba(15,39,31,0.7)] hover:bg-primary hover:text-white"
      : "border border-border bg-bg-elevated text-ink hover:border-primary hover:bg-surface";

  const classes = `${base} ${styles} ${className}`;

  if (external || href.startsWith("http")) {
    return (
      <a href={href} target="_blank" rel="noopener noreferrer" className={classes}>
        {children}
      </a>
    );
  }

  if (href.startsWith("#")) {
    return (
      <a href={href} className={classes}>
        {children}
      </a>
    );
  }

  return (
    <Link href={href} className={classes}>
      {children}
    </Link>
  );
}

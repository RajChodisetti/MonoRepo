"use client";

import { motion, useMotionValue, useSpring } from "framer-motion";

export default function MagneticButton({
  href,
  children,
  variant = "primary",
  className = "",
}: {
  href: string;
  children: React.ReactNode;
  variant?: "primary" | "secondary";
  className?: string;
}) {
  const x = useMotionValue(0);
  const y = useMotionValue(0);
  const sx = useSpring(x, { stiffness: 300, damping: 20 });
  const sy = useSpring(y, { stiffness: 300, damping: 20 });

  function onMove(e: React.MouseEvent<HTMLAnchorElement>) {
    const rect = e.currentTarget.getBoundingClientRect();
    x.set((e.clientX - rect.left - rect.width / 2) * 0.15);
    y.set((e.clientY - rect.top - rect.height / 2) * 0.15);
  }

  function onLeave() {
    x.set(0);
    y.set(0);
  }

  const base =
    variant === "primary" ? "aurora-btn-primary" : "aurora-btn-secondary";

  return (
    <motion.a
      href={href}
      style={{ x: sx, y: sy }}
      onMouseMove={onMove}
      onMouseLeave={onLeave}
      className={`${base} ${className}`}
    >
      {children}
    </motion.a>
  );
}

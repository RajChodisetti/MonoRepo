"use client";

import { useEffect, useState } from "react";
import { usePrefersReducedMotion } from "@/hooks/usePrefersReducedMotion";

export default function CursorGlow() {
  const reduced = usePrefersReducedMotion();
  const [pos, setPos] = useState({ x: 0, y: 0 });
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (reduced || window.matchMedia("(max-width: 768px)").matches) return;

    const onMove = (e: MouseEvent) => {
      setPos({ x: e.clientX, y: e.clientY });
      setVisible(true);
    };
    const onLeave = () => setVisible(false);

    window.addEventListener("mousemove", onMove);
    document.body.addEventListener("mouseleave", onLeave);
    return () => {
      window.removeEventListener("mousemove", onMove);
      document.body.removeEventListener("mouseleave", onLeave);
    };
  }, [reduced]);

  if (reduced) return null;

  return (
    <div
      className="pointer-events-none fixed z-[5] hidden md:block"
      style={{
        left: pos.x - 200,
        top: pos.y - 200,
        width: 400,
        height: 400,
        opacity: visible ? 0.15 : 0,
        transition: "opacity 0.3s",
        background: "radial-gradient(circle, rgba(124,58,237,0.4) 0%, transparent 70%)",
      }}
      aria-hidden
    />
  );
}

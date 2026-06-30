"use client";

import { motion } from "framer-motion";
import { useRef, useState } from "react";

export default function GlassCard({
  children,
  className = "",
  hover = true,
}: {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [transform, setTransform] = useState("");

  function onMove(e: React.MouseEvent) {
    if (!hover || !ref.current) return;
    const rect = ref.current.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width - 0.5;
    const y = (e.clientY - rect.top) / rect.height - 0.5;
    setTransform(
      `perspective(800px) rotateY(${x * 8}deg) rotateX(${-y * 8}deg) translateY(-4px)`
    );
  }

  function onLeave() {
    setTransform("");
  }

  return (
    <motion.div
      ref={ref}
      onMouseMove={onMove}
      onMouseLeave={onLeave}
      style={{ transform }}
      className={`glass-card ${hover ? "glass-card-hover" : ""} ${className}`}
    >
      {children}
    </motion.div>
  );
}

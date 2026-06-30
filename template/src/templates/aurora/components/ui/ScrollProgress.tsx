"use client";

import { motion, useScroll, useSpring } from "framer-motion";

export default function ScrollProgress() {
  const { scrollYProgress } = useScroll();
  const scaleX = useSpring(scrollYProgress, { stiffness: 100, damping: 30 });

  return (
    <motion.div
      className="fixed left-0 top-0 z-[60] h-0.5 w-full origin-left bg-gradient-to-r from-purple-500 via-blue-500 to-cyan-400"
      style={{ scaleX }}
    />
  );
}

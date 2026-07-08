"use client";

import { useEffect, useRef } from "react";

export default function ElysianCursor() {
  const glowRef = useRef<HTMLDivElement>(null);
  const dotRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let mouseX = 0;
    let mouseY = 0;
    let glowX = 0;
    let glowY = 0;
    let raf = 0;

    const onMove = (e: MouseEvent) => {
      mouseX = e.clientX;
      mouseY = e.clientY;
      if (dotRef.current) {
        dotRef.current.style.transform = `translate(${mouseX}px, ${mouseY}px) translate(-50%,-50%)`;
      }
    };

    const animate = () => {
      glowX += (mouseX - glowX) * 0.08;
      glowY += (mouseY - glowY) * 0.08;
      if (glowRef.current) {
        glowRef.current.style.transform = `translate(${glowX}px, ${glowY}px) translate(-50%,-50%)`;
      }
      raf = requestAnimationFrame(animate);
    };

    const interactive = "a, button, .dish-card, .masonry-item, input, select, textarea";
    const onEnter = () => dotRef.current?.classList.add("active");
    const onLeave = () => dotRef.current?.classList.remove("active");

    window.addEventListener("mousemove", onMove);
    raf = requestAnimationFrame(animate);

    const bind = () => {
      document.querySelectorAll(interactive).forEach((el) => {
        el.addEventListener("mouseenter", onEnter);
        el.addEventListener("mouseleave", onLeave);
      });
    };
    bind();
    const mo = new MutationObserver(bind);
    mo.observe(document.body, { childList: true, subtree: true });

    return () => {
      window.removeEventListener("mousemove", onMove);
      cancelAnimationFrame(raf);
      mo.disconnect();
    };
  }, []);

  return (
    <>
      <div className="cursor-glow" ref={glowRef} id="cursorGlow" />
      <div className="cursor-dot" ref={dotRef} id="cursorDot" />
    </>
  );
}

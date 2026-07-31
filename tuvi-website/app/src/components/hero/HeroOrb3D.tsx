"use client";

import { useEffect, useRef, useState } from "react";
import { useReducedMotion } from "framer-motion";

/**
 * Hero orb — voice-orb style organic canvas blob.
 * Canvas is oversized vs the layout slot so the soft glow fully fades
 * to transparent before any square edge — no visible box on black bg.
 */
export default function HeroOrb3D() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const wrapRef = useRef<HTMLDivElement>(null);
  const reduceMotion = useReducedMotion();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const canvas = canvasRef.current;
    const wrap = wrapRef.current;
    if (!canvas || !wrap) return;

    const preferReduced =
      reduceMotion ||
      (typeof window !== "undefined" &&
        window.matchMedia("(prefers-reduced-motion: reduce)").matches);

    // Explicit alpha so clear stays transparent on black page
    const ctx = canvas.getContext("2d", { alpha: true, desynchronized: true });
    if (!ctx) return;

    let slot = 320; // layout box (orb visual size reference)
    let canvasCss = 320; // actual canvas (larger, for soft edge fade)
    let raf = 0;
    let last = performance.now();
    let swirlPos = 0;
    let revealed = false;

    const vis = {
      energy: preferReduced ? 0.18 : 0.42,
      scale: 1,
      swirl: preferReduced ? 0.12 : 0.55,
    };

    // Extra padding so halo never hits the square canvas edge
    const PAD = 1.65;

    const resize = () => {
      const box = wrap.getBoundingClientRect();
      slot = Math.max(180, Math.min(box.width, box.height));
      canvasCss = Math.ceil(slot * PAD);
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      canvas.width = canvasCss * dpr;
      canvas.height = canvasCss * dpr;
      canvas.style.width = `${canvasCss}px`;
      canvas.style.height = `${canvasCss}px`;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    resize();
    const ro = new ResizeObserver(resize);
    ro.observe(wrap);

    const draw = (now: number) => {
      const dt = Math.min(0.05, (now - last) / 1000);
      last = now;
      const t = now / 1000;

      const breath = preferReduced ? 0.02 * Math.sin(t * 0.7) : 0.07 * Math.sin(t * 1.1);
      const pulse = preferReduced ? 0 : 0.05 * Math.sin(t * 2.4);
      const targetEnergy = (preferReduced ? 0.2 : 0.48) + breath + pulse;
      const targetScale = 1 + (preferReduced ? 0.015 : 0.04) * Math.sin(t * 0.65);

      const k = preferReduced ? 0.2 : 0.07;
      vis.energy += (targetEnergy - vis.energy) * k;
      vis.scale += (targetScale - vis.scale) * k;
      swirlPos += vis.swirl * dt * (preferReduced ? 0.4 : 1);

      const c = canvasCss / 2;
      // Size relative to layout slot (chips stay put); glow lives inside oversized canvas
      const baseR = slot * 0.4 * vis.scale;
      const def = preferReduced ? vis.energy * 0.35 : vis.energy * 0.85;

      ctx.clearRect(0, 0, canvasCss, canvasCss);
      ctx.globalCompositeOperation = "source-over";

      // Halo must fully reach alpha 0 well inside canvas (never clip to a square)
      const haloR = Math.min(baseR * 1.55, canvasCss * 0.48);
      const halo = ctx.createRadialGradient(c, c, baseR * 0.15, c, c, haloR);
      halo.addColorStop(0, `hsla(0,0%,52%,${0.14 + vis.energy * 0.1})`);
      halo.addColorStop(0.5, `hsla(150,14%,22%,${0.07 + vis.energy * 0.04})`);
      halo.addColorStop(0.82, "hsla(0,0%,0%,0.02)");
      halo.addColorStop(1, "hsla(0,0%,0%,0)");
      ctx.fillStyle = halo;
      ctx.beginPath();
      ctx.arc(c, c, haloR, 0, Math.PI * 2);
      ctx.fill();

      for (let L = 0; L < 4; L++) {
        const R = baseR * (1.08 - L * 0.12) * (1 + 0.035 * Math.sin(t * 0.85 + L));
        ctx.beginPath();
        for (let i = 0; i <= 96; i++) {
          const th = (i / 96) * Math.PI * 2;
          const w =
            0.45 * Math.sin(3 * th + t * 1.15 + L * 1.8) +
            0.28 * Math.sin(5 * th - t * 1.55 + L * 3.2) +
            0.16 * Math.sin(7 * th + t * 2.05 + L * 0.7) +
            0.08 * Math.sin(9 * th - t * 1.1 + L);
          const r = R * (1 + def * w);
          const a = th + swirlPos * (0.12 + L * 0.04);
          const x = c + Math.cos(a) * r;
          const y = c + Math.sin(a) * r * 0.97;
          if (i === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        }
        ctx.closePath();

        const g = ctx.createRadialGradient(
          c - R * 0.32,
          c - R * 0.35,
          R * 0.08,
          c,
          c,
          R * 1.15,
        );
        const light = 20 + L * 10;
        const sat = 5 + L * 2;
        g.addColorStop(0, `hsla(0,${sat}%,${Math.min(68, light + 36)}%,${0.52 - L * 0.08})`);
        g.addColorStop(0.5, `hsla(160,${8 + L}%,${light + 6}%,${0.36 - L * 0.06})`);
        g.addColorStop(1, "hsla(0,0%,0%,0)");
        ctx.fillStyle = g;
        ctx.fill();
      }

      const core = ctx.createRadialGradient(
        c - baseR * 0.18,
        c - baseR * 0.22,
        0,
        c,
        c,
        baseR * 0.48,
      );
      core.addColorStop(0, `hsla(0,0%,90%,${0.5 + vis.energy * 0.12})`);
      core.addColorStop(0.35, `hsla(0,0%,65%,0.18)`);
      core.addColorStop(0.7, `hsla(150,10%,18%,0.08)`);
      core.addColorStop(1, "hsla(0,0%,0%,0)");
      ctx.fillStyle = core;
      ctx.beginPath();
      ctx.arc(c, c, baseR * 0.48, 0, Math.PI * 2);
      ctx.fill();

      if (!revealed) {
        revealed = true;
        setReady(true);
      }

      raf = requestAnimationFrame(draw);
    };

    raf = requestAnimationFrame(draw);
    return () => {
      cancelAnimationFrame(raf);
      ro.disconnect();
    };
  }, [reduceMotion]);

  return (
    <div
      ref={wrapRef}
      className="pointer-events-none absolute inset-0 flex items-center justify-center overflow-visible bg-transparent"
      aria-hidden="true"
      // Soft circular mask kills any residual square canvas edge
      style={{
        opacity: ready ? 1 : 0,
        transition: reduceMotion ? undefined : "opacity 0.55s cubic-bezier(0.22, 1, 0.36, 1)",
        background: "transparent",
        WebkitMaskImage: "radial-gradient(circle, #000 42%, transparent 72%)",
        maskImage: "radial-gradient(circle, #000 42%, transparent 72%)",
      }}
    >
      <canvas
        ref={canvasRef}
        className="block bg-transparent"
        style={{ background: "transparent", backgroundColor: "transparent" }}
      />
    </div>
  );
}

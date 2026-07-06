"use client";

import { useEffect, useRef } from "react";
import type { VoiceSessionStatus } from "@/hooks/useVoiceAgentSession";

type OrbMood = "idle" | "listening" | "thinking" | "speaking";

const MOODS: Record<
  OrbMood,
  { hue: [number, number]; glow: string; energy: number; swirl: number; scale: number }
> = {
  idle: { hue: [168, 200], glow: "#4de8cf", energy: 0.12, swirl: 0.15, scale: 1 },
  listening: { hue: [196, 235], glow: "#59b8ff", energy: 0.55, swirl: 0.25, scale: 1.08 },
  thinking: { hue: [258, 288], glow: "#a58bff", energy: 0.28, swirl: 2.4, scale: 0.92 },
  speaking: { hue: [14, 36], glow: "#ff9d6b", energy: 0.42, swirl: 0.35, scale: 1.06 },
};

function toMood(status: VoiceSessionStatus): OrbMood {
  if (status === "listening" || status === "user-speaking") return "listening";
  if (status === "thinking") return "thinking";
  if (status === "speaking") return "speaking";
  return "idle";
}

type Props = {
  status: VoiceSessionStatus;
  size?: number;
  className?: string;
};

export default function VoiceOrbCanvas({ status, size = 148, className }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const moodRef = useRef<OrbMood>(toMood(status));
  const visRef = useRef({
    hueA: 168,
    hueB: 200,
    energy: 0.12,
    swirl: 0.15,
    scale: 1,
    swirlPos: 0,
  });

  useEffect(() => {
    moodRef.current = toMood(status);
  }, [status]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const preferReduced =
      typeof window !== "undefined" &&
      window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const css = size;
    canvas.width = css * dpr;
    canvas.height = css * dpr;
    canvas.style.width = `${css}px`;
    canvas.style.height = `${css}px`;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

    let raf = 0;
    let last = performance.now();

    const draw = (now: number) => {
      const dt = Math.min(0.05, (now - last) / 1000);
      last = now;
      const t = now / 1000;
      const mood = MOODS[moodRef.current];
      const vis = visRef.current;
      const k = preferReduced ? 0.25 : 0.08;

      const breath =
        moodRef.current === "idle"
          ? 0.04 * Math.sin(t * 0.9)
          : moodRef.current === "thinking"
            ? 0.06 * Math.sin(t * 3)
            : moodRef.current === "listening"
              ? 0.08 * Math.sin(t * 4.2)
              : 0.1 * Math.sin(t * 5.5);

      const targetEnergy = mood.energy + breath;
      const targetScale = mood.scale + (moodRef.current === "idle" ? 0.03 * Math.sin(t * 0.7) : 0);

      vis.hueA += (mood.hue[0] - vis.hueA) * 0.04;
      vis.hueB += (mood.hue[1] - vis.hueB) * 0.04;
      vis.energy += (targetEnergy - vis.energy) * k;
      vis.scale += (targetScale - vis.scale) * k;
      vis.swirl += (mood.swirl - vis.swirl) * 0.05;
      vis.swirlPos += vis.swirl * dt;

      const c = css / 2;
      // Keep orb well inside canvas so edges stay transparent (no square plate).
      const baseR = css * 0.26 * vis.scale;
      const def = preferReduced ? vis.energy * 0.3 : vis.energy * 0.65;

      ctx.clearRect(0, 0, css, css);
      ctx.globalCompositeOperation = "source-over";

      // Soft circular glow only — never fillRect (that paints a visible square).
      const haloR = baseR * 1.85;
      const halo = ctx.createRadialGradient(c, c, baseR * 0.15, c, c, haloR);
      halo.addColorStop(0, `hsla(${vis.hueA},90%,60%,${0.22 + vis.energy * 0.18})`);
      halo.addColorStop(0.55, `hsla(${vis.hueB},85%,55%,${0.1 + vis.energy * 0.06})`);
      halo.addColorStop(1, "hsla(0,0%,0%,0)");
      ctx.fillStyle = halo;
      ctx.beginPath();
      ctx.arc(c, c, haloR, 0, Math.PI * 2);
      ctx.fill();

      for (let L = 0; L < 3; L++) {
        const R = baseR * (1.08 - L * 0.14) * (1 + 0.04 * Math.sin(t * 0.9 + L));
        const hue = vis.hueA + (vis.hueB - vis.hueA) * (L / 2);
        ctx.beginPath();
        for (let i = 0; i <= 72; i++) {
          const th = (i / 72) * Math.PI * 2;
          const w =
            0.4 * Math.sin(3 * th + t * 1.2 + L * 2) +
            0.25 * Math.sin(5 * th - t * 1.6 + L * 4) +
            0.12 * Math.sin(7 * th + t * 2.1 + L);
          const r = R * (1 + def * w);
          const a = th + vis.swirlPos * (0.15 + L * 0.05);
          const x = c + Math.cos(a) * r;
          const y = c + Math.sin(a) * r * 0.98;
          if (i === 0) ctx.moveTo(x, y);
          else ctx.lineTo(x, y);
        }
        ctx.closePath();
        const g = ctx.createRadialGradient(c - R * 0.3, c - R * 0.3, R * 0.1, c, c, R * 1.15);
        const light = 62 + L * 6;
        g.addColorStop(0, `hsla(${hue},95%,${light}%,${0.5 - L * 0.1})`);
        g.addColorStop(1, `hsla(${hue},90%,${light - 14}%,0)`);
        ctx.fillStyle = g;
        ctx.fill();
      }

      const core = ctx.createRadialGradient(c, c, 0, c, c, baseR * 0.42);
      core.addColorStop(0, `hsla(${vis.hueA},90%,88%,0.8)`);
      core.addColorStop(0.55, `hsla(${vis.hueB},80%,70%,0.28)`);
      core.addColorStop(1, "hsla(0,0%,0%,0)");
      ctx.fillStyle = core;
      ctx.beginPath();
      ctx.arc(c, c, baseR * 0.42, 0, Math.PI * 2);
      ctx.fill();

      raf = requestAnimationFrame(draw);
    };

    raf = requestAnimationFrame(draw);
    return () => cancelAnimationFrame(raf);
  }, [size]);

  return (
    <canvas
      ref={canvasRef}
      className={className}
      width={size * 2}
      height={size * 2}
      aria-hidden
      style={{
        width: size,
        height: size,
        background: "transparent",
        display: "block",
      }}
    />
  );
}

export function statusGlow(status: VoiceSessionStatus): string {
  return MOODS[toMood(status)].glow;
}

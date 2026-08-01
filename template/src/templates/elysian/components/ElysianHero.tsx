"use client";

import { useEffect, useRef } from "react";
import type { ElysianContent } from "../lib/mapContent";
import ElysianImage from "./ElysianImage";
import PhotoAttribution from "@/components/PhotoAttribution";

export default function ElysianHero({
  hero,
  loaded,
}: {
  hero: ElysianContent["hero"];
  loaded: boolean;
}) {
  const particlesRef = useRef<HTMLDivElement>(null);
  const imgRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = particlesRef.current;
    if (!container || !loaded) return;
    container.innerHTML = "";
    const count = window.innerWidth < 700 ? 18 : 34;
    for (let i = 0; i < count; i++) {
      const p = document.createElement("span");
      p.className = "particle";
      const size = 2 + Math.random() * 3;
      p.style.width = `${size}px`;
      p.style.height = `${size}px`;
      p.style.left = `${Math.random() * 100}%`;
      p.style.animationDuration = `${8 + Math.random() * 10}s`;
      p.style.animationDelay = `${Math.random() * 10}s`;
      container.appendChild(p);
    }
  }, [loaded]);

  return (
    <section className={`hero${loaded ? " loaded" : ""}`} id="hero">
      <div className="hero-media">
        {hero.poster ? (
          <ElysianImage
            ref={imgRef}
            src={hero.poster}
            alt={hero.name}
            media={hero.posterMedia}
            fill
            className="hero-img"
            id="heroImg"
            priority
            sizes="100vw"
            onMouseMove={(e) => {
              const img = imgRef.current;
              if (!img) return;
              const { innerWidth: w, innerHeight: h } = window;
              const x = (e.clientX / w - 0.5) * 20;
              const y = (e.clientY / h - 0.5) * 20;
              img.style.transform = `scale(1.06) translate(${x}px, ${y}px)`;
            }}
          />
        ) : null}
        <div className="hero-overlay" />
        <div className="hero-grain" />
      </div>
      {hero.posterMedia?.sourceKind === "google_places_live" ? (
        <div className="absolute bottom-5 left-5 z-20 rounded bg-black/55 px-3 py-2 text-white/70">
          <PhotoAttribution media={hero.posterMedia} compact />
        </div>
      ) : null}
      <div className="hero-particles" ref={particlesRef} id="heroParticles" />
      <div className="hero-content">
        <p className="eyebrow reveal-line">
          <span>{hero.eyebrow}</span>
        </p>
        <h1 className="hero-title">
          <span className="line">
            <span className="line-inner">{hero.titleLine1}</span>
          </span>
          <span className="line gold-line">
            <span className="line-inner">{hero.titleLine2}</span>
          </span>
        </h1>
        <p className="hero-sub reveal-line">
          <span>{hero.subtitle}</span>
        </p>
        <div className="hero-actions reveal-line">
          <span>
            <a href={hero.primaryCTA.href} className="btn btn-gold btn-lg ripple">
              {hero.primaryCTA.label}
            </a>
            <a href={hero.secondaryCTA.href} className="btn btn-ghost btn-lg ripple">
              {hero.secondaryCTA.label}
            </a>
          </span>
        </div>
      </div>
      <button
        type="button"
        className="scroll-indicator"
        id="scrollIndicator"
        aria-label="Scroll to explore"
        onClick={() => document.getElementById("about")?.scrollIntoView({ behavior: "smooth" })}
      >
        <span className="scroll-mouse">
          <span className="scroll-dot" />
        </span>
        <span className="scroll-label">Scroll</span>
      </button>
    </section>
  );
}

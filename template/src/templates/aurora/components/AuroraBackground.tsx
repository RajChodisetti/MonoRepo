"use client";

import ParticleField from "./ui/ParticleField";

export default function AuroraBackground() {
  return (
    <div className="pointer-events-none fixed inset-0 -z-10 overflow-hidden bg-gradient-to-b from-[#09090B] via-[#0B1220] to-[#09090B]">
      <div className="aurora-grid-bg absolute inset-0 opacity-40" />
      <div className="aurora-blob aurora-blob-1" />
      <div className="aurora-blob aurora-blob-2" />
      <div className="aurora-blob aurora-blob-3" />
      <ParticleField />
    </div>
  );
}

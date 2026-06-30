"use client";

export default function ParticleField() {
  const particles = Array.from({ length: 40 }, (_, i) => ({
    id: i,
    left: `${(i * 17 + 7) % 100}%`,
    top: `${(i * 23 + 11) % 100}%`,
    size: 2 + (i % 3),
    delay: (i % 10) * 0.5,
    duration: 4 + (i % 6),
  }));

  return (
    <div className="pointer-events-none absolute inset-0 overflow-hidden" aria-hidden>
      {particles.map((p) => (
        <span
          key={p.id}
          className="absolute rounded-full bg-cyan-400/30"
          style={{
            left: p.left,
            top: p.top,
            width: p.size,
            height: p.size,
            animation: `particleDrift ${p.duration}s ease-in-out ${p.delay}s infinite alternate`,
          }}
        />
      ))}
      <style jsx>{`
        @keyframes particleDrift {
          from { transform: translateY(0); opacity: 0.2; }
          to { transform: translateY(-20px); opacity: 0.6; }
        }
      `}</style>
    </div>
  );
}

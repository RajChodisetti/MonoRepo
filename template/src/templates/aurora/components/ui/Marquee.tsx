"use client";

export default function Marquee({ items }: { items: string[] }) {
  const doubled = [...items, ...items];

  return (
    <div className="relative overflow-hidden border-y border-white/10 py-4">
      <div className="flex animate-marquee whitespace-nowrap">
        {doubled.map((item, i) => (
          <span
            key={`${item}-${i}`}
            className="mx-8 text-sm font-medium uppercase tracking-[0.2em] text-white/40"
          >
            {item}
          </span>
        ))}
      </div>
      <style jsx>{`
        .animate-marquee {
          animation: marquee 30s linear infinite;
        }
        @keyframes marquee {
          from { transform: translateX(0); }
          to { transform: translateX(-50%); }
        }
      `}</style>
    </div>
  );
}

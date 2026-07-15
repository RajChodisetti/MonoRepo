"use client";

export default function ElysianScrollProgress({ width }: { width: number }) {
  return (
    <div className="scroll-progress">
      <div className="scroll-progress-bar" id="scrollBar" style={{ width: `${width}%` }} />
    </div>
  );
}

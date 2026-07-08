"use client";

import { useEffect, useState } from "react";

export default function ElysianLoader({
  name,
  onDone,
}: {
  name: string;
  onDone: () => void;
}) {
  const [hidden, setHidden] = useState(false);
  const glyph = name.trim().charAt(0).toUpperCase() || "E";

  useEffect(() => {
    const t1 = setTimeout(() => {
      setHidden(true);
      onDone();
    }, 600);
    const t2 = setTimeout(() => {
      setHidden(true);
      onDone();
    }, 3000);
    return () => {
      clearTimeout(t1);
      clearTimeout(t2);
    };
  }, [onDone]);

  return (
    <div className={`loader${hidden ? " hidden" : ""}`} id="loader">
      <div className="loader-inner">
        <span className="loader-glyph">{glyph}</span>
        <div className="loader-line">
          <span />
        </div>
        <p className="loader-text">{name.toUpperCase()}</p>
      </div>
    </div>
  );
}

"use client";

import { useCallback, useEffect, useState } from "react";

export function useTestimonialCarousel(count: number) {
  const [index, setIndex] = useState(0);
  const [paused, setPaused] = useState(false);

  const goTo = useCallback(
    (i: number) => {
      if (count <= 0) return;
      setIndex(((i % count) + count) % count);
    },
    [count],
  );

  useEffect(() => {
    if (count <= 1 || paused) return;
    const timer = setInterval(() => goTo(index + 1), 5500);
    return () => clearInterval(timer);
  }, [count, paused, index, goTo]);

  return { index, goTo, setPaused };
}

"use client";

import { useEffect, useRef } from "react";

type RestaurantFeatureVideoProps = {
  poster?: string;
  src: string;
  title: string;
};

export default function RestaurantFeatureVideo({ poster, src, title }: RestaurantFeatureVideoProps) {
  const ref = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.load();
    void el.play().catch(() => {
      /* autoplay may be blocked; muted+playsInline usually ok */
    });
  }, [src]);

  return (
    <div className="relative aspect-video w-full overflow-hidden bg-black">
      <video
        key={src}
        ref={ref}
        aria-label={`${title} product demo`}
        autoPlay
        className="pointer-events-none block h-full w-full select-none object-cover"
        controls={false}
        controlsList="nodownload nofullscreen noplaybackrate"
        disablePictureInPicture
        poster={poster || undefined}
        loop
        muted
        playsInline
        preload="auto"
      >
        <source src={src} type="video/mp4" />
        Your browser does not support embedded videos.
      </video>
    </div>
  );
}

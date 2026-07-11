"use client";

import { useCallback, useEffect, useRef, useState } from "react";

type RestaurantFeatureVideoProps = {
  poster: string;
  src: string;
  title: string;
};

let activeRestaurantVideo: HTMLVideoElement | null = null;

export default function RestaurantFeatureVideo({
  poster,
  src,
  title,
}: RestaurantFeatureVideoProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const sourceAttachedRef = useRef(false);
  const isVisibleRef = useRef(false);
  const manualPlayRequestedRef = useRef(false);
  const [shouldLoad, setShouldLoad] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [autoplayBlocked, setAutoplayBlocked] = useState(false);

  const attemptPlayback = useCallback(async () => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    setIsLoading(video.readyState < HTMLMediaElement.HAVE_FUTURE_DATA);

    if (activeRestaurantVideo && activeRestaurantVideo !== video) {
      activeRestaurantVideo.pause();
    }

    activeRestaurantVideo = video;

    try {
      await video.play();
      setAutoplayBlocked(false);
    } catch {
      if (activeRestaurantVideo === video) {
        activeRestaurantVideo = null;
      }

      // Mobile browsers can reject autoplay even for muted video. Keep a clear
      // manual play action visible so the demo never looks broken.
      setAutoplayBlocked(true);
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    const container = containerRef.current;

    if (!container) {
      return;
    }

    const preloadObserver = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setShouldLoad(true);
          preloadObserver.disconnect();
        }
      },
      { rootMargin: "400px 0px" },
    );

    preloadObserver.observe(container);

    return () => preloadObserver.disconnect();
  }, []);

  useEffect(() => {
    const container = containerRef.current;

    if (!container) {
      return;
    }

    const playbackObserver = new IntersectionObserver(
      ([entry]) => {
        isVisibleRef.current = entry.isIntersecting;

        if (entry.isIntersecting && sourceAttachedRef.current) {
          void attemptPlayback();
        } else if (!entry.isIntersecting) {
          videoRef.current?.pause();
        }
      },
      // A narrow center band makes the focused demo switch cleanly in either
      // scroll direction and prevents two decoders from competing.
      { rootMargin: "-46% 0px -46% 0px", threshold: 0.01 },
    );

    playbackObserver.observe(container);

    return () => playbackObserver.disconnect();
  }, [attemptPlayback]);

  useEffect(() => {
    const video = videoRef.current;

    return () => {
      if (activeRestaurantVideo === video) {
        activeRestaurantVideo = null;
      }
    };
  }, []);

  useEffect(() => {
    const video = videoRef.current;

    if (!shouldLoad || !video) {
      return;
    }

    sourceAttachedRef.current = true;
    video.load();

    if (isVisibleRef.current || manualPlayRequestedRef.current) {
      void attemptPlayback();
    }
  }, [attemptPlayback, shouldLoad]);

  const handlePlay = () => {
    manualPlayRequestedRef.current = true;
    setIsLoading(true);

    if (!shouldLoad) {
      setShouldLoad(true);
      return;
    }

    void attemptPlayback();
  };

  return (
    <div ref={containerRef} className="relative">
      <video
        ref={videoRef}
        aria-label={`${title} product demo`}
        className="aspect-video w-full bg-zinc-900 object-cover"
        poster={poster}
        loop
        muted
        playsInline
        preload={shouldLoad ? "auto" : "none"}
        onCanPlay={() => setIsLoading(false)}
        onPause={() => {
          if (activeRestaurantVideo === videoRef.current) {
            activeRestaurantVideo = null;
          }

          setIsPlaying(false);
        }}
        onPlay={() => {
          manualPlayRequestedRef.current = false;
          setIsLoading(false);
          setIsPlaying(true);
        }}
        onWaiting={() => setIsLoading(true)}
      >
        {shouldLoad ? <source src={src} type="video/mp4" /> : null}
        Your browser does not support embedded videos.
      </video>

      {!isPlaying ? (
        <button
          type="button"
          aria-label={`Play ${title} demo`}
          className="absolute inset-0 flex items-center justify-center bg-ink/10 transition hover:bg-ink/20 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-4px] focus-visible:outline-white"
          onClick={handlePlay}
        >
          <span className="flex items-center gap-3 rounded-full bg-ink/90 px-5 py-3 text-sm font-semibold text-white shadow-lg backdrop-blur">
            <svg
              aria-hidden="true"
              className="h-4 w-4"
              viewBox="0 0 24 24"
              fill="currentColor"
            >
              <path d="M8 5.14v13.72a1 1 0 0 0 1.52.85l10.28-6.86a1 1 0 0 0 0-1.66L9.52 4.29A1 1 0 0 0 8 5.14Z" />
            </svg>
            {isLoading
              ? "Loading preview…"
              : autoplayBlocked
                ? "Tap to play demo"
                : "Play demo"}
          </span>
        </button>
      ) : null}
    </div>
  );
}

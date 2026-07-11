"use client";

import { useCallback, useEffect, useRef, useState } from "react";

type RestaurantFeatureVideoProps = {
  poster: string;
  src: string;
  title: string;
};

export default function RestaurantFeatureVideo({
  poster,
  src,
  title,
}: RestaurantFeatureVideoProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [autoplayBlocked, setAutoplayBlocked] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);

  const playVideo = useCallback(() => {
    const video = videoRef.current;

    if (!video) {
      return;
    }

    video.muted = true;
    const playPromise = video.play();

    void playPromise
      .then(() => {
        setAutoplayBlocked(false);
        setLoadFailed(false);
      })
      .catch(() => setAutoplayBlocked(true));
  }, []);

  useEffect(() => {
    playVideo();

    const resumePlayback = () => {
      if (!document.hidden && videoRef.current?.paused) {
        playVideo();
      }
    };

    document.addEventListener("visibilitychange", resumePlayback);

    return () => document.removeEventListener("visibilitychange", resumePlayback);
  }, [playVideo, src]);

  return (
    <div className="relative">
      <video
        ref={videoRef}
        aria-label={`${title} product demo`}
        className="aspect-video w-full bg-zinc-900 object-cover"
        src={src}
        poster={poster}
        autoPlay
        loop
        muted
        playsInline
        preload="auto"
        onCanPlay={playVideo}
        onError={() => setLoadFailed(true)}
        onPlaying={() => {
          setAutoplayBlocked(false);
          setLoadFailed(false);
        }}
      >
        Your browser does not support embedded videos.
      </video>

      {autoplayBlocked ? (
        <button
          type="button"
          aria-label={`Start ${title} video`}
          className="absolute inset-0 flex items-center justify-center bg-ink/10 transition hover:bg-ink/20 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-4px] focus-visible:outline-white"
          onClick={playVideo}
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
            Tap to start video
          </span>
        </button>
      ) : null}

      {loadFailed ? (
        <div
          role="alert"
          className="absolute inset-0 flex items-center justify-center bg-ink/20"
        >
          <span className="rounded-full bg-ink/90 px-5 py-3 text-sm font-semibold text-white shadow-lg backdrop-blur">
            Video unavailable
          </span>
        </div>
      ) : null}
    </div>
  );
}

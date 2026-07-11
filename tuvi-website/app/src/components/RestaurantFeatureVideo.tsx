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
  const objectUrlRef = useRef<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isPlaying, setIsPlaying] = useState(false);
  const [autoplayBlocked, setAutoplayBlocked] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);

  const playVideo = useCallback(() => {
    const video = videoRef.current;

    if (!video || !objectUrlRef.current) {
      return;
    }

    // Keep play() directly in the click call stack so mobile browsers can use
    // the user gesture when autoplay policy requires it.
    const playPromise = video.play();

    void playPromise
      .then(() => {
        setAutoplayBlocked(false);
        setLoadFailed(false);
      })
      .catch(() => setAutoplayBlocked(true))
      .finally(() => setIsLoading(false));
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    const video = videoRef.current;
    let disposed = false;

    setIsLoading(true);
    setIsPlaying(false);
    setAutoplayBlocked(false);
    setLoadFailed(false);

    void (async () => {
      try {
        const response = await fetch(src, {
          cache: "force-cache",
          signal: controller.signal,
        });

        if (!response.ok) {
          throw new Error(`Video request failed with status ${response.status}`);
        }

        // Materialize the complete MP4 before playback. The retained Blob URL
        // then loops from browser memory instead of streaming from the VM.
        const blob = await response.blob();

        if (disposed || !video) {
          return;
        }

        const objectUrl = URL.createObjectURL(blob);
        objectUrlRef.current = objectUrl;
        video.src = objectUrl;
        video.muted = true;
        video.load();
        playVideo();
      } catch {
        if (!disposed && !controller.signal.aborted) {
          setLoadFailed(true);
          setIsLoading(false);
        }
      }
    })();

    return () => {
      disposed = true;
      controller.abort();

      if (video) {
        video.pause();
        video.removeAttribute("src");
        video.load();
      }

      if (objectUrlRef.current) {
        URL.revokeObjectURL(objectUrlRef.current);
        objectUrlRef.current = null;
      }
    };
  }, [playVideo, src]);

  useEffect(() => {
    const resumePlayback = () => {
      if (!document.hidden && videoRef.current?.paused) {
        playVideo();
      }
    };

    document.addEventListener("visibilitychange", resumePlayback);

    return () => document.removeEventListener("visibilitychange", resumePlayback);
  }, [playVideo]);

  return (
    <div className="relative">
      <video
        ref={videoRef}
        aria-label={`${title} product demo`}
        className="aspect-video w-full bg-zinc-900 object-cover"
        poster={poster}
        autoPlay
        loop
        muted
        playsInline
        preload="auto"
        onPause={() => setIsPlaying(false)}
        onPlaying={() => {
          setAutoplayBlocked(false);
          setIsLoading(false);
          setIsPlaying(true);
        }}
        onWaiting={() => setIsLoading(true)}
      >
        Your browser does not support embedded videos.
      </video>

      {!isPlaying && isLoading ? (
        <div
          role="status"
          className="absolute inset-0 flex items-center justify-center bg-ink/10"
        >
          <span className="rounded-full bg-ink/90 px-5 py-3 text-sm font-semibold text-white shadow-lg backdrop-blur">
            Loading video…
          </span>
        </div>
      ) : null}

      {!isPlaying && autoplayBlocked ? (
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

      {!isPlaying && loadFailed ? (
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

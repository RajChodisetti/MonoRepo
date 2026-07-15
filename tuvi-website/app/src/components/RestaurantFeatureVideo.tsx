"use client";

import { useCallback, useEffect, useRef, useState } from "react";

type RestaurantFeatureVideoProps = {
  poster: string;
  src: string;
  title: string;
};

export default function RestaurantFeatureVideo({ poster, src, title }: RestaurantFeatureVideoProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [playing, setPlaying] = useState(false);
  const [autoplayBlocked, setAutoplayBlocked] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);
  const [userPaused, setUserPaused] = useState(false);

  const playVideo = useCallback(() => {
    const video = videoRef.current;
    if (!video) return;

    video.muted = true;
    void video
      .play()
      .then(() => {
        setPlaying(true);
        setAutoplayBlocked(false);
        setLoadFailed(false);
      })
      .catch(() => setAutoplayBlocked(true));
  }, []);

  const pauseVideo = useCallback(() => {
    videoRef.current?.pause();
    setPlaying(false);
  }, []);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reducedMotion) {
      setUserPaused(true);
      return;
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !userPaused) playVideo();
        else pauseVideo();
      },
      { threshold: 0.45 },
    );

    observer.observe(video);
    return () => observer.disconnect();
  }, [pauseVideo, playVideo, src, userPaused]);

  const togglePlayback = () => {
    if (playing) {
      setUserPaused(true);
      pauseVideo();
    } else {
      setUserPaused(false);
      playVideo();
    }
  };

  return (
    <div className="relative">
      <video
        ref={videoRef}
        aria-label={`${title} product demo`}
        className="aspect-video w-full bg-ink object-cover"
        src={src}
        poster={poster}
        loop
        muted
        playsInline
        preload="metadata"
        onError={() => setLoadFailed(true)}
        onPause={() => setPlaying(false)}
        onPlaying={() => {
          setPlaying(true);
          setAutoplayBlocked(false);
          setLoadFailed(false);
        }}
      >
        Your browser does not support embedded videos.
      </video>

      {!loadFailed ? (
        <button
          type="button"
          aria-label={`${playing ? "Pause" : "Play"} ${title} video`}
          className="absolute bottom-4 right-4 inline-flex cursor-pointer items-center gap-2 rounded-full bg-ink/90 px-4 py-2 text-xs font-semibold text-[#fffef8] shadow-lg backdrop-blur transition-colors hover:bg-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
          onClick={togglePlayback}
        >
          <svg aria-hidden="true" className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor">
            {playing ? <path d="M7 5h4v14H7zM13 5h4v14h-4z" /> : <path d="M8 5.14v13.72a1 1 0 0 0 1.52.85l10.28-6.86a1 1 0 0 0 0-1.66L9.52 4.29A1 1 0 0 0 8 5.14Z" />}
          </svg>
          {autoplayBlocked && !playing ? "Play demo" : playing ? "Pause" : "Play"}
        </button>
      ) : null}

      {loadFailed ? (
        <div role="alert" className="absolute inset-0 flex items-center justify-center bg-ink/25">
          <span className="rounded-full bg-ink/90 px-5 py-3 text-sm font-semibold text-[#fffef8] shadow-lg backdrop-blur">
            Video unavailable
          </span>
        </div>
      ) : null}
    </div>
  );
}

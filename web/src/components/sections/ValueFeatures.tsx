"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import AppDownloadsPanel from "@/components/sections/features/AppDownloadsPanel";
import GoogleTrafficPanel from "@/components/sections/features/GoogleTrafficPanel";
import OnlineSalesPanel from "@/components/sections/features/OnlineSalesPanel";
import RepeatOrdersPanel from "@/components/sections/features/RepeatOrdersPanel";

const tabs = [
  { label: "More Google Traffic" },
  { label: "More Online Sales" },
  { label: "More Repeat Orders" },
  { label: "More App Downloads" },
] as const;

export default function ValueFeatures() {
  const [active, setActive] = useState(0);
  const [progress, setProgress] = useState(0);
  const [displayProgress, setDisplayProgress] = useState(0);
  const advancingRef = useRef(false);
  const targetRef = useRef(0);
  const displayRef = useRef(0);
  const rafRef = useRef(0);

  // Water-smooth lerp toward target progress
  useEffect(() => {
    const tick = () => {
      const curr = displayRef.current;
      const target = targetRef.current;
      const next = curr + (target - curr) * 0.12;
      if (Math.abs(target - next) < 0.001) {
        displayRef.current = target;
        setDisplayProgress(target);
      } else {
        displayRef.current = next;
        setDisplayProgress(next);
        rafRef.current = requestAnimationFrame(tick);
      }
    };
    cancelAnimationFrame(rafRef.current);
    rafRef.current = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(rafRef.current);
  }, [progress]);

  const handleProgress = useCallback((value: number) => {
    const clamped = Math.min(1, Math.max(0, value));
    targetRef.current = clamped;
    setProgress(clamped);
  }, []);

  const goNext = useCallback(() => {
    if (advancingRef.current) return;
    advancingRef.current = true;
    window.setTimeout(() => {
      setActive((a) => (a + 1) % tabs.length);
      targetRef.current = 0;
      displayRef.current = 0;
      setProgress(0);
      setDisplayProgress(0);
      advancingRef.current = false;
    }, 450);
  }, []);

  const selectTab = (index: number) => {
    advancingRef.current = false;
    setActive(index);
    targetRef.current = 0;
    displayRef.current = 0;
    setProgress(0);
    setDisplayProgress(0);
  };

  return (
    <section className="bg-bg px-4 pb-14 pt-6 sm:px-8 sm:pb-16 sm:pt-8 md:px-12">
      <div className="mx-auto max-w-[1100px]">
        <div
          className="grid grid-cols-2 gap-x-4 gap-y-4 sm:grid-cols-4 sm:gap-x-6"
          role="tablist"
          aria-label="Value pillars"
        >
          {tabs.map((tab, index) => {
            const isActive = active === index;
            return (
              <button
                key={tab.label}
                type="button"
                role="tab"
                aria-selected={isActive}
                onClick={() => selectTab(index)}
                className="group flex cursor-pointer flex-col items-stretch gap-3 text-left"
              >
                <span
                  className={`text-[15px] leading-snug tracking-[-0.01em] transition-colors sm:text-[16px] md:text-[17px] ${
                    isActive
                      ? "font-semibold text-[#1a1a1a]"
                      : "font-medium text-[#9a9a9a] group-hover:text-[#6b6b6b]"
                  }`}
                >
                  {tab.label}
                </span>

                <span
                  className="relative block h-[2px] w-full overflow-hidden rounded-full bg-[#efefef]"
                  aria-hidden="true"
                >
                  {isActive ? (
                    <span
                      className="absolute inset-y-0 left-0 rounded-full"
                      style={{
                        width: `${displayProgress * 100}%`,
                        background:
                          "linear-gradient(90deg, #d4d4d4 0%, #c0c0c0 45%, #b8b8b8 100%)",
                      }}
                    />
                  ) : null}
                </span>
              </button>
            );
          })}
        </div>

        <div className="mt-8 sm:mt-10" role="tabpanel">
          {active === 0 ? (
            <GoogleTrafficPanel
              key="google-traffic"
              onProgress={handleProgress}
              onComplete={goNext}
            />
          ) : active === 1 ? (
            <OnlineSalesPanel
              key="online-sales"
              onProgress={handleProgress}
              onComplete={goNext}
            />
          ) : active === 2 ? (
            <RepeatOrdersPanel
              key="repeat-orders"
              onProgress={handleProgress}
              onComplete={goNext}
            />
          ) : (
            <AppDownloadsPanel
              key="app-downloads"
              onProgress={handleProgress}
              onComplete={goNext}
            />
          )}
        </div>
      </div>
    </section>
  );
}

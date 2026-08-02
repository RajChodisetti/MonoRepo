"use client";

import { useEffect, useRef, useState } from "react";
import CompetitorsCard from "@/components/sections/report/CompetitorsCard";
import HealthCard from "@/components/sections/report/HealthCard";
import IssuesCard from "@/components/sections/report/IssuesCard";
import ReportSplash from "@/components/sections/report/ReportSplash";

const SPLASH_MS = 1400;
const FADE_MS = 400;
const HOLD_MS = 1800;
const SCROLL_DOWN_MS = 4000;
const ISSUES_HOLD_MS = 2800;
const SCROLL_UP_MS = 1400;

function easeInOutQuint(t: number) {
  return t < 0.5 ? 16 * t * t * t * t * t : 1 - Math.pow(-2 * t + 2, 5) / 2;
}

export default function ReportDemoScreen() {
  const contentRef = useRef<HTMLDivElement>(null);
  const issuesRef = useRef<HTMLDivElement>(null);
  const offsetRef = useRef(0);
  const [splashVisible, setSplashVisible] = useState(true);
  const [splashFading, setSplashFading] = useState(false);
  const [offsetY, setOffsetY] = useState(0);

  useEffect(() => {
    let cancelled = false;
    const timers: number[] = [];
    let raf = 0;

    const setOffset = (value: number) => {
      offsetRef.current = value;
      setOffsetY(value);
    };

    const sleep = (ms: number) =>
      new Promise<void>((resolve) => {
        timers.push(window.setTimeout(resolve, ms));
      });

    const animateTo = (to: number, duration: number) =>
      new Promise<void>((resolve) => {
        const from = offsetRef.current;
        const start = performance.now();

        const tick = (now: number) => {
          if (cancelled) {
            resolve();
            return;
          }
          const t = Math.min(1, (now - start) / duration);
          setOffset(from + (to - from) * easeInOutQuint(t));
          if (t < 1) {
            raf = requestAnimationFrame(tick);
          } else {
            setOffset(to);
            resolve();
          }
        };

        raf = requestAnimationFrame(tick);
      });

    const run = async () => {
      while (!cancelled) {
        setSplashVisible(true);
        setSplashFading(false);
        setOffset(0);

        await sleep(SPLASH_MS);
        if (cancelled) break;

        setSplashFading(true);
        await sleep(FADE_MS);
        if (cancelled) break;

        setSplashVisible(false);
        setSplashFading(false);

        await sleep(HOLD_MS);
        if (cancelled) break;

        const content = contentRef.current;
        const issues = issuesRef.current;
        const viewport = content?.parentElement;
        if (content && issues && viewport) {
          const maxScroll = Math.max(0, content.scrollHeight - viewport.clientHeight);
          const target = -Math.min(maxScroll, Math.max(0, issues.offsetTop - 12));
          await animateTo(target, SCROLL_DOWN_MS);
        }
        if (cancelled) break;

        await sleep(ISSUES_HOLD_MS);
        if (cancelled) break;

        await animateTo(0, SCROLL_UP_MS);
      }
    };

    void run();

    return () => {
      cancelled = true;
      timers.forEach((id) => window.clearTimeout(id));
      cancelAnimationFrame(raf);
    };
  }, []);

  return (
    <div className="relative h-full overflow-hidden bg-white">
      <ReportSplash visible={splashVisible} fading={splashFading} />

      <div className="relative h-full overflow-hidden bg-white">
        <div
          ref={contentRef}
          className="pointer-events-none px-3.5 pb-10 pt-1 will-change-transform"
          style={{ transform: `translate3d(0, ${offsetY}px, 0)` }}
        >
          <div className="flex flex-col gap-3">
            <HealthCard />
            <CompetitorsCard />
            <div ref={issuesRef}>
              <IssuesCard />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

"use client";

import { useEffect, useRef } from "react";
import { useRouter } from "next/navigation";

const HEARTBEAT_MS = 30_000;

export function IdleGuard() {
  const router = useRouter();
  const lastSent = useRef(0);

  useEffect(() => {
    let cancelled = false;

    const send = async () => {
      const now = Date.now();
      if (now - lastSent.current < 5_000) return;
      lastSent.current = now;
      try {
        const res = await fetch("/api/auth/heartbeat", { method: "POST" });
        if (!res.ok && !cancelled) {
          router.replace("/login?reason=idle");
        }
      } catch {
        /* ignore transient network errors */
      }
    };

    const onActivity = () => {
      void send();
    };

    window.addEventListener("mousemove", onActivity, { passive: true });
    window.addEventListener("keydown", onActivity);
    window.addEventListener("click", onActivity);
    window.addEventListener("scroll", onActivity, { passive: true });
    const timer = window.setInterval(() => void send(), HEARTBEAT_MS);
    void send();

    return () => {
      cancelled = true;
      window.clearInterval(timer);
      window.removeEventListener("mousemove", onActivity);
      window.removeEventListener("keydown", onActivity);
      window.removeEventListener("click", onActivity);
      window.removeEventListener("scroll", onActivity);
    };
  }, [router]);

  return null;
}

"use client";

import { useEffect, useRef } from "react";

type SessionCapability = {
  session_id: string;
  session_token: string;
};

type VoiceTranscriptEvent = CustomEvent<{
  role?: "user" | "assistant" | "system";
  text?: string;
}>;

export default function DemoEngagementTracker({
  slug,
  demoToken,
  templateID,
}: {
  slug?: string;
  demoToken?: string;
  templateID: "1" | "2" | "3" | "4";
}) {
  const sessionRef = useRef<SessionCapability | null>(null);
  const activeSecondsRef = useRef(0);
  const activeStartedAtRef = useRef<number | null>(null);

  useEffect(() => {
    const apiBase = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "";
    if (!apiBase) return;

    const fragment = new URLSearchParams(window.location.hash.replace(/^#/, ""));
    const issuedSession = fragment.get("engagement_session");
    const issuedToken = fragment.get("engagement_token");
    const hasIssuedCapability = Boolean(issuedSession && issuedToken);
    if (!hasIssuedCapability && (!slug || !demoToken)) return;

    let cancelled = false;
    activeSecondsRef.current = 0;
    activeStartedAtRef.current = null;

    const beginActive = () => {
      if (document.visibilityState === "visible" && activeStartedAtRef.current === null) {
        activeStartedAtRef.current = performance.now();
      }
    };
    const freezeActive = () => {
      if (activeStartedAtRef.current === null) return;
      activeSecondsRef.current += Math.max(
        0,
        (performance.now() - activeStartedAtRef.current) / 1000,
      );
      activeStartedAtRef.current = null;
    };
    const activeSeconds = () => {
      const inProgress =
        activeStartedAtRef.current === null
          ? 0
          : Math.max(0, (performance.now() - activeStartedAtRef.current) / 1000);
      return Math.min(86400, Math.floor(activeSecondsRef.current + inProgress));
    };
    const post = async (path: string, body: Record<string, unknown>, keepalive = false) => {
      const response = await fetch(`${apiBase}${path}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
        cache: "no-store",
        keepalive,
      });
      if (!response.ok) throw new Error(`engagement request failed (${response.status})`);
      return response;
    };

    const recordEvent = (event: "heartbeat" | "end", keepalive = false) => {
      const session = sessionRef.current;
      if (!session) return Promise.resolve();
      return post(
        `/api/public/v1/demo-sessions/${session.session_id}/events`,
        {
          session_token: session.session_token,
          event,
          active_seconds: activeSeconds(),
        },
        keepalive,
      ).then(() => undefined);
    };

    const onTranscript = (event: Event) => {
      const session = sessionRef.current;
      const detail = (event as VoiceTranscriptEvent).detail;
      if (!session || !detail?.role || !detail.text?.trim()) return;
      void post(`/api/public/v1/demo-sessions/${session.session_id}/transcript`, {
        session_token: session.session_token,
        role: detail.role,
        content: detail.text.trim().slice(0, 4000),
      }).catch(() => undefined);
    };

    const onPageHide = () => {
      freezeActive();
      void recordEvent("end", true).catch(() => undefined);
    };
    const onVisibility = () => {
      if (document.visibilityState === "hidden") {
        freezeActive();
        void recordEvent("heartbeat", true).catch(() => undefined);
      } else if (sessionRef.current) {
        beginActive();
      }
    };

    window.addEventListener("tuvi:voice-transcript", onTranscript);
    window.addEventListener("pagehide", onPageHide);
    document.addEventListener("visibilitychange", onVisibility);

    if (hasIssuedCapability) {
      sessionRef.current = {
        session_id: issuedSession as string,
        session_token: issuedToken as string,
      };
      beginActive();
      window.history.replaceState(null, "", `${window.location.pathname}${window.location.search}`);
    } else {
      void post(`/api/public/v1/demo/${encodeURIComponent(slug as string)}/sessions`, {
        demo_token: demoToken,
        template_id: templateID,
      })
        .then(async (response) => {
          const capability = (await response.json()) as SessionCapability;
          if (cancelled || !capability.session_id || !capability.session_token) return;
          sessionRef.current = capability;
          beginActive();
        })
        .catch(() => undefined);
    }

    const heartbeat = window.setInterval(() => {
      if (document.visibilityState === "visible") {
        void recordEvent("heartbeat").catch(() => undefined);
      }
    }, 15_000);

    return () => {
      cancelled = true;
      window.clearInterval(heartbeat);
      window.removeEventListener("tuvi:voice-transcript", onTranscript);
      window.removeEventListener("pagehide", onPageHide);
      document.removeEventListener("visibilitychange", onVisibility);
      freezeActive();
      void recordEvent("end", true).catch(() => undefined);
      sessionRef.current = null;
      activeStartedAtRef.current = null;
    };
  }, [slug, demoToken, templateID]);

  return null;
}

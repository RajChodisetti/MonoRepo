import {
  getVoiceAgentBaseUrl,
  type RealEstateLanguage,
  type VoiceAgentKind,
} from "@/lib/voiceAgentConfig";

/** Simple in-memory IP throttle for outbound dial proxy. */
const hits = new Map<string, { count: number; resetAt: number }>();
const WINDOW_MS = 60_000;
const MAX_PER_WINDOW = 5;

function clientIp(req: Request): string {
  const forwarded = req.headers.get("x-forwarded-for");
  if (forwarded) return forwarded.split(",")[0]?.trim() || "unknown";
  return req.headers.get("x-real-ip") || "unknown";
}

function rateLimited(ip: string): boolean {
  const now = Date.now();
  const entry = hits.get(ip);
  if (!entry || now > entry.resetAt) {
    hits.set(ip, { count: 1, resetAt: now + WINDOW_MS });
    return false;
  }
  entry.count += 1;
  return entry.count > MAX_PER_WINDOW;
}

function looksLikePhone(raw: string): boolean {
  const digits = raw.replace(/\D/g, "");
  return digits.length >= 8 && digits.length <= 15;
}

function parseAgent(raw: unknown): VoiceAgentKind {
  return raw === "real_estate" ? "real_estate" : "corporate";
}

function parseLanguage(raw: unknown): RealEstateLanguage {
  const v = String(raw || "en").trim().toLowerCase();
  if (v === "hi" || v === "te" || v === "auto" || v === "en") return v;
  return "en";
}

export async function POST(req: Request) {
  const ip = clientIp(req);
  if (rateLimited(ip)) {
    return Response.json(
      { status: "error", message: "Too many callback requests. Try again in a minute." },
      { status: 429 },
    );
  }

  let body: {
    phone?: string;
    name?: string;
    agent?: string;
    language?: string;
  };
  try {
    body = await req.json();
  } catch {
    return Response.json({ status: "error", message: "Invalid JSON body." }, { status: 400 });
  }

  const phone = (body.phone || "").trim();
  if (!phone || !looksLikePhone(phone)) {
    return Response.json(
      {
        status: "error",
        message: "Enter a valid phone number with country code (e.g. +61 412 345 678).",
      },
      { status: 400 },
    );
  }

  const kind = parseAgent(body.agent);
  const language = parseLanguage(body.language);
  const base = getVoiceAgentBaseUrl(kind);
  const secret =
    (kind === "real_estate"
      ? process.env.REAL_ESTATE_CALL_API_SECRET || process.env.CALL_API_SECRET
      : process.env.CALL_API_SECRET) || "";
  const isDev = process.env.NODE_ENV !== "production";

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (secret) {
    headers["X-Call-Api-Key"] = secret;
  }

  const payload =
    kind === "real_estate"
      ? {
          to: phone,
          agent: "andre",
          language,
          campaign_id: "website_form_real_estate",
          skip_compliance: isDev,
        }
      : {
          to: phone,
          agent: "corporate",
          campaign_id: "website_form",
          skip_compliance: isDev,
        };

  try {
    const res = await fetch(`${base}/call`, {
      method: "POST",
      headers,
      body: JSON.stringify(payload),
      cache: "no-store",
      signal: AbortSignal.timeout(20_000),
    });

    const data = (await res.json().catch(() => ({}))) as {
      status?: string;
      reason?: string;
      message?: string;
      call_sid?: string;
      to?: string;
    };

    if (data.status === "calling") {
      return Response.json({
        status: "calling",
        message: "We're calling you now — please answer your phone.",
        to: data.to,
      });
    }

    if (data.status === "queued") {
      return Response.json(
        {
          status: "queued",
          message:
            data.message ||
            "We're outside our calling hours right now. Book a consultation or call us during business hours.",
          reason: data.reason,
        },
        { status: 200 },
      );
    }

    if (data.status === "blocked") {
      const reason = data.reason || "blocked";
      const message =
        reason === "internal_opt_out"
          ? "That number is on our do-not-call list."
          : reason === "invalid_number"
            ? "That phone number doesn't look valid. Use a country code (e.g. +61…)."
            : data.message || "We couldn't place that call.";
      return Response.json({ status: "blocked", reason, message }, { status: 400 });
    }

    if (!res.ok) {
      return Response.json(
        {
          status: "error",
          message: data.message || "Voice agent could not place the call.",
        },
        { status: res.status >= 400 ? res.status : 502 },
      );
    }

    return Response.json({
      status: data.status || "error",
      message: data.message || "Unexpected response from voice agent.",
    });
  } catch {
    return Response.json(
      {
        status: "unavailable",
        message:
          kind === "real_estate"
            ? "Real estate agent is not reachable. Start andre-voice-agent on port 8001."
            : "Voice agent is not reachable. Start voice-sales-agent and ensure PUBLIC_BASE_URL is set for Twilio.",
      },
      { status: 503 },
    );
  }
}

export type VoiceAgentKind = "corporate" | "real_estate";

export type RealEstateLanguage = "en" | "hi" | "te" | "auto";

export function getVoiceAgentBaseUrl(kind: VoiceAgentKind = "corporate"): string {
  if (kind === "real_estate") {
    return (
      process.env.REAL_ESTATE_VOICE_AGENT_URL?.replace(/\/$/, "") ||
      process.env.NEXT_PUBLIC_REAL_ESTATE_VOICE_AGENT_URL?.replace(/\/$/, "") ||
      "http://localhost:8001"
    );
  }
  return (
    process.env.VOICE_AGENT_URL?.replace(/\/$/, "") ||
    process.env.NEXT_PUBLIC_VOICE_AGENT_URL?.replace(/\/$/, "") ||
    "http://localhost:8000"
  );
}

export function getVoiceAgentWsUrl(
  kind: VoiceAgentKind = "corporate",
  language: RealEstateLanguage = "en",
): string {
  const base = getVoiceAgentBaseUrl(kind);
  const wsBase = base.replace(/^http/i, "ws");
  if (kind === "real_estate") {
    const lang = encodeURIComponent(language || "en");
    return `${wsBase}/browser-stream?language=${lang}`;
  }
  return `${wsBase}/browser-stream?agent=corporate`;
}

export function getDefaultRealEstateLanguage(): RealEstateLanguage {
  const raw = (process.env.NEXT_PUBLIC_REAL_ESTATE_DEFAULT_LANGUAGE || "en").trim().toLowerCase();
  if (raw === "hi" || raw === "te" || raw === "auto" || raw === "en") return raw;
  return "en";
}

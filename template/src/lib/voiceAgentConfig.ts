/** Base URL of the voice-sales-agent FastAPI server (no trailing slash). */
export function getVoiceAgentBaseUrl(): string {
  return (
    process.env.NEXT_PUBLIC_VOICE_AGENT_URL?.replace(/\/$/, "") ||
    "http://localhost:8000"
  );
}

export function getVoiceAgentWsUrl(restaurantIndex?: number): string {
  const base = getVoiceAgentBaseUrl();
  const wsBase = base.replace(/^http/i, "ws");
  const idx = restaurantIndex ?? 0;
  return `${wsBase}/browser-stream?restaurant_index=${idx}`;
}

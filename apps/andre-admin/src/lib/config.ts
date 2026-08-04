export const SESSION_COOKIE = "andre_admin_session";

export function idleTimeoutMs(): number {
  const raw = Number(process.env.IDLE_TIMEOUT_MS || 600_000);
  return Number.isFinite(raw) && raw > 0 ? raw : 600_000;
}

export function sessionSecret(): string {
  const secret = (process.env.SESSION_SECRET || "").trim();
  if (secret.length >= 16) return secret;
  return "local-andre-admin-session-secret-32chars";
}

export function adminCredentials(): { email: string; password: string } {
  return {
    email: (process.env.ANDRE_ADMIN_EMAIL || "admin@tuvi.local").trim().toLowerCase(),
    password: process.env.ANDRE_ADMIN_PASSWORD || "andre-admin-123",
  };
}

export function agentBaseUrl(): string {
  return (process.env.ANDRE_AGENT_URL || "http://127.0.0.1:8001").replace(/\/$/, "");
}

export function callApiSecret(): string {
  return (process.env.CALL_API_SECRET || "").trim();
}

export function getConsultationApiUrl(): string {
  return (
    process.env.CONSULTATION_API_URL?.replace(/\/$/, "") ||
    process.env.MONOREPO_API_URL?.replace(/\/$/, "") ||
    "http://localhost:8080"
  );
}

export function getConsultationApiToken(): string {
  return process.env.TUVI_API_TOKEN || "local-dev-tuvi-api-token-change-me";
}

export async function consultationFetch(path: string, init?: RequestInit) {
  const base = getConsultationApiUrl();
  const token = getConsultationApiToken();
  const headers = new Headers(init?.headers);
  headers.set("Authorization", `Bearer ${token}`);
  if (init?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  return fetch(`${base}${path}`, { ...init, headers });
}

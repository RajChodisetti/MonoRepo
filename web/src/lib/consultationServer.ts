export function getConsultationApiUrl(): string {
  return (
    process.env.CONSULTATION_API_URL?.replace(/\/$/, "") ||
    process.env.MONOREPO_API_URL?.replace(/\/$/, "") ||
    "http://localhost:8080"
  );
}

function getConsultationApiToken(): string {
  const token = process.env.TUVI_API_TOKEN?.trim();
  if (token) return token;
  if (process.env.NODE_ENV === "production") {
    throw new Error("TUVI_API_TOKEN is required in production");
  }
  return "local-dev-tuvi-api-token-change-me";
}

export async function consultationFetch(path: string, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers);
  headers.set("Authorization", `Bearer ${getConsultationApiToken()}`);
  if (init?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  return fetch(`${getConsultationApiUrl()}${path}`, {
    ...init,
    headers,
    cache: "no-store",
  });
}

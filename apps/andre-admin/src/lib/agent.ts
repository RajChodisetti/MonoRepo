import { agentBaseUrl, callApiSecret } from "@/lib/config";

type AgentFetchOptions = {
  method?: string;
  body?: unknown;
  auth?: boolean;
};

export async function agentFetch<T = unknown>(
  path: string,
  options: AgentFetchOptions = {},
): Promise<{ ok: boolean; status: number; data: T; error?: string }> {
  const url = `${agentBaseUrl()}${path.startsWith("/") ? path : `/${path}`}`;
  const headers: Record<string, string> = {
    Accept: "application/json",
  };
  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  if (options.auth !== false) {
    const secret = callApiSecret();
    if (secret) headers["X-Call-Api-Key"] = secret;
  }

  try {
    const res = await fetch(url, {
      method: options.method || "GET",
      headers,
      body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
      cache: "no-store",
    });
    const text = await res.text();
    let data: T;
    try {
      data = text ? (JSON.parse(text) as T) : ({} as T);
    } catch {
      data = { message: text } as T;
    }
    if (!res.ok) {
      const errObj = data as { detail?: string; message?: string; error?: string };
      return {
        ok: false,
        status: res.status,
        data,
        error: errObj.detail || errObj.message || errObj.error || `Agent error ${res.status}`,
      };
    }
    return { ok: true, status: res.status, data };
  } catch (exc) {
    return {
      ok: false,
      status: 502,
      data: {} as T,
      error: exc instanceof Error ? exc.message : "Agent unreachable",
    };
  }
}

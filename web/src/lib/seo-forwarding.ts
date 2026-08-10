const SAFE_FORWARDED_VALUE = /^[0-9a-fA-F:. ,]+$/;
const MAX_FORWARDED_VALUE_LENGTH = 512;

function safeForwardedValue(value: string | null): string {
  const normalized = value?.trim() || "";
  if (
    !normalized ||
    normalized.length > MAX_FORWARDED_VALUE_LENGTH ||
    !SAFE_FORWARDED_VALUE.test(normalized)
  ) {
    return "";
  }
  return normalized;
}

/**
 * Preserve the client chain Caddy supplied to the public Next route when the
 * BFF calls Go over the private Docker network. Go trusts this header only from
 * a loopback/private immediate peer and selects its rightmost valid address.
 */
export function seoForwardingHeaders(request: Request): Record<string, string> {
  const headers: Record<string, string> = {};
  const forwardedFor = safeForwardedValue(request.headers.get("x-forwarded-for"));
  const realIP = safeForwardedValue(request.headers.get("x-real-ip"));
  if (forwardedFor) headers["X-Forwarded-For"] = forwardedFor;
  if (realIP) headers["X-Real-IP"] = realIP;
  return headers;
}

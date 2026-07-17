export const CITIES = [
  "Adelaide",
  "Brisbane",
  "Melbourne",
  "Perth",
  "Sydney",
] as const;

export const NICHES = ["restaurant", "dentist", "plumber"] as const;

export const RESTAURANT_STATUSES = [
  "lead",
  "demo_ready",
  "emailed",
  "interested",
  "client_onboarding",
  "active_client",
  "lost",
  "archived",
] as const;

export function formatDate(iso?: string | null): string {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export function statusTone(status: string): string {
  const s = status.toLowerCase();
  if (["running", "approved", "published", "verified", "completed", "sent"].includes(s)) {
    return "badge-ok";
  }
  if (["queued", "waiting", "draft", "pending"].includes(s)) {
    return "badge-warn";
  }
  if (["failed", "rejected", "cancelled", "lost", "archived"].includes(s)) {
    return "badge-bad";
  }
  return "badge-neutral";
}

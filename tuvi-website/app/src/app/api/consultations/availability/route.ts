import { consultationFetch } from "@/lib/consultationServer";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const date = searchParams.get("date") ?? "";
  const days = searchParams.get("days") ?? "7";
  const qs = new URLSearchParams();
  if (date) qs.set("date", date);
  if (days) qs.set("days", days);

  try {
    const res = await consultationFetch(`/api/v1/consultations/availability?${qs}`, {
      cache: "no-store",
      signal: AbortSignal.timeout(15_000),
    });
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json(
      { status: "error", message: "Consultation service unavailable. Is the API running on port 8090?" },
      { status: 503 },
    );
  }
}

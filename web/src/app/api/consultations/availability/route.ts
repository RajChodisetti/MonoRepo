import { consultationFetch } from "@/lib/consultationServer";

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const query = new URLSearchParams();
  const date = searchParams.get("date")?.trim();
  const days = searchParams.get("days")?.trim();
  if (date) query.set("date", date);
  if (days) query.set("days", days);

  try {
    const response = await consultationFetch(
      `/api/v1/company/consultations/availability?${query.toString()}`,
      { signal: AbortSignal.timeout(15_000) },
    );
    const data = (await response.json()) as unknown;
    return Response.json(data, {
      status: response.status,
      headers: { "Cache-Control": "no-store" },
    });
  } catch {
    return Response.json(
      { status: "error", message: "Consultation availability is temporarily unavailable." },
      { status: 503, headers: { "Cache-Control": "no-store" } },
    );
  }
}

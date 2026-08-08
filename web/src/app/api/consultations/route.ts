import { consultationFetch } from "@/lib/consultationServer";

export async function POST(request: Request) {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return Response.json(
      { status: "error", message: "Invalid request body." },
      { status: 400, headers: { "Cache-Control": "no-store" } },
    );
  }

  try {
    const response = await consultationFetch("/api/v1/company/consultations", {
      method: "POST",
      body: JSON.stringify(body),
      signal: AbortSignal.timeout(30_000),
    });
    const data = (await response.json()) as unknown;
    return Response.json(data, {
      status: response.status,
      headers: { "Cache-Control": "no-store" },
    });
  } catch {
    return Response.json(
      { status: "error", message: "Consultation booking is temporarily unavailable." },
      { status: 503, headers: { "Cache-Control": "no-store" } },
    );
  }
}

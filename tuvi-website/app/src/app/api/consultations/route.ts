import { consultationFetch } from "@/lib/consultationServer";

export async function POST(request: Request) {
  let body: unknown;
  try {
    body = await request.json();
  } catch {
    return Response.json({ status: "error", message: "Invalid request body" }, { status: 400 });
  }

  try {
    const res = await consultationFetch("/api/v1/company/consultations", {
      method: "POST",
      body: JSON.stringify(body),
      signal: AbortSignal.timeout(30_000),
    });
    const data = await res.json();
    return Response.json(data, { status: res.status });
  } catch {
    return Response.json(
      { status: "error", message: "Consultation service unavailable. Please try again shortly." },
      { status: 503 },
    );
  }
}

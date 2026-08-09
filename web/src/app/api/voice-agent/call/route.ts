const OUTBOUND_CALLS_DISABLED = {
  status: "disabled",
  message:
    "Outbound AI callbacks are not available. Use the browser voice assistant to book a consultation.",
} as const;

export async function POST() {
  return Response.json(OUTBOUND_CALLS_DISABLED, {
    status: 403,
    headers: { "Cache-Control": "no-store" },
  });
}

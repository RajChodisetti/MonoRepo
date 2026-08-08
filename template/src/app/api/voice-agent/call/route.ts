const OUTBOUND_CALLS_DISABLED = {
  status: "disabled",
  message: "Outbound AI calls are disabled for this release.",
} as const;

export async function POST() {
  return Response.json(OUTBOUND_CALLS_DISABLED, {
    status: 403,
    headers: { "Cache-Control": "no-store" },
  });
}

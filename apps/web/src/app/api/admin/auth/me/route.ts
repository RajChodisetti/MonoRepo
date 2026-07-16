import { NextResponse } from "next/server";
import { getSessionToken } from "@/lib/session";
import { apiFetch, ApiError } from "@/lib/api";
import type { AdminUser } from "@/lib/types";

export async function GET() {
  const token = await getSessionToken();
  if (!token) {
    return NextResponse.json(
      { error: { message: "Not authenticated." } },
      { status: 401 },
    );
  }
  try {
    const me = await apiFetch<AdminUser>("/api/v1/admin/me", { token });
    return NextResponse.json({ user: me });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json(err.body ?? { error: { message: err.message } }, {
        status: err.status,
      });
    }
    return NextResponse.json(
      { error: { message: "Failed to load profile." } },
      { status: 500 },
    );
  }
}

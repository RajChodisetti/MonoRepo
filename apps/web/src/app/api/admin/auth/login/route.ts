import { NextResponse } from "next/server";
import { apiFetch, ApiError } from "@/lib/api";
import { setSessionToken } from "@/lib/session";
import type { AdminUser } from "@/lib/types";

type LoginResponse = {
  access_token: string;
  user: AdminUser;
};

export async function POST(request: Request) {
  try {
    const body = (await request.json()) as {
      email?: string;
      password?: string;
    };
    if (!body.email || !body.password) {
      return NextResponse.json(
        { error: { message: "Email and password are required." } },
        { status: 400 },
      );
    }

    const login = await apiFetch<LoginResponse>("/api/v1/auth/login", {
      method: "POST",
      body: { email: body.email, password: body.password },
      token: null,
    });

    const me = await apiFetch<AdminUser>("/api/v1/admin/me", {
      token: login.access_token,
    });

    if (me.role !== "internal_admin") {
      return NextResponse.json(
        { error: { message: "internal_admin role required." } },
        { status: 403 },
      );
    }

    await setSessionToken(login.access_token);

    return NextResponse.json({
      user: {
        id: me.id || login.user?.id,
        email: me.email || login.user?.email,
        full_name: me.full_name || login.user?.full_name,
        role: me.role,
      },
    });
  } catch (err) {
    if (err instanceof ApiError) {
      return NextResponse.json(err.body ?? { error: { message: err.message } }, {
        status: err.status,
      });
    }
    return NextResponse.json(
      { error: { message: "Login failed." } },
      { status: 500 },
    );
  }
}

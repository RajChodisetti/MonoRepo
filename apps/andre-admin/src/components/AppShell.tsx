"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { IdleGuard } from "@/components/IdleGuard";

const NAV = [
  { href: "/properties", label: "Properties" },
  { href: "/voice", label: "Voice" },
];

export function AppShell({
  email,
  children,
}: {
  email: string;
  children: React.ReactNode;
}) {
  const pathname = usePathname();
  const router = useRouter();

  async function logout() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.replace("/login");
  }

  return (
    <div className="min-h-screen grid lg:grid-cols-[240px_1fr]">
      <IdleGuard />
      <aside className="bg-sidebar text-sidebar-text px-5 py-6 flex flex-col gap-8">
        <div>
          <p
            className="text-xl tracking-tight leading-tight"
            style={{ fontFamily: "var(--font-fraunces), Georgia, serif" }}
          >
            Real Voice Agent Admin
          </p>
          <p className="text-sm text-sidebar-muted mt-1">Listings &amp; voice ops</p>
        </div>
        <nav className="flex flex-col gap-1">
          {NAV.map((item) => {
            const active = pathname.startsWith(item.href);
            return (
              <Link
                key={item.href}
                href={item.href}
                className={`rounded-lg px-3 py-2 text-sm font-semibold transition ${
                  active ? "bg-white/10 text-white" : "text-sidebar-muted hover:text-white"
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>
        <div className="mt-auto space-y-3">
          <p className="text-xs text-sidebar-muted break-all">{email}</p>
          <button type="button" className="btn btn-ghost w-full !text-sidebar-text !border-white/15" onClick={logout}>
            Log out
          </button>
        </div>
      </aside>
      <main className="px-5 py-6 md:px-8 md:py-8">{children}</main>
    </div>
  );
}

import { AppShell } from "@/components/AppShell";
import { PropertiesClient } from "@/components/PropertiesClient";
import { readSession } from "@/lib/session";
import { redirect } from "next/navigation";

export default async function PropertiesPage() {
  const session = await readSession();
  if (!session) redirect("/login?reason=auth");
  return (
    <AppShell email={session.email}>
      <PropertiesClient />
    </AppShell>
  );
}

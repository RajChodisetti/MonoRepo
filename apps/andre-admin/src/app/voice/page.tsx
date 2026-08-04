import { AppShell } from "@/components/AppShell";
import { VoiceClient } from "@/components/VoiceClient";
import { readSession } from "@/lib/session";
import { redirect } from "next/navigation";

export default async function VoicePage() {
  const session = await readSession();
  if (!session) redirect("/login?reason=auth");
  return (
    <AppShell email={session.email}>
      <VoiceClient />
    </AppShell>
  );
}

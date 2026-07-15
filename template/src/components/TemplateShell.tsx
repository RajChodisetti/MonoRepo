"use client";

import { Suspense, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import type { TemplateId } from "@/lib/templateConfig";
import { parseTemplateId } from "@/lib/templateConfig";
import VoiceAssistantWidget from "@/components/VoiceAssistantWidget";
import SiteWalkthrough from "@/components/SiteWalkthrough";

function TemplateShellInner({
  defaultTemplateId,
  children,
}: {
  defaultTemplateId: TemplateId;
  children: React.ReactNode;
}) {
  const searchParams = useSearchParams();
  const templateId = parseTemplateId(searchParams.get("template")) ?? defaultTemplateId;
  const restaurantIndex = parseInt(searchParams.get("id") ?? "0", 10) || 0;
  const signedSlug = searchParams.get("slug") ?? "";
  const signedToken = searchParams.get("token") ?? "";
  const signedDemo = Boolean(signedSlug || signedToken);

  useEffect(() => {
    document.documentElement.dataset.template = templateId;
  }, [templateId]);

  return (
    <>
      {children}
      {!signedDemo ? (
        <VoiceAssistantWidget templateId={templateId} restaurantIndex={restaurantIndex} />
      ) : null}
      {!signedDemo ? <SiteWalkthrough templateId={templateId} /> : null}
    </>
  );
}

export default function TemplateShell({
  templateId,
  children,
}: {
  templateId: TemplateId;
  children: React.ReactNode;
}) {
  return (
    <Suspense fallback={<>{children}</>}>
      <TemplateShellInner defaultTemplateId={templateId}>{children}</TemplateShellInner>
    </Suspense>
  );
}

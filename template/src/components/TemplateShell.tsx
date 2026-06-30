"use client";

import { Suspense, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import type { TemplateId } from "@/lib/templateConfig";
import { parseTemplateId } from "@/lib/templateConfig";
import VoiceAssistantWidget from "@/components/VoiceAssistantWidget";
import TemplateSwitchPopup from "@/components/TemplateSwitchPopup";

function TemplateShellInner({
  defaultTemplateId,
  children,
}: {
  defaultTemplateId: TemplateId;
  children: React.ReactNode;
}) {
  const searchParams = useSearchParams();
  const templateId = parseTemplateId(searchParams.get("template")) ?? defaultTemplateId;

  useEffect(() => {
    document.documentElement.dataset.template = templateId;
  }, [templateId]);

  return (
    <>
      {children}
      <TemplateSwitchPopup currentTemplateId={templateId} />
      <VoiceAssistantWidget templateId={templateId} />
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

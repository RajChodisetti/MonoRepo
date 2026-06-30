"use client";

import { useEffect } from "react";
import type { TemplateId } from "@/lib/templateConfig";
import VoiceAssistantWidget from "@/components/VoiceAssistantWidget";

export default function TemplateShell({
  templateId,
  children,
}: {
  templateId: TemplateId;
  children: React.ReactNode;
}) {
  useEffect(() => {
    document.documentElement.dataset.template = templateId;
  }, [templateId]);

  return (
    <>
      {children}
      <VoiceAssistantWidget templateId={templateId} />
    </>
  );
}

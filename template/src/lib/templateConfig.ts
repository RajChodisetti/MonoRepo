export type TemplateId = "1" | "2";

export function getActiveTemplate(): TemplateId {
  const raw = process.env.TEMPLATE ?? "1";
  return raw === "2" ? "2" : "1";
}

export function parseTemplateId(value?: string | null): TemplateId | null {
  if (value === "1" || value === "2") return value;
  return null;
}

export function resolveTemplate(override?: string | null): TemplateId {
  return parseTemplateId(override) ?? getActiveTemplate();
}

export function getOtherTemplate(id: TemplateId): TemplateId {
  return id === "1" ? "2" : "1";
}

export function getTemplateLabel(id: TemplateId): string {
  return id === "2" ? "Aurora" : "Cinematic";
}

export function getTemplateSwitchCopy(current: TemplateId) {
  if (current === "1") {
    return {
      eyebrow: "New look available",
      title: "Try our Aurora template",
      description:
        "Switch to a futuristic glass design with interactive sections and a bold tech feel.",
      cta: "Switch to Aurora",
      targetLabel: "Aurora",
    };
  }

  return {
    eyebrow: "Classic option",
    title: "Try our Cinematic template",
    description:
      "Switch to a warm, editorial dining experience with scroll storytelling and elegant typography.",
    cta: "Switch to Cinematic",
    targetLabel: "Cinematic",
  };
}

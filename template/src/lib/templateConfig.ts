export type TemplateId = "1" | "2" | "3" | "4";

const TEMPLATE_CYCLE: TemplateId[] = ["1", "2", "3", "4"];

export function getActiveTemplate(): TemplateId {
  const raw = process.env.TEMPLATE ?? "1";
  if (raw === "2") return "2";
  if (raw === "3") return "3";
  if (raw === "4") return "4";
  return "1";
}

export function parseTemplateId(value?: string | null): TemplateId | null {
  if (value === "1" || value === "2" || value === "3" || value === "4") return value;
  return null;
}

export function resolveTemplate(override?: string | null): TemplateId {
  return parseTemplateId(override) ?? getActiveTemplate();
}

export function getNextTemplate(id: TemplateId): TemplateId {
  const idx = TEMPLATE_CYCLE.indexOf(id);
  return TEMPLATE_CYCLE[(idx + 1) % TEMPLATE_CYCLE.length];
}

/** @deprecated Use getNextTemplate — kept for compatibility */
export function getOtherTemplate(id: TemplateId): TemplateId {
  return getNextTemplate(id);
}

export function getTemplateLabel(id: TemplateId): string {
  if (id === "2") return "Aurora";
  if (id === "3") return "Elysian";
  if (id === "4") return "Foodie";
  return "Cinematic";
}

export function getTemplateSwitchCopy(current: TemplateId) {
  const next = getNextTemplate(current);
  const targetLabel = getTemplateLabel(next);

  if (current === "1") {
    return {
      eyebrow: "New look available",
      title: "Try our Aurora template",
      description:
        "Switch to a futuristic glass design with interactive sections and a bold tech feel.",
      cta: `Switch to ${targetLabel}`,
      targetLabel,
    };
  }

  if (current === "2") {
    return {
      eyebrow: "Premium option",
      title: "Try our Elysian template",
      description:
        "Switch to an ultra-premium gold and black fine dining experience with cinematic interactions.",
      cta: `Switch to ${targetLabel}`,
      targetLabel,
    };
  }

  if (current === "3") {
    return {
      eyebrow: "Fresh & friendly",
      title: "Try our Foodie template",
      description:
        "Switch to a bright, appetizing cream and orange layout built for casual dining and quick orders.",
      cta: `Switch to ${targetLabel}`,
      targetLabel,
    };
  }

  return {
    eyebrow: "Classic option",
    title: "Try our Cinematic template",
    description:
      "Switch to a warm, editorial dining experience with scroll storytelling and elegant typography.",
    cta: `Switch to ${targetLabel}`,
    targetLabel,
  };
}

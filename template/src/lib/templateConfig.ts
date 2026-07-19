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
  if (id === "4") return "Italian Villa";
  return "Cinematic";
}

export function getTemplateSwitchCopy(current: TemplateId) {
  const next = getNextTemplate(current);
  const targetLabel = getTemplateLabel(next);
  const currentLabel = getTemplateLabel(current);
  const position = TEMPLATE_CYCLE.indexOf(current) + 1;
  const shared = { currentLabel, targetLabel, position, total: TEMPLATE_CYCLE.length };

  if (current === "1") {
    return {
      ...shared,
      eyebrow: "Same restaurant, new look",
      title: "Preview the Aurora design",
      description:
        "Keep the same restaurant details and photos in a futuristic glass design with a bold visual feel.",
      cta: `Preview ${targetLabel}`,
    };
  }

  if (current === "2") {
    return {
      ...shared,
      eyebrow: "Same restaurant, new look",
      title: "Preview the Elysian design",
      description:
        "Keep the same restaurant details and photos in a premium gold-and-black dining design.",
      cta: `Preview ${targetLabel}`,
    };
  }

  if (current === "3") {
    return {
      ...shared,
      eyebrow: "Same restaurant, new look",
      title: "Preview the Italian Villa design",
      description:
        "Keep the same restaurant details and photos in an Italian cuisine-first editorial design.",
      cta: `Preview ${targetLabel}`,
    };
  }

  return {
    ...shared,
    eyebrow: "Same restaurant, new look",
    title: "Preview the Cinematic design",
    description:
      "Keep the same restaurant details and photos in a warm editorial design with elegant typography.",
    cta: `Preview ${targetLabel}`,
  };
}

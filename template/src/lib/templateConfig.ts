export type TemplateId = "1" | "2";

export function getActiveTemplate(): TemplateId {
  const raw = process.env.TEMPLATE ?? "1";
  return raw === "2" ? "2" : "1";
}

export function getTemplateLabel(id: TemplateId): string {
  return id === "2" ? "Aurora" : "Cinematic";
}

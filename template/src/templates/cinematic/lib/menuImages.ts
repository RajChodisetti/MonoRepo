import imageClassifications from "../../../../../data/image_classifications.json";

/** Skip printed-menu photos; only show actual food shots on menu cards. */
export function isFoodMenuImage(url: string): boolean {
  const entry =
    imageClassifications.classifications[
      url as keyof typeof imageClassifications.classifications
    ];
  if (!entry) return true;
  return entry.type === "food";
}

export function shortenCategoryLabel(cat: string): string {
  return cat
    .replace(/^PLATS DU JOUR\s*-\s*LES PLATS DU JOUR\s*-\s*/i, "")
    .replace(/^A LA CARTE\s*-\s*/i, "")
    .replace(/\s+/g, " ")
    .trim();
}

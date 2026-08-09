/** Menu images now arrive only through approved owner/licensed media. */
export function isFoodMenuImage(url: string): boolean {
  return url.trim().length > 0;
}

export function shortenCategoryLabel(cat: string): string {
  return cat
    .replace(/^PLATS DU JOUR\s*-\s*LES PLATS DU JOUR\s*-\s*/i, "")
    .replace(/^A LA CARTE\s*-\s*/i, "")
    .replace(/\s+/g, " ")
    .trim();
}

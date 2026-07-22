import type { Metadata } from "next";
import { foodieContent } from "./lib/foodieContent";

export function buildFoodieMetadata(): Metadata {
  const { brand, hero } = foodieContent;
  const title = `${brand.name} Restaurant — Enjoy The Food`;

  return {
    title,
    description: hero.description,
    openGraph: {
      title,
      description: hero.description,
      images: [{ url: hero.plate }],
    },
  };
}

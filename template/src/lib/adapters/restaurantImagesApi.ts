import type { GalleryImage } from "@/data/types/gallery";

type ApiGalleryImage = {
  url: string;
  thumbnail_url?: string;
  image_type?: string;
  confidence?: number;
  title?: string;
};

export type SiteImagesResponse = {
  restaurant_id: string;
  gallery_images: ApiGalleryImage[];
};

function galleryType(imageType?: string): GalleryImage["type"] {
  const t = (imageType || "").toLowerCase();
  if (t === "food_photo" || t === "food") return "food";
  if (t === "interior" || t === "ambience") return "ambience";
  return "other";
}

export function mapSiteImagesResponse(
  data: SiteImagesResponse,
  restaurantName: string,
): { galleryImages: GalleryImage[] } {
  const galleryImages: GalleryImage[] = (data.gallery_images || [])
    .filter((image) => !["menu_document", "menu_list", "menu_ocr"].includes((image.image_type || "").toLowerCase()))
    .map((img, i) => ({
      url: img.url,
      alt: img.title ? `${restaurantName} — ${img.title}` : `${restaurantName} gallery ${i + 1}`,
      type: galleryType(img.image_type),
    }));

  return { galleryImages };
}

export async function fetchSiteImagesByPlaceID(
  placeID: string,
  restaurantName: string,
): Promise<{ galleryImages: GalleryImage[] } | null> {
  const apiBase = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "");
  if (!apiBase || !placeID) return null;

  try {
    const res = await fetch(
      `${apiBase}/api/public/v1/restaurants/by-place/${encodeURIComponent(placeID)}/site-images`,
      { cache: "no-store" },
    );
    if (!res.ok) return null;
    const data = (await res.json()) as SiteImagesResponse;
    return mapSiteImagesResponse(data, restaurantName);
  } catch {
    return null;
  }
}

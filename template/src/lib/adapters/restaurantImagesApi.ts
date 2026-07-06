import type { GalleryImage } from "@/data/types/gallery";
import type { MenuListImage } from "@/data/types/menuImages";

type ApiMenuImage = {
  url: string;
  thumbnail_url?: string;
  image_type?: string;
  confidence?: number;
  title?: string;
};

type ApiGalleryImage = {
  url: string;
  thumbnail_url?: string;
  image_type?: string;
  confidence?: number;
  title?: string;
};

export type SiteImagesResponse = {
  restaurant_id: string;
  menu_images: ApiMenuImage[];
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
): { menuListImages: MenuListImage[]; galleryImages: GalleryImage[] } {
  const menuListImages: MenuListImage[] = (data.menu_images || []).map((img, i) => ({
    url: img.url,
    thumbnail: img.thumbnail_url || img.url,
    alt: img.title ? `${restaurantName} menu — ${img.title}` : `${restaurantName} menu ${i + 1}`,
    confidence: img.confidence,
  }));

  const galleryImages: GalleryImage[] = (data.gallery_images || []).map((img, i) => ({
    url: img.url,
    alt: img.title ? `${restaurantName} — ${img.title}` : `${restaurantName} gallery ${i + 1}`,
    type: galleryType(img.image_type),
  }));

  return { menuListImages, galleryImages };
}

export async function fetchSiteImagesByPlaceID(
  placeID: string,
  restaurantName: string,
): Promise<{ menuListImages: MenuListImage[]; galleryImages: GalleryImage[] } | null> {
  const apiBase = process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "");
  if (!apiBase || !placeID) return null;

  try {
    const res = await fetch(
      `${apiBase}/api/public/v1/restaurants/by-place/${encodeURIComponent(placeID)}/site-images`,
      { next: { revalidate: 60 } },
    );
    if (!res.ok) return null;
    const data = (await res.json()) as SiteImagesResponse;
    return mapSiteImagesResponse(data, restaurantName);
  } catch {
    return null;
  }
}

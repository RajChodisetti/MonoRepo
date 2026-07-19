export interface GalleryImage {
  url: string;
  alt: string;
  type: "food" | "ambience" | "other";
  mediaType?: "exterior" | "interior" | "food" | "drink" | "logo" | "team" | "event" | "other";
  sourceKind?: "google_places_live" | "owner_upload" | "licensed" | "legacy_public_url";
  caption?: string;
  tags?: string[];
  qualityScore?: number;
  heroScore?: number;
  width?: number;
  height?: number;
  orientation?: string;
  subjectPosition?: string;
  placementRole?: string;
  unoptimized?: boolean;
  authorAttributions?: { displayName: string; uri?: string; photoUri?: string }[];
  googleMapsUri?: string;
  flagContentUri?: string;
}

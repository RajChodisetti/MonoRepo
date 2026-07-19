import type { GalleryImage } from "@/data/types/gallery";

export default function PhotoAttribution({ media, compact = false }: { media: GalleryImage; compact?: boolean }) {
  if (media.sourceKind !== "google_places_live") {
    return media.caption && !compact ? <p className="mt-3 text-sm opacity-70">{media.caption}</p> : null;
  }
  return (
    <div className={`flex flex-wrap items-center gap-x-2 gap-y-1 ${compact ? "text-[10px]" : "text-xs"}`}>
      <span translate="no">Google Maps</span>
      {media.authorAttributions?.map((author, index) => (
        <span key={`${author.displayName}-${index}`}>
          Photo by{" "}
          {author.uri ? (
            <a href={author.uri} target="_blank" rel="noreferrer" className="underline underline-offset-2">
              {author.displayName || "contributor"}
            </a>
          ) : (
            author.displayName || "contributor"
          )}
        </span>
      ))}
      {media.googleMapsUri ? (
        <a href={media.googleMapsUri} target="_blank" rel="noreferrer" className="underline underline-offset-2">
          View source
        </a>
      ) : null}
      {media.flagContentUri ? (
        <a href={media.flagContentUri} target="_blank" rel="noreferrer" className="underline underline-offset-2">
          Report photo
        </a>
      ) : null}
    </div>
  );
}

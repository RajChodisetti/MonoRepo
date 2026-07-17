"use client";

/* eslint-disable @next/next/no-img-element -- Places media hosts and URLs are dynamic/temporary. */

import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import type {
  GooglePlacePhoto,
  GooglePlacePhotos,
  RestaurantImage,
  RestaurantImages,
} from "@/lib/types";
import { EmptyState, ErrorBanner } from "@/components/ui";

function ImageTile({
  image,
  kind,
  restaurantId,
  onChanged,
}: {
  image: RestaurantImage;
  kind: "menu" | "gallery";
  restaurantId: string;
  onChanged: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const hidden = !!image.hidden_at;

  async function toggle() {
    setBusy(true);
    try {
      if (hidden) {
        await adminFetch(`restaurants/${restaurantId}/images/${kind}/${image.id}/restore`, {
          method: "POST",
        });
      } else {
        await adminFetch(`restaurants/${restaurantId}/images/${kind}/${image.id}`, {
          method: "DELETE",
        });
      }
      onChanged();
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="photo-tile" data-hidden={hidden}>
      <img src={image.thumbnail_url || image.url} alt={image.title || kind} loading="lazy" />
      <div className="photo-tile-body">
        <span style={{ color: "var(--muted)" }}>
          {image.title || image.image_type || kind}
        </span>
        <button
          type="button"
          className={hidden ? "btn btn-secondary" : "btn btn-danger"}
          onClick={toggle}
          disabled={busy}
        >
          {hidden ? "Restore" : "Remove"}
        </button>
        <a href={image.url} target="_blank" rel="noreferrer">
          Open full image
        </a>
        <code className="photo-url">{image.url}</code>
      </div>
    </div>
  );
}

function GooglePhotoTile({ photo, index }: { photo: GooglePlacePhoto; index: number }) {
  return (
    <div className="photo-tile">
      <img src={photo.url} alt={`Google Places photo ${index + 1}`} loading="lazy" />
      <div className="photo-tile-body">
        <span style={{ color: "var(--muted)" }}>
          Google Places · {photo.width_px || "?"}×{photo.height_px || "?"}
        </span>
        <a href={photo.url} target="_blank" rel="noreferrer">
          Open full image
        </a>
        <code className="photo-url">{photo.url}</code>
        {photo.author_attributions.length > 0 ? (
          <span style={{ color: "var(--muted)", fontSize: "0.78rem" }}>
            Photo by{" "}
            {photo.author_attributions.map((item, attributionIndex) => (
              <span key={`${item.display_name}-${attributionIndex}`}>
                {attributionIndex > 0 ? ", " : ""}
                {item.uri ? (
                  <a href={item.uri} target="_blank" rel="noreferrer">
                    {item.display_name || "Google Maps contributor"}
                  </a>
                ) : (
                  item.display_name || "Google Maps contributor"
                )}
              </span>
            ))}
          </span>
        ) : null}
      </div>
    </div>
  );
}

export function PhotoGallery({ restaurantId }: { restaurantId: string }) {
  const [images, setImages] = useState<RestaurantImages | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [googlePhotos, setGooglePhotos] = useState<GooglePlacePhotos | null>(null);
  const [googleError, setGoogleError] = useState<string | null>(null);
  const [googleBusy, setGoogleBusy] = useState(false);
  const [showHidden, setShowHidden] = useState(false);

  const load = useCallback(async () => {
    setError(null);
    try {
      const data = await adminFetch<RestaurantImages>(`restaurants/${restaurantId}/images`);
      setImages(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load photos");
    }
  }, [restaurantId]);

  const loadGooglePhotos = useCallback(async () => {
    setGoogleBusy(true);
    setGoogleError(null);
    try {
      const data = await adminFetch<GooglePlacePhotos>(
        `restaurants/${restaurantId}/images/google`,
      );
      setGooglePhotos(data);
    } catch (err) {
      setGoogleError(
        err instanceof Error ? err.message : "Failed to refresh Google photos",
      );
    } finally {
      setGoogleBusy(false);
    }
  }, [restaurantId]);

  useEffect(() => {
    load();
    loadGooglePhotos();
  }, [load, loadGooglePhotos]);

  if (error) return <ErrorBanner message={error} />;
  if (!images) return <EmptyState message="Loading photos…" />;

  const menu = showHidden ? images.menu_images : images.menu_images.filter((i) => !i.hidden_at);
  const gallery = showHidden ? images.gallery_images : images.gallery_images.filter((i) => !i.hidden_at);

  return (
    <div style={{ display: "grid", gap: "1rem" }}>
      <div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: "0.75rem",
            marginBottom: "0.5rem",
          }}
        >
          <div>
            <h3 style={{ margin: 0, fontSize: "0.95rem" }}>
              Google Places photos ({googlePhotos?.photos.length || 0})
            </h3>
            <p style={{ color: "var(--muted)", fontSize: "0.82rem", margin: "0.25rem 0 0" }}>
              Live server-resolved URLs. They can expire, so this list refreshes when the tab opens.
            </p>
          </div>
          <button
            type="button"
            className="btn btn-secondary"
            onClick={loadGooglePhotos}
            disabled={googleBusy}
          >
            {googleBusy ? "Refreshing…" : "Refresh URLs"}
          </button>
        </div>
        <ErrorBanner message={googleError} />
        {googleBusy && !googlePhotos ? <EmptyState message="Resolving Google photo URLs…" /> : null}
        {!googleBusy && googlePhotos && googlePhotos.photos.length === 0 ? (
          <EmptyState message="Google Places returned no photos for this restaurant." />
        ) : null}
        {googlePhotos && googlePhotos.photos.length > 0 ? (
          <div className="photo-grid">
            {googlePhotos.photos.map((photo, index) => (
              <GooglePhotoTile key={`${photo.url}-${index}`} photo={photo} index={index} />
            ))}
          </div>
        ) : null}
      </div>

      <label style={{ display: "flex", alignItems: "center", gap: "0.45rem", fontSize: "0.9rem" }}>
        <input type="checkbox" checked={showHidden} onChange={(e) => setShowHidden(e.target.checked)} />
        Show removed photos
      </label>

      <div>
        <h3 style={{ margin: "0 0 0.5rem", fontSize: "0.95rem" }}>OCR menu images ({menu.length})</h3>
        {menu.length === 0 ? (
          <EmptyState message="No menu images." />
        ) : (
          <div className="photo-grid">
            {menu.map((img) => (
              <ImageTile key={img.id} image={img} kind="menu" restaurantId={restaurantId} onChanged={load} />
            ))}
          </div>
        )}
      </div>

      <div>
        <h3 style={{ margin: "0 0 0.5rem", fontSize: "0.95rem" }}>OCR gallery images ({gallery.length})</h3>
        {gallery.length === 0 ? (
          <EmptyState message="No gallery images." />
        ) : (
          <div className="photo-grid">
            {gallery.map((img) => (
              <ImageTile key={img.id} image={img} kind="gallery" restaurantId={restaurantId} onChanged={load} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

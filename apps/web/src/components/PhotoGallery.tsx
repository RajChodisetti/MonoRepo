"use client";

/* eslint-disable @next/next/no-img-element -- Places media hosts and URLs are dynamic/temporary. */

import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import type {
  GooglePlacePhoto,
  GooglePlacePhotos,
  RestaurantImage,
  RestaurantImages,
  RestaurantOwnedMedia,
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
        {photo.google_maps_uri ? (
          <a href={photo.google_maps_uri} target="_blank" rel="noreferrer">
            View source on Google Maps
          </a>
        ) : null}
        {photo.flag_content_uri ? (
          <a href={photo.flag_content_uri} target="_blank" rel="noreferrer">
            Report this photo
          </a>
        ) : null}
      </div>
    </div>
  );
}

function OwnedMediaTile({
  image,
  restaurantId,
  onChanged,
}: {
  image: RestaurantOwnedMedia;
  restaurantId: string;
  onChanged: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const hidden = !!image.hidden_at;

  async function toggle() {
    setBusy(true);
    try {
      await adminFetch(
        hidden
          ? `restaurants/${restaurantId}/media/${image.id}/restore`
          : `restaurants/${restaurantId}/media/${image.id}`,
        { method: hidden ? "POST" : "DELETE" },
      );
      onChanged();
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="photo-tile" data-hidden={hidden}>
      <img src={image.url} alt={image.alt_text || image.caption || image.media_type} loading="lazy" />
      <div className="photo-tile-body">
        <strong>{image.caption || image.media_type}</strong>
        <span style={{ color: "var(--muted)" }}>
          {image.source_kind} · {image.placement_role || "gallery"} · OCR {image.vision_status || "pending"}
          {image.approval_status ? ` · ${image.approval_status}` : ""}
        </span>
        {image.vision_last_error ? (
          <span style={{ color: "var(--danger)", fontSize: "0.78rem" }}>{image.vision_last_error}</span>
        ) : null}
        <button
          type="button"
          className={hidden ? "btn btn-secondary" : "btn btn-danger"}
          onClick={toggle}
          disabled={busy}
        >
          {hidden ? "Restore" : "Remove"}
        </button>
        <a href={image.url} target="_blank" rel="noreferrer">Open full image</a>
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
  const [uploadBusy, setUploadBusy] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
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

  async function uploadOwnedMedia(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setUploadBusy(true);
    setUploadError(null);
    const form = event.currentTarget;
    try {
      await adminFetch(`restaurants/${restaurantId}/media`, {
        method: "POST",
        body: new FormData(form),
      });
      form.reset();
      await load();
    } catch (err) {
      setUploadError(err instanceof Error ? err.message : "Failed to upload media");
    } finally {
      setUploadBusy(false);
    }
  }

  if (error) return <ErrorBanner message={error} />;
  if (!images) return <EmptyState message="Loading photos…" />;

  const menu = showHidden ? images.menu_images : images.menu_images.filter((i) => !i.hidden_at);
  const gallery = showHidden ? images.gallery_images : images.gallery_images.filter((i) => !i.hidden_at);
  const owned = showHidden ? images.owned_media : images.owned_media.filter((i) => !i.hidden_at);

  return (
    <div style={{ display: "grid", gap: "1rem" }}>
      <div>
        <h3 style={{ margin: "0 0 0.35rem", fontSize: "0.95rem" }}>Owned or licensed website media</h3>
        <p style={{ color: "var(--muted)", fontSize: "0.82rem", margin: "0 0 0.75rem" }}>
          Uploaded files remain private until background OCR confirms they are not menu documents. Only approved media appears on personalized sites.
        </p>
        <form onSubmit={uploadOwnedMedia} style={{ display: "grid", gap: "0.65rem", maxWidth: "52rem" }}>
          <input name="file" type="file" accept="image/jpeg,image/png,image/gif" required />
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(11rem, 1fr))", gap: "0.65rem" }}>
            <label>Type
              <select name="media_type" defaultValue="other" required>
                <option value="exterior">Exterior</option>
                <option value="interior">Interior</option>
                <option value="food">Food</option>
                <option value="drink">Drink</option>
                <option value="logo">Logo</option>
                <option value="team">Team</option>
                <option value="event">Event</option>
                <option value="other">Other</option>
              </select>
            </label>
            <label>Placement
              <select name="placement_role" defaultValue="gallery">
                <option value="hero">Hero</option>
                <option value="about">About</option>
                <option value="gallery">Gallery</option>
                <option value="food_gallery">Food gallery</option>
                <option value="ambience_gallery">Ambience gallery</option>
                <option value="logo">Logo</option>
              </select>
            </label>
            <label>Rights
              <select name="rights_status" defaultValue="owner_granted" required>
                <option value="owner_granted">Owner granted</option>
                <option value="licensed">Separately licensed</option>
              </select>
            </label>
          </div>
          <input name="caption" type="text" maxLength={180} placeholder="Optional factual caption" />
          <input name="alt_text" type="text" maxLength={180} placeholder="Optional accessible description" />
          <ErrorBanner message={uploadError} />
          <button type="submit" className="btn btn-primary" disabled={uploadBusy} style={{ justifySelf: "start" }}>
            {uploadBusy ? "Uploading…" : "Upload for OCR review"}
          </button>
        </form>
        {owned.length === 0 ? (
          <EmptyState message="No owned or licensed website media uploaded yet." />
        ) : (
          <div className="photo-grid" style={{ marginTop: "0.75rem" }}>
            {owned.map((image) => (
              <OwnedMediaTile key={image.id} image={image} restaurantId={restaurantId} onChanged={load} />
            ))}
          </div>
        )}
      </div>

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
        <h3 style={{ margin: "0 0 0.5rem", fontSize: "0.95rem" }}>OCR menu images — admin only ({menu.length})</h3>
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

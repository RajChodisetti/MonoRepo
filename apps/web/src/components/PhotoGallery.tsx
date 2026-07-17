"use client";

import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import type { RestaurantImage, RestaurantImages } from "@/lib/types";
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
      </div>
    </div>
  );
}

export function PhotoGallery({ restaurantId }: { restaurantId: string }) {
  const [images, setImages] = useState<RestaurantImages | null>(null);
  const [error, setError] = useState<string | null>(null);
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

  useEffect(() => {
    load();
  }, [load]);

  if (error) return <ErrorBanner message={error} />;
  if (!images) return <EmptyState message="Loading photos…" />;

  const menu = showHidden ? images.menu_images : images.menu_images.filter((i) => !i.hidden_at);
  const gallery = showHidden ? images.gallery_images : images.gallery_images.filter((i) => !i.hidden_at);

  return (
    <div style={{ display: "grid", gap: "1rem" }}>
      <label style={{ display: "flex", alignItems: "center", gap: "0.45rem", fontSize: "0.9rem" }}>
        <input type="checkbox" checked={showHidden} onChange={(e) => setShowHidden(e.target.checked)} />
        Show removed photos
      </label>

      <div>
        <h3 style={{ margin: "0 0 0.5rem", fontSize: "0.95rem" }}>Menu images ({menu.length})</h3>
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
        <h3 style={{ margin: "0 0 0.5rem", fontSize: "0.95rem" }}>Gallery images ({gallery.length})</h3>
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

"use client";

import type {
  ProfileReviewPreview,
  RestaurantProfile,
  SiteContent,
  SiteImage,
  SiteImages,
} from "@/lib/types";
import { StatusBadge } from "@/components/ui";

function asText(value: unknown): string {
  if (value === null || value === undefined || value === "") return "—";
  if (typeof value === "string" || typeof value === "number") return String(value);
  if (Array.isArray(value)) {
    return value
      .map((v) => (typeof v === "string" ? v : JSON.stringify(v)))
      .filter(Boolean)
      .join(", ");
  }
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

function collectProfileImageUrls(images: unknown): string[] {
  const urls: string[] = [];
  if (!images) return urls;

  if (typeof images === "string" && /^https?:\/\//i.test(images)) {
    urls.push(images);
    return urls;
  }

  if (Array.isArray(images)) {
    for (const item of images) {
      if (typeof item === "string" && /^https?:\/\//i.test(item)) urls.push(item);
      else if (item && typeof item === "object") {
        const o = item as Record<string, unknown>;
        const u = o.url || o.thumbnail_url || o.thumbnail || o.src;
        if (typeof u === "string" && /^https?:\/\//i.test(u)) urls.push(u);
      }
    }
    return urls;
  }

  if (typeof images === "object") {
    const o = images as Record<string, unknown>;
    for (const key of [
      "thumbnail",
      "hero",
      "cover",
      "primary",
      "main",
      "url",
    ]) {
      const v = o[key];
      if (typeof v === "string" && /^https?:\/\//i.test(v)) urls.push(v);
    }
    for (const [k, v] of Object.entries(o)) {
      if (k === "thumbnail" || k === "hero") continue;
      if (typeof v === "string" && /^https?:\/\//i.test(v)) urls.push(v);
      if (Array.isArray(v)) {
        urls.push(...collectProfileImageUrls(v));
      }
    }
  }

  return Array.from(new Set(urls));
}

function ImageGrid({
  title,
  images,
  emptyHint,
}: {
  title: string;
  images: { url: string; caption?: string }[];
  emptyHint?: string;
}) {
  if (!images.length) {
    return (
      <div className="card">
        <h3 style={{ margin: "0 0 0.5rem", fontSize: "1rem" }}>{title}</h3>
        <p style={{ margin: 0, color: "var(--muted)" }}>
          {emptyHint || "No images found."}
        </p>
      </div>
    );
  }

  return (
    <div className="card">
      <h3 style={{ margin: "0 0 0.75rem", fontSize: "1rem" }}>
        {title}{" "}
        <span style={{ color: "var(--muted)", fontWeight: 500 }}>
          ({images.length})
        </span>
      </h3>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fill, minmax(140px, 1fr))",
          gap: "0.65rem",
        }}
      >
        {images.map((img) => (
          <a
            key={img.url}
            href={img.url}
            target="_blank"
            rel="noreferrer"
            style={{
              display: "block",
              border: "1px solid var(--line)",
              background: "var(--bg)",
              overflow: "hidden",
            }}
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={img.url}
              alt={img.caption || title}
              style={{
                width: "100%",
                height: 120,
                objectFit: "cover",
                display: "block",
              }}
              loading="lazy"
            />
            {img.caption ? (
              <div
                style={{
                  padding: "0.35rem 0.45rem",
                  fontSize: "0.75rem",
                  color: "var(--muted)",
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                }}
              >
                {img.caption}
              </div>
            ) : null}
          </a>
        ))}
      </div>
    </div>
  );
}

function mapSiteImages(list?: SiteImage[]) {
  return (list || [])
    .map((img) => ({
      url: img.thumbnail_url || img.url || "",
      caption: img.title || img.image_type || img.source,
    }))
    .filter((img) => !!img.url);
}

function Field({ label, value }: { label: string; value: unknown }) {
  const text = asText(value);
  const multiline = text.includes("\n") || text.length > 120;
  return (
    <div style={{ display: "grid", gap: "0.2rem" }}>
      <div
        style={{
          fontSize: "0.75rem",
          textTransform: "uppercase",
          letterSpacing: "0.04em",
          color: "var(--muted)",
          fontWeight: 600,
        }}
      >
        {label}
      </div>
      {multiline ? (
        <pre
          style={{
            margin: 0,
            whiteSpace: "pre-wrap",
            fontSize: "0.85rem",
            fontFamily: "inherit",
          }}
        >
          {text}
        </pre>
      ) : (
        <div style={{ fontSize: "0.95rem" }}>{text}</div>
      )}
    </div>
  );
}

export function ScrapedDetailsPanel({
  preview,
  siteImages,
  siteContent,
  loading,
}: {
  preview: ProfileReviewPreview | null;
  siteImages: SiteImages | null;
  siteContent: SiteContent | null;
  loading?: boolean;
}) {
  if (loading) {
    return (
      <div className="card" style={{ color: "var(--muted)" }}>
        Loading scraped details…
      </div>
    );
  }

  if (!preview && !siteImages && !siteContent) {
    return (
      <div className="card" style={{ color: "var(--muted)" }}>
        No scraped profile found for this restaurant yet.
      </div>
    );
  }

  const profile: RestaurantProfile = preview?.profile || {};
  const profileUrls = collectProfileImageUrls(profile.images);
  if (siteContent?.thumbnail) profileUrls.unshift(siteContent.thumbnail);

  const imagesMeta =
    profile.images && typeof profile.images === "object" && !Array.isArray(profile.images)
      ? (profile.images as Record<string, unknown>)
      : {};
  const googlePhotoCount =
    typeof imagesMeta.google_photo_count === "number"
      ? imagesMeta.google_photo_count
      : 0;
  const pendingPhotosHint =
    googlePhotoCount > 0
      ? `Google reports ${googlePhotoCount} photo(s), but expiring URLs are not stored in the database. Open the Photos tab to resolve the current attributed reference URLs.`
      : undefined;

  const gallery = [
    ...mapSiteImages(siteImages?.gallery_images),
    ...mapSiteImages(siteContent?.gallery_images),
  ];
  const menus = [
    ...mapSiteImages(siteImages?.menu_images),
    ...mapSiteImages(siteContent?.menu_images),
  ];
  const profileGallery = profileUrls.map((url) => ({ url, caption: "profile" }));

  // de-dupe by url
  const uniq = (items: { url: string; caption?: string }[]) => {
    const seen = new Set<string>();
    return items.filter((i) => {
      if (!i.url || seen.has(i.url)) return false;
      seen.add(i.url);
      return true;
    });
  };

  const hours =
    profile.opening_hours ??
    (siteContent as { hours?: unknown } | null)?.hours;

  return (
    <div style={{ display: "grid", gap: "1rem" }}>
      <div className="card">
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "0.5rem",
            alignItems: "center",
            marginBottom: "0.85rem",
          }}
          >
          <h2 style={{ margin: 0, fontSize: "1.15rem" }}>Scraped details</h2>
          <StatusBadge status={preview?.review_status || "review"} />
          {profile.scrape_status ? (
            <StatusBadge status={String(profile.scrape_status)} />
          ) : null}
        </div>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "0.85rem",
          }}
        >
          <Field label="Name" value={preview?.restaurant_name} />
          <Field label="Contact email" value={preview?.contact_email} />
          <Field label="Phone" value={profile.phone} />
          <Field label="Website" value={profile.website} />
          <Field label="Address" value={profile.address} />
          <Field label="City" value={profile.city} />
          <Field label="State" value={profile.state} />
          <Field label="Country" value={profile.country} />
          <Field label="Rating" value={profile.rating} />
          <Field label="Reviews count" value={profile.reviews_count} />
          <Field label="Price level" value={profile.price_level} />
          <Field label="Google Place ID" value={profile.google_place_id} />
          <Field label="Cuisines" value={profile.cuisines} />
          <Field label="Owners" value={profile.owners} />
          <Field label="Dietary" value={profile.dietary_options} />
          <Field label="Parking" value={profile.parking_info} />
          <Field
            label="Lat / Lng"
            value={
              profile.latitude != null || profile.longitude != null
                ? `${profile.latitude ?? "—"}, ${profile.longitude ?? "—"}`
                : null
            }
          />
        </div>

        {profile.description ? (
          <div style={{ marginTop: "1rem" }}>
            <Field label="Description" value={profile.description} />
          </div>
        ) : null}

        <div style={{ marginTop: "1rem" }}>
          <Field label="Opening hours" value={hours} />
        </div>

        {profile.scrape_errors ? (
          <div style={{ marginTop: "1rem" }}>
            <Field label="Scrape errors" value={profile.scrape_errors} />
          </div>
        ) : null}
      </div>

      <ImageGrid
        title="Profile / hero images"
        images={uniq(profileGallery)}
        emptyHint={pendingPhotosHint}
      />
      <ImageGrid
        title="Gallery photos"
        images={uniq(gallery)}
        emptyHint={
          pendingPhotosHint ||
          "No gallery rows in gallery_images / site content yet."
        }
      />
      <ImageGrid
        title="Menu board photos"
        images={uniq(menus)}
        emptyHint="No menu board reference images are stored."
      />

      {(siteContent?.menu_items?.length || 0) > 0 ? (
        <div className="card">
          <h3 style={{ margin: "0 0 0.75rem", fontSize: "1rem" }}>
            Menu items ({siteContent?.menu_items?.length})
          </h3>
          <div className="table-wrap" style={{ border: "none" }}>
            <table className="data">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Category</th>
                  <th>Price</th>
                  <th>Description</th>
                </tr>
              </thead>
              <tbody>
                {(siteContent?.menu_items || []).slice(0, 100).map((item, i) => (
                  <tr key={`${item.name}-${i}`}>
                    <td>{item.name || "—"}</td>
                    <td>{item.category || "—"}</td>
                    <td>{item.price || "—"}</td>
                    <td
                      style={{
                        maxWidth: 280,
                        whiteSpace: "normal",
                      }}
                    >
                      {item.description || "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : null}

      {(siteContent?.reviews?.length || 0) > 0 ? (
        <div className="card">
          <h3 style={{ margin: "0 0 0.75rem", fontSize: "1rem" }}>
            Reviews ({siteContent?.reviews?.length})
          </h3>
          <div style={{ display: "grid", gap: "0.65rem" }}>
            {(siteContent?.reviews || []).slice(0, 20).map((r, i) => (
              <div
                key={`${r.reviewer}-${i}`}
                style={{
                  borderTop: i ? "1px solid var(--line)" : undefined,
                  paddingTop: i ? "0.65rem" : 0,
                }}
              >
                <div style={{ fontWeight: 600 }}>
                  {r.reviewer || "Reviewer"}
                  {r.stars != null ? ` · ${r.stars}★` : ""}
                  {r.date ? (
                    <span style={{ color: "var(--muted)", fontWeight: 400 }}>
                      {" "}
                      · {r.date}
                    </span>
                  ) : null}
                </div>
                <div style={{ color: "var(--ink)", marginTop: "0.25rem" }}>
                  {r.review || "—"}
                </div>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  );
}

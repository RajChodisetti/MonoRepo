"use client";

import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { FormEvent, Suspense, useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { RESTAURANT_STATUSES, formatDate } from "@/lib/constants";
import type {
  Campaign,
  DemoLink,
  DemoSite,
  GeneratedSite,
  Member,
  ProfileReviewPreview,
  Restaurant,
} from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";
import { PhotoGallery } from "@/components/PhotoGallery";
import { SendPreviewModal } from "@/components/SendPreviewModal";

type Tab = "overview" | "photos" | "profile" | "demo" | "campaign" | "members";

function RestaurantDetailInner() {
  const params = useParams<{ id: string }>();
  const search = useSearchParams();
  const id = params.id;
  const initialTab = (search.get("tab") as Tab) || "overview";

  const [tab, setTab] = useState<Tab>(initialTab);
  const [restaurant, setRestaurant] = useState<Restaurant | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  // overview form
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [isContacted, setIsContacted] = useState(false);
  const [shownInterest, setShownInterest] = useState(false);
  const [status, setStatus] = useState("lead");

  // profile
  const [preview, setPreview] = useState<ProfileReviewPreview | null>(null);

  // demo
  const [demoId, setDemoId] = useState("");
  const [demoPreview, setDemoPreview] = useState<Record<string, unknown> | null>(
    null,
  );
  const [demoLinks, setDemoLinks] = useState<DemoLink[]>([]);
  const [demoToken, setDemoToken] = useState("");
  const [generatedSite, setGeneratedSite] = useState<GeneratedSite | null>(null);
  const [generatedSiteError, setGeneratedSiteError] = useState<string | null>(null);

  // ad hoc send
  const [sendPreviewOpen, setSendPreviewOpen] = useState(false);

  // campaigns
  const [campaigns, setCampaigns] = useState<Campaign[]>([]);
  const [selectedCampaign, setSelectedCampaign] = useState<Campaign | null>(null);

  // members
  const [members, setMembers] = useState<Member[]>([]);
  const [memberUserId, setMemberUserId] = useState("");
  const [memberRole, setMemberRole] = useState("owner");

  const loadRestaurant = useCallback(async () => {
    const data = await adminFetch<Restaurant>(`restaurants/${id}`);
    setRestaurant(data);
    setName(data.name || "");
    setEmail(data.email || "");
    setIsContacted(!!data.is_contacted);
    setShownInterest(!!data.shown_interest);
    setStatus(data.status || "lead");
  }, [id]);

  const loadProfile = useCallback(async () => {
    const data = await adminFetch<ProfileReviewPreview>(
      `restaurants/${id}/profile/review-preview`,
    );
    setPreview(data);
  }, [id]);

  const loadCampaigns = useCallback(async () => {
    const data = await adminFetch<{ items: Campaign[] }>(
      `restaurants/${id}/campaigns`,
    );
    setCampaigns(data.items || []);
  }, [id]);

  const loadDemoLinks = useCallback(async () => {
    const data = await adminFetch<{ items: DemoLink[] }>(
      `restaurants/${id}/demo-links`,
    );
    setDemoLinks(data.items || []);
  }, [id]);

  const loadGeneratedSite = useCallback(async () => {
    setGeneratedSiteError(null);
    try {
      const data = await adminFetch<GeneratedSite>(
        `restaurants/${id}/generated-site`,
      );
      setGeneratedSite(data);
    } catch (err) {
      setGeneratedSite(null);
      setGeneratedSiteError(
        err instanceof Error ? err.message : "Generated website is unavailable",
      );
    }
  }, [id]);

  const loadMembers = useCallback(async () => {
    const data = await adminFetch<{ items: Member[] }>(
      `restaurants/${id}/members`,
    );
    setMembers(data.items || []);
  }, [id]);

  useEffect(() => {
    let cancelled = false;
    async function boot() {
      setError(null);
      try {
        await loadRestaurant();
        try {
          await loadProfile();
        } catch {
          // A manually created restaurant may not have a scrape profile yet.
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load");
        }
      }
    }
    boot();
    return () => {
      cancelled = true;
    };
  }, [loadRestaurant, loadProfile]);

  useEffect(() => {
    setMessage(null);
    setError(null);
    async function loadTab() {
      try {
        if (tab === "profile") await loadProfile();
        if (tab === "campaign") await loadCampaigns();
        if (tab === "members") await loadMembers();
        if (tab === "demo") {
          await Promise.all([loadDemoLinks(), loadGeneratedSite()]);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load tab");
      }
    }
    loadTab();
  }, [
    tab,
    loadProfile,
    loadCampaigns,
    loadMembers,
    loadDemoLinks,
    loadGeneratedSite,
  ]);

  async function saveOverview(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      await adminFetch(`restaurants/${id}`, {
        method: "PATCH",
        body: {
          name,
          email,
          is_contacted: isContacted,
          shown_interest: shownInterest,
        },
      });
      if (status !== restaurant?.status) {
        await adminFetch(`restaurants/${id}/status`, {
          method: "PATCH",
          body: { status },
        });
      }
      await loadRestaurant();
      setMessage("Restaurant updated.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Update failed");
    } finally {
      setBusy(false);
    }
  }

  async function archive() {
    if (!confirm("Archive this restaurant?")) return;
    setBusy(true);
    try {
      await adminFetch(`restaurants/${id}`, { method: "DELETE" });
      await loadRestaurant();
      setMessage("Restaurant archived.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Archive failed");
    } finally {
      setBusy(false);
    }
  }

  async function reviewProfile(nextStatus: "approved" | "rejected" | "draft") {
    if (!preview) return;
    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      const body: Record<string, unknown> = { status: nextStatus };
      if (nextStatus !== "draft") {
        body.expected_restaurant_updated_at = preview.restaurant_updated_at;
        body.expected_profile_updated_at = preview.profile_updated_at;
      }
      await adminFetch(`restaurants/${id}/profile/review`, {
        method: "PATCH",
        body,
      });
      await loadProfile();
      await loadRestaurant();
      setMessage(`Profile marked ${nextStatus}.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Review failed");
    } finally {
      setBusy(false);
    }
  }

  async function createDemo() {
    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      const slugBase =
        (restaurant?.name || "demo")
          .toLowerCase()
          .replace(/[^a-z0-9]+/g, "-")
          .replace(/^-|-$/g, "")
          .slice(0, 40) || "demo";
      const slug = `${slugBase}-${Date.now().toString(36)}`;
      const res = await adminFetch<Record<string, unknown>>(
        `restaurants/${id}/demo-sites`,
        { method: "POST", body: { slug, status: "draft" } },
      );
      const newId =
        (res.demo_site_id as string) ||
        (res.id as string) ||
        ((res.demo_site as DemoSite | undefined)?.id as string);
      if (newId) {
        setDemoId(newId);
        const previewRes = await adminFetch<Record<string, unknown>>(
          `demo-sites/${newId}/review-preview`,
        );
        setDemoPreview(previewRes);
      }
      setDemoToken(typeof res.token === "string" ? res.token : "");
      await loadDemoLinks();
      setMessage(
        `Restaurant-specific demo draft created (${slug}). Its one-time token is held in this browser session so you can create a campaign draft.`,
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Demo create failed");
    } finally {
      setBusy(false);
    }
  }

  async function selectDemo(link: DemoLink) {
    setDemoId(link.demo_site_id);
    setDemoToken("");
    setDemoPreview(null);
    setBusy(true);
    setError(null);
    try {
      const previewRes = await adminFetch<Record<string, unknown>>(
        `demo-sites/${link.demo_site_id}/review-preview`,
      );
      setDemoPreview(previewRes);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to inspect demo");
    } finally {
      setBusy(false);
    }
  }

  async function loadDemoPreview() {
    if (!demoId) return;
    setBusy(true);
    setError(null);
    try {
      const previewRes = await adminFetch<Record<string, unknown>>(
        `demo-sites/${demoId}/review-preview`,
      );
      setDemoPreview(previewRes);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load demo");
    } finally {
      setBusy(false);
    }
  }

  async function setDemoStatus(next: "published" | "draft") {
    if (!demoId || !demoPreview) return;
    setBusy(true);
    setError(null);
    try {
      const updatedAt =
        (demoPreview.updated_at as string) ||
        ((demoPreview.demo_site as { updated_at?: string })?.updated_at as string);
      await adminFetch(`demo-sites/${demoId}/status`, {
        method: "PATCH",
        body: {
          status: next,
          expected_updated_at: updatedAt,
        },
      });
      await loadDemoPreview();
      await loadDemoLinks();
      setMessage(`Demo ${next}.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Demo status update failed");
    } finally {
      setBusy(false);
    }
  }

  async function createCampaign() {
    if (!demoId || !demoToken) {
      setError(
        "Create a demo draft in this browser session first. Existing demo tokens are intentionally not returned by the API.",
      );
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await adminFetch<Campaign>(`restaurants/${id}/campaigns`, {
        method: "POST",
        body: {
          demo_site_id: demoId,
          demo_token: demoToken,
          campaign_type: "outreach",
        },
      });
      await loadCampaigns();
      await loadDemoLinks();
      setSelectedCampaign(res);
      setMessage("Campaign draft created.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Campaign create failed");
    } finally {
      setBusy(false);
    }
  }

  async function openCampaign(campaignId: string) {
    setBusy(true);
    setError(null);
    try {
      const res = await adminFetch<{ campaign: Campaign }>(
        `campaigns/${campaignId}`,
      );
      setSelectedCampaign(res.campaign);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load campaign");
    } finally {
      setBusy(false);
    }
  }

  async function approveCampaign() {
    if (!selectedCampaign) return;
    setBusy(true);
    setError(null);
    try {
      const res = await adminFetch<Campaign>(
        `campaigns/${selectedCampaign.id}/approve`,
        {
          method: "POST",
          body: { expected_updated_at: selectedCampaign.updated_at },
        },
      );
      setSelectedCampaign(res);
      await loadCampaigns();
      setMessage("Campaign approved.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Approve failed");
    } finally {
      setBusy(false);
    }
  }

  async function regenerateCampaign() {
    if (!selectedCampaign) return;
    setBusy(true);
    try {
      const res = await adminFetch<Campaign>(
        `campaigns/${selectedCampaign.id}/regenerate`,
        { method: "POST", body: {} },
      );
      setSelectedCampaign(res);
      await loadCampaigns();
      setMessage("Campaign regenerated.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Regenerate failed");
    } finally {
      setBusy(false);
    }
  }

  async function stopCampaign() {
    if (!selectedCampaign) return;
    setBusy(true);
    try {
      const res = await adminFetch<Campaign>(
        `campaigns/${selectedCampaign.id}/stop`,
        { method: "POST", body: {} },
      );
      setSelectedCampaign(res);
      await loadCampaigns();
      setMessage("Campaign stopped.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Stop failed");
    } finally {
      setBusy(false);
    }
  }

  async function addMember(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await adminFetch(`restaurants/${id}/members`, {
        method: "POST",
        body: {
          user_id: memberUserId,
          member_role: memberRole || "owner",
        },
      });
      setMemberUserId("");
      setMemberRole("owner");
      await loadMembers();
      setMessage("Member assigned.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add member");
    } finally {
      setBusy(false);
    }
  }

  if (!restaurant && !error) {
    return <EmptyState message="Loading restaurant…" />;
  }

  const tabs: { id: Tab; label: string }[] = [
    { id: "overview", label: "Overview" },
    { id: "photos", label: "Photos" },
    { id: "profile", label: "Profile review" },
    { id: "demo", label: "Demo" },
    { id: "campaign", label: "Campaign" },
    { id: "members", label: "Members" },
  ];
  const ocrStatus = preview?.ocr_status || "unknown";
  const ocrChecked = Boolean(preview?.ocr_checked || (preview?.ocr_attempts || 0) > 0);
  const ocrStateLabel =
    ocrStatus === "running"
      ? "OCR is checking this restaurant now"
      : ocrStatus === "pending" && !ocrChecked
        ? "Not checked by OCR yet"
        : ocrStatus === "verified"
          ? "Checked and verified"
          : ocrStatus === "no_images"
            ? "Checked — no usable images found"
            : ocrStatus === "failed"
              ? "Checked — OCR failed"
              : ocrChecked
                ? `Checked — ${ocrStatus}`
                : "OCR state unavailable";
  const ocrErrors = Array.isArray(preview?.ocr_verification_errors)
    ? preview.ocr_verification_errors.map(String)
    : [];

  return (
    <div>
      <PageHeader
        title={restaurant?.name || "Restaurant"}
        subtitle={restaurant?.email || restaurant?.id}
        actions={
          <>
            <button
              className="btn btn-primary"
              type="button"
              onClick={() => setSendPreviewOpen(true)}
              disabled={!restaurant}
            >
              Send email
            </button>
            <Link className="btn btn-secondary" href="/restaurants">
              All restaurants
            </Link>
          </>
        }
      />
      <SendPreviewModal
        open={sendPreviewOpen}
        restaurantIds={restaurant ? [restaurant.id] : []}
        onClose={() => setSendPreviewOpen(false)}
        onSent={loadRestaurant}
      />
      <ErrorBanner message={error} />
      {message ? (
        <div className="alert alert-info" style={{ marginBottom: "1rem" }}>
          {message}
        </div>
      ) : null}

      <div className="tabs">
        {tabs.map((t) => (
          <button
            key={t.id}
            type="button"
            className="tab"
            data-active={tab === t.id}
            onClick={() => setTab(t.id)}
            style={{
              background: "transparent",
              border: "none",
              cursor: "pointer",
            }}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "overview" && restaurant ? (
        <form onSubmit={saveOverview} className="card" style={{ display: "grid", gap: "0.85rem", maxWidth: 640 }}>
          <div style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
            <StatusBadge status={restaurant.status} />
            <span style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
              Updated {formatDate(restaurant.updated_at)}
            </span>
          </div>
          <label style={{ display: "grid", gap: "0.35rem" }}>
            <span style={{ fontWeight: 600, fontSize: "0.85rem" }}>Name</span>
            <input className="input" value={name} onChange={(e) => setName(e.target.value)} />
          </label>
          <label style={{ display: "grid", gap: "0.35rem" }}>
            <span style={{ fontWeight: 600, fontSize: "0.85rem" }}>Email</span>
            <input className="input" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
          </label>
          <label style={{ display: "grid", gap: "0.35rem" }}>
            <span style={{ fontWeight: 600, fontSize: "0.85rem" }}>Status</span>
            <select className="select" value={status} onChange={(e) => setStatus(e.target.value)}>
              {RESTAURANT_STATUSES.map((s) => (
                <option key={s} value={s}>
                  {s}
                </option>
              ))}
            </select>
          </label>
          <label style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
            <input type="checkbox" checked={isContacted} onChange={(e) => setIsContacted(e.target.checked)} />
            Contacted
          </label>
          <label style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
            <input type="checkbox" checked={shownInterest} onChange={(e) => setShownInterest(e.target.checked)} />
            Shown interest
          </label>
          <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
            <button className="btn btn-primary" type="submit" disabled={busy}>
              Save changes
            </button>
            <button className="btn btn-danger" type="button" onClick={archive} disabled={busy}>
              Archive
            </button>
          </div>
          <div style={{ color: "var(--muted)", fontSize: "0.9rem" }}>
            Email sent: {restaurant.email_sent ? "Yes" : "No"} · send count:{" "}
            {restaurant.email_send_count ?? 0} · last:{" "}
            {formatDate(restaurant.last_email_sent_at)}
          </div>
          <div className="alert alert-info" style={{ margin: 0 }}>
            <strong>OCR:</strong> {ocrStateLabel} · status {ocrStatus} · attempts{" "}
            {preview?.ocr_attempts ?? 0}
            {(preview?.ocr_images_discovered ?? 0) > 0
              ? ` · photos ${preview?.ocr_images_succeeded ?? 0}/${preview?.ocr_images_discovered ?? 0} successful`
              : ""}
            {preview?.ocr_completed_at
              ? ` · completed ${formatDate(preview.ocr_completed_at)}`
              : ""}
          </div>
        </form>
      ) : null}

      {tab === "photos" ? (
        <div className="card">
          <PhotoGallery restaurantId={id} />
        </div>
      ) : null}

      {tab === "profile" ? (
        <div className="card" style={{ display: "grid", gap: "0.85rem" }}>
          {!preview ? (
            <p style={{ color: "var(--muted)", margin: 0 }}>Loading preview…</p>
          ) : (
            <>
              <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                <StatusBadge status={preview.review_status || "draft"} />
                <StatusBadge status={preview.ocr_status || "ocr"} />
              </div>
              <div className="alert alert-info" style={{ margin: 0 }}>
                <strong>{ocrStateLabel}</strong>
                <div style={{ marginTop: "0.3rem", fontSize: "0.85rem" }}>
                  Attempts: {preview.ocr_attempts ?? 0} · started:{" "}
                  {formatDate(preview.ocr_started_at)} · completed:{" "}
                  {formatDate(preview.ocr_completed_at)}
                </div>
                <div style={{ marginTop: "0.3rem", fontSize: "0.85rem" }}>
                  Photos discovered: {preview.ocr_images_discovered ?? 0} · analyzed:{" "}
                  {preview.ocr_images_analyzed ?? 0} · successful:{" "}
                  {preview.ocr_images_succeeded ?? 0} · failed:{" "}
                  {preview.ocr_images_failed ?? 0} · all processed:{" "}
                  {preview.ocr_all_images_processed ? "yes" : "no"}
                </div>
                {preview.ocr_model ? (
                  <div style={{ marginTop: "0.3rem", fontSize: "0.85rem" }}>
                    Model: <code>{preview.ocr_model}</code>
                    {preview.ocr_provider ? ` · provider: ${preview.ocr_provider}` : ""}
                  </div>
                ) : null}
                {ocrErrors.length > 0 ? (
                  <ul style={{ margin: "0.45rem 0 0", paddingLeft: "1.2rem" }}>
                    {ocrErrors.map((ocrError, index) => (
                      <li key={`${ocrError}-${index}`}>{ocrError}</li>
                    ))}
                  </ul>
                ) : null}
              </div>
              <div>
                <strong>Contact:</strong> {preview.contact_email || "—"}
              </div>
              <pre
                style={{
                  margin: 0,
                  padding: "0.85rem",
                  background: "var(--bg)",
                  border: "1px solid var(--line)",
                  overflow: "auto",
                  maxHeight: 360,
                  fontSize: "0.82rem",
                }}
              >
                {JSON.stringify(preview.profile ?? preview, null, 2)}
              </pre>
              <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                <button className="btn btn-primary" type="button" disabled={busy} onClick={() => reviewProfile("approved")}>
                  Approve profile
                </button>
                <button className="btn btn-danger" type="button" disabled={busy} onClick={() => reviewProfile("rejected")}>
                  Reject
                </button>
                <button className="btn btn-secondary" type="button" disabled={busy} onClick={() => reviewProfile("draft")}>
                  Reset to draft
                </button>
                <button className="btn btn-secondary" type="button" disabled={busy} onClick={loadProfile}>
                  Refresh preview
                </button>
              </div>
            </>
          )}
        </div>
      ) : null}

      {tab === "demo" ? (
        <div style={{ display: "grid", gap: "1rem" }}>
          <div className="card" style={{ display: "grid", gap: "0.75rem" }}>
            <div>
              <h3 style={{ margin: 0, fontSize: "1rem" }}>Generated restaurant website</h3>
              <p style={{ color: "var(--muted)", margin: "0.35rem 0 0", fontSize: "0.86rem" }}>
                This is the existing database-driven website generator mapped from restaurant UUID{" "}
                <code>{id}</code> to its current site index. It is an internal preview, not the
                token-gated link to send a restaurant.
              </p>
            </div>
            {generatedSiteError ? <ErrorBanner message={generatedSiteError} /> : null}
            {generatedSite ? (
              <>
                <div style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
                  Generator index: {generatedSite.site_index} · Google Place ID:{" "}
                  <code>{generatedSite.google_place_id}</code>
                </div>
                <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                  {generatedSite.templates.map((template) => (
                    <a
                      className="btn btn-primary"
                      href={template.url}
                      target="_blank"
                      rel="noreferrer"
                      key={template.id}
                    >
                      Open {template.name}
                    </a>
                  ))}
                </div>
              </>
            ) : !generatedSiteError ? (
              <p style={{ color: "var(--muted)", margin: 0 }}>Loading generated website links…</p>
            ) : null}
          </div>

          <div className="card" style={{ display: "grid", gap: "0.65rem" }}>
            <h3 style={{ margin: 0, fontSize: "1rem" }}>What the demo controls do</h3>
            <ol style={{ margin: 0, paddingLeft: "1.2rem", color: "var(--muted)", lineHeight: 1.65 }}>
              <li>
                <strong style={{ color: "var(--ink)" }}>Create demo draft</strong> snapshots this
                restaurant&apos;s public-safe profile/menu data and creates a one-time access token.
              </li>
              <li>
                <strong style={{ color: "var(--ink)" }}>Inspect payload</strong> shows the exact
                server-side data the published website will receive; it does not publish anything.
              </li>
              <li>
                <strong style={{ color: "var(--ink)" }}>Publish</strong> makes the token-gated
                website available, but only after OCR is verified and the profile is approved.
              </li>
              <li>
                <strong style={{ color: "var(--ink)" }}>Unpublish</strong> returns it to draft and
                immediately revokes public access without deleting the snapshot.
              </li>
            </ol>
          </div>

          <div className="card" style={{ display: "grid", gap: "0.6rem" }}>
            <h3 style={{ margin: 0, fontSize: "0.95rem" }}>Token-gated demo records</h3>
            {demoLinks.length === 0 ? (
              <p style={{ color: "var(--muted)", margin: 0 }}>
                No demo snapshot exists yet. OCR automatically creates one after a successful check,
                or you can create one manually below.
              </p>
            ) : (
              <div className="table-wrap">
                <table className="data">
                  <thead>
                    <tr>
                      <th>Slug</th>
                      <th>Status</th>
                      <th>Created</th>
                      <th>Published website</th>
                      <th>Manage</th>
                    </tr>
                  </thead>
                  <tbody>
                    {demoLinks.map((link) => (
                      <tr key={link.demo_site_id}>
                        <td>{link.slug}</td>
                        <td>
                          <StatusBadge status={link.status} />
                        </td>
                        <td>{formatDate(link.created_at)}</td>
                        <td>
                          {link.preview_url && link.status === "published" ? (
                            <a href={link.preview_url} target="_blank" rel="noreferrer">
                              Open published demo
                            </a>
                          ) : (
                            <span style={{ color: "var(--muted)" }}>
                              {link.status === "published"
                                ? "Create a campaign to retain the shareable token"
                                : "Unavailable while draft"}
                            </span>
                          )}
                        </td>
                        <td>
                          <button
                            type="button"
                            className="btn btn-secondary"
                            onClick={() => selectDemo(link)}
                            disabled={busy}
                          >
                            Inspect payload
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
          <div className="card" style={{ display: "grid", gap: "0.85rem" }}>
            <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap", alignItems: "center" }}>
              <button className="btn btn-primary" type="button" disabled={busy} onClick={createDemo}>
                Create restaurant demo draft
              </button>
              <button className="btn btn-secondary" type="button" disabled={busy || !demoId} onClick={loadDemoPreview}>
                Refresh payload
              </button>
              <button className="btn btn-primary" type="button" disabled={busy || !demoId || !demoPreview} onClick={() => setDemoStatus("published")}>
                Publish
              </button>
              <button className="btn btn-secondary" type="button" disabled={busy || !demoId || !demoPreview} onClick={() => setDemoStatus("draft")}>
                Unpublish
              </button>
            </div>
            {demoId ? (
              <div style={{ color: "var(--muted)", fontSize: "0.82rem" }}>
                Selected demo: <code>{demoId}</code>
              </div>
            ) : null}
            {demoToken ? (
              <div className="alert alert-info" style={{ margin: 0 }}>
                The new demo&apos;s one-time token is held only in this page session. Create a campaign
                draft before leaving if you need a retained shareable link.
              </div>
            ) : null}
            {demoPreview ? (
              <div>
                <h4 style={{ margin: "0 0 0.45rem", fontSize: "0.9rem" }}>Reviewed public payload</h4>
                <pre
                  style={{
                    margin: 0,
                    padding: "0.85rem",
                    background: "var(--bg)",
                    border: "1px solid var(--line)",
                    overflow: "auto",
                    maxHeight: 420,
                    fontSize: "0.82rem",
                  }}
                >
                  {JSON.stringify(demoPreview, null, 2)}
                </pre>
              </div>
            ) : (
              <p style={{ color: "var(--muted)", margin: 0 }}>
                Choose “Inspect payload” on an existing record, or create a new restaurant snapshot.
              </p>
            )}
          </div>
        </div>
      ) : null}

      {tab === "campaign" ? (
        <div style={{ display: "grid", gap: "1rem" }}>
          <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
            <button
              className="btn btn-primary"
              type="button"
              disabled={busy || !demoId || !demoToken}
              onClick={createCampaign}
            >
              Create campaign draft
            </button>
            <button className="btn btn-secondary" type="button" disabled={busy} onClick={loadCampaigns}>
              Refresh list
            </button>
          </div>
          {!demoToken ? (
            <p style={{ color: "var(--muted)", margin: 0, fontSize: "0.86rem" }}>
              Create a new demo draft in the Demo tab first. Demo access tokens are returned only
              once and are intentionally not recoverable from existing records.
            </p>
          ) : null}
          <div className="table-wrap">
            <table className="data">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Status</th>
                  <th>Subject</th>
                  <th>Updated</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {campaigns.map((c) => (
                  <tr key={c.id}>
                    <td style={{ fontFamily: "monospace", fontSize: "0.8rem" }}>
                      {c.id.slice(0, 8)}…
                    </td>
                    <td>
                      <StatusBadge status={c.status} />
                    </td>
                    <td>{c.subject || "—"}</td>
                    <td>{formatDate(c.updated_at)}</td>
                    <td>
                      <button
                        type="button"
                        className="btn btn-secondary"
                        onClick={() => openCampaign(c.id)}
                      >
                        Inspect
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {campaigns.length === 0 ? (
            <EmptyState message="No campaigns yet for this restaurant." />
          ) : null}
          {selectedCampaign ? (
            <div className="card" style={{ display: "grid", gap: "0.75rem" }}>
              <div style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
                <strong>Selected campaign</strong>
                <StatusBadge status={selectedCampaign.status} />
              </div>
              <div>
                <strong>Subject:</strong> {selectedCampaign.subject || "—"}
              </div>
              <div
                style={{
                  border: "1px solid var(--line)",
                  padding: "0.85rem",
                  background: "var(--bg)",
                  maxHeight: 280,
                  overflow: "auto",
                }}
                dangerouslySetInnerHTML={{
                  __html: selectedCampaign.body_html || "<em>No HTML body</em>",
                }}
              />
              <pre
                style={{
                  margin: 0,
                  whiteSpace: "pre-wrap",
                  fontSize: "0.82rem",
                  maxHeight: 160,
                  overflow: "auto",
                }}
              >
                {selectedCampaign.body_text || ""}
              </pre>
              <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                <button className="btn btn-primary" type="button" disabled={busy} onClick={approveCampaign}>
                  Approve
                </button>
                <button className="btn btn-secondary" type="button" disabled={busy} onClick={regenerateCampaign}>
                  Regenerate
                </button>
                <button className="btn btn-danger" type="button" disabled={busy} onClick={stopCampaign}>
                  Stop
                </button>
              </div>
            </div>
          ) : null}
        </div>
      ) : null}

      {tab === "members" ? (
        <div style={{ display: "grid", gap: "1rem" }}>
          <form
            onSubmit={addMember}
            className="card"
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))",
              gap: "0.75rem",
              alignItems: "end",
            }}
          >
            <label style={{ display: "grid", gap: "0.35rem" }}>
              <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>
                User ID (UUID)
              </span>
              <input
                className="input"
                required
                value={memberUserId}
                onChange={(e) => setMemberUserId(e.target.value)}
                placeholder="existing user uuid"
              />
            </label>
            <label style={{ display: "grid", gap: "0.35rem" }}>
              <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>
                Member role
              </span>
              <select
                className="select"
                value={memberRole}
                onChange={(e) => setMemberRole(e.target.value)}
              >
                <option value="owner">owner</option>
              </select>
            </label>
            <button className="btn btn-primary" type="submit" disabled={busy}>
              Assign member
            </button>
          </form>
          <p style={{ color: "var(--muted)", margin: 0, fontSize: "0.9rem" }}>
            Members API assigns an existing user by <code>user_id</code> (create
            users via auth signup as restaurant_owner first).
          </p>
          <div className="table-wrap">
            <table className="data">
              <thead>
                <tr>
                  <th>Member ID</th>
                  <th>User ID</th>
                  <th>Role</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {members.map((m, i) => (
                  <tr key={m.id || m.user_id || i}>
                    <td style={{ fontFamily: "monospace", fontSize: "0.8rem" }}>
                      {m.id || "—"}
                    </td>
                    <td style={{ fontFamily: "monospace", fontSize: "0.8rem" }}>
                      {m.user_id || "—"}
                    </td>
                    <td>{(m as Member & { member_role?: string }).member_role || m.role || "—"}</td>
                    <td>{formatDate(m.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {members.length === 0 ? (
            <EmptyState message="No members assigned yet." />
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

export default function RestaurantDetailPage() {
  return (
    <Suspense fallback={<EmptyState message="Loading…" />}>
      <RestaurantDetailInner />
    </Suspense>
  );
}

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
  }, [loadRestaurant]);

  useEffect(() => {
    setMessage(null);
    setError(null);
    async function loadTab() {
      try {
        if (tab === "profile") await loadProfile();
        if (tab === "campaign") await loadCampaigns();
        if (tab === "members") await loadMembers();
        if (tab === "demo") await loadDemoLinks();
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load tab");
      }
    }
    loadTab();
  }, [tab, loadProfile, loadCampaigns, loadMembers, loadDemoLinks]);

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
      const token = res.token ? ` Token (save once): ${String(res.token)}` : "";
      setMessage(`Demo draft created (${slug}).${token}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Demo create failed");
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
      setMessage(`Demo ${next}.`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Demo status update failed");
    } finally {
      setBusy(false);
    }
  }

  async function createCampaign() {
    setBusy(true);
    setError(null);
    try {
      const res = await adminFetch<Campaign>(`restaurants/${id}/campaigns`, {
        method: "POST",
        body: { campaign_type: "outreach" },
      });
      await loadCampaigns();
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
          <div className="card" style={{ display: "grid", gap: "0.6rem" }}>
            <h3 style={{ margin: 0, fontSize: "0.95rem" }}>Demo links</h3>
            {demoLinks.length === 0 ? (
              <p style={{ color: "var(--muted)", margin: 0 }}>
                No demo sites created yet for this lead.
              </p>
            ) : (
              <div className="table-wrap">
                <table className="data">
                  <thead>
                    <tr>
                      <th>Slug</th>
                      <th>Status</th>
                      <th>Created</th>
                      <th>Link</th>
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
                          {link.preview_url ? (
                            <a href={link.preview_url} target="_blank" rel="noreferrer">
                              Open demo
                            </a>
                          ) : (
                            <span style={{ color: "var(--muted)" }}>
                              No link yet (create a campaign for this demo)
                            </span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
          <div className="card" style={{ display: "grid", gap: "0.85rem" }}>
          <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap", alignItems: "end" }}>
            <button className="btn btn-primary" type="button" disabled={busy} onClick={createDemo}>
              Create demo draft
            </button>
            <label style={{ display: "grid", gap: "0.35rem", minWidth: 280 }}>
              <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Demo site ID</span>
              <input className="input" value={demoId} onChange={(e) => setDemoId(e.target.value)} placeholder="uuid" />
            </label>
            <button className="btn btn-secondary" type="button" disabled={busy || !demoId} onClick={loadDemoPreview}>
              Load preview
            </button>
            <button className="btn btn-primary" type="button" disabled={busy || !demoId || !demoPreview} onClick={() => setDemoStatus("published")}>
              Publish
            </button>
            <button className="btn btn-secondary" type="button" disabled={busy || !demoId || !demoPreview} onClick={() => setDemoStatus("draft")}>
              Unpublish
            </button>
          </div>
          {demoPreview ? (
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
          ) : (
            <p style={{ color: "var(--muted)", margin: 0 }}>
              Create a demo or paste an existing demo site id to review/publish.
            </p>
          )}
          </div>
        </div>
      ) : null}

      {tab === "campaign" ? (
        <div style={{ display: "grid", gap: "1rem" }}>
          <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
            <button className="btn btn-primary" type="button" disabled={busy} onClick={createCampaign}>
              Create campaign draft
            </button>
            <button className="btn btn-secondary" type="button" disabled={busy} onClick={loadCampaigns}>
              Refresh list
            </button>
          </div>
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

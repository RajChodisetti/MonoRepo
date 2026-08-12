"use client";

import Link from "next/link";
import { useParams, useSearchParams } from "next/navigation";
import { FormEvent, Suspense, useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { RESTAURANT_STATUSES, formatDate } from "@/lib/constants";
import type {
  DemoLink,
  DemoSession,
  DemoSite,
  GeneratedSite,
  Member,
  ProfileReviewPreview,
  Restaurant,
} from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";
import { PhotoGallery } from "@/components/PhotoGallery";

type Tab = "overview" | "photos" | "profile" | "demo" | "engagement" | "members";

function formatDuration(seconds: number) {
  const safe = Math.max(0, Math.round(seconds || 0));
  const minutes = Math.floor(safe / 60);
  const remainder = safe % 60;
  return minutes > 0 ? `${minutes}m ${remainder}s` : `${remainder}s`;
}

function templateName(templateID: DemoSession["template_id"]) {
  if (templateID === "2") return "Aurora";
  if (templateID === "3") return "Elysian";
  return "Cinematic";
}

function apolloStatusMessage(status?: string, emailFound?: boolean) {
  if (emailFound) return "Apollo found a work email for this restaurant.";
  switch (status) {
    case "skipped_no_domain":
      return "Apollo could not run because the restaurant has no usable website or company domain.";
    case "no_candidate":
      return "Apollo searched the available domain but found no suitable owner/manager candidate.";
    case "no_match":
      return "Apollo found candidates, but none produced a verified work email match.";
    case "enriched":
      return "Apollo enriched the restaurant, but did not return a usable work email.";
    case "not_recorded":
      return "No Apollo enrichment result has been recorded for this restaurant yet.";
    default:
      return status ? `Apollo result: ${status}. No usable email was returned.` : "Apollo status is unavailable.";
  }
}

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
  const [status, setStatus] = useState("lead");

  // profile
  const [preview, setPreview] = useState<ProfileReviewPreview | null>(null);

  // demo
  const [demoId, setDemoId] = useState("");
  const [demoPreview, setDemoPreview] = useState<Record<string, unknown> | null>(
    null,
  );
  const [demoLinks, setDemoLinks] = useState<DemoLink[]>([]);
  const [generatedSite, setGeneratedSite] = useState<GeneratedSite | null>(null);
  const [generatedSiteError, setGeneratedSiteError] = useState<string | null>(null);

  // demo engagement
  const [demoSessions, setDemoSessions] = useState<DemoSession[]>([]);

  // members
  const [members, setMembers] = useState<Member[]>([]);
  const [memberUserId, setMemberUserId] = useState("");
  const [memberRole, setMemberRole] = useState("owner");
  const stableTemplates = generatedSite?.templates || [];

  const loadRestaurant = useCallback(async () => {
    const data = await adminFetch<Restaurant>(`restaurants/${id}`);
    setRestaurant(data);
    setName(data.name || "");
    setEmail(data.email || "");
    setStatus(data.status || "lead");
  }, [id]);

  const loadProfile = useCallback(async () => {
    const data = await adminFetch<ProfileReviewPreview>(
      `restaurants/${id}/profile/review-preview`,
    );
    setPreview(data);
  }, [id]);

  const loadEngagement = useCallback(async () => {
    const data = await adminFetch<{ items: DemoSession[] }>(
      `restaurants/${id}/demo-engagement`,
    );
    setDemoSessions(data.items || []);
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
        if (tab === "engagement") await loadEngagement();
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
    loadEngagement,
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
      await loadDemoLinks();
      setMessage(`Restaurant-specific demo draft created (${slug}).`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Demo create failed");
    } finally {
      setBusy(false);
    }
  }

  async function selectDemo(link: DemoLink) {
    setDemoId(link.demo_site_id);
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

  async function viewPersonalizedWebsite(
    template: NonNullable<typeof generatedSite>["templates"][number],
  ) {
    const previewWindow = window.open("about:blank", "_blank");
    if (!previewWindow) {
      setError("Allow pop-ups for the admin portal to view the personalized website.");
      return;
    }
    previewWindow.opener = null;
    setBusy(true);
    setError(null);
    try {
      const capability = await adminFetch<{
        session_id: string;
        session_token: string;
      }>(`restaurants/${id}/demo-engagement/preview`, {
        method: "POST",
        body: { template_id: template.id },
      });
      const url = new URL(template.url);
      url.hash = new URLSearchParams({
        engagement_session: capability.session_id,
        engagement_token: capability.session_token,
      }).toString();
      previewWindow.location.replace(url.toString());
      setMessage(`${template.name} personalized website opened; engagement tracking started.`);
      await loadEngagement();
    } catch (err) {
		previewWindow.close();
		setError(err instanceof Error ? err.message : "Personalized website could not be opened");
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
    { id: "engagement", label: "Engagement" },
    { id: "members", label: "Members" },
  ];

  return (
    <div>
      <PageHeader
        title={restaurant?.name || "Restaurant"}
        subtitle={restaurant?.email || restaurant?.id}
        actions={
          <>
            <Link className="btn btn-primary" href="/outreach">
              View outreach progress
            </Link>
            <Link className="btn btn-secondary" href="/restaurants">
              All restaurants
            </Link>
          </>
        }
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
            <span style={{ fontWeight: 600, fontSize: "0.85rem" }}>Phone</span>
            <input className="input" value={restaurant.phone || ""} readOnly disabled />
          </label>
          <label style={{ display: "grid", gap: "0.35rem" }}>
            <span style={{ fontWeight: 600, fontSize: "0.85rem" }}>Address</span>
            <textarea className="input" value={restaurant.address || ""} readOnly disabled rows={2} />
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
            <input type="checkbox" checked={!!restaurant.is_contacted} readOnly disabled />
            Contacted
          </label>
          <label style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
            <input type="checkbox" checked={!!restaurant.shown_interest} readOnly disabled />
            Shown interest
          </label>
          <div style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
            Contacted is set after Gmail confirms a send. When an owner responds or expresses
            interest, mark the lifecycle accordingly so automated follow-ups pause for a person.
          </div>
          <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
            <button className="btn btn-primary" type="submit" disabled={busy}>
              Save changes
            </button>
            <button className="btn btn-danger" type="button" onClick={archive} disabled={busy}>
              Archive
            </button>
          </div>
          <div style={{ color: "var(--muted)", fontSize: "0.9rem" }}>
            Confirmed sequence sends: {restaurant.email_send_count ?? 0} · last:{" "}
            {formatDate(restaurant.last_email_sent_at)}
          </div>
          <div className="alert alert-info" style={{ margin: 0 }}>
            Outreach eligibility uses restaurant name, valid email, recorded business-consent
            evidence, and lifecycle. Profile approval is managed separately.
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
              </div>
              <div>
                <strong>Contact:</strong> {preview.contact_email || "—"}
              </div>
              <div className="alert alert-info" style={{ margin: 0 }}>
                <strong>Apollo:</strong> {apolloStatusMessage(preview.apollo_status, preview.apollo_email_found)}
                {preview.apollo_status ? (
                  <div style={{ marginTop: "0.25rem", fontSize: "0.82rem" }}>
                    Provider result: <code>{preview.apollo_status}</code>
                  </div>
                ) : null}
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
                <div style={{ display: "grid", gap: "0.75rem" }}>
                  <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
                    {stableTemplates.map((template) => (
                      <button
                        className="btn btn-primary"
                        type="button"
                        onClick={() => viewPersonalizedWebsite(template)}
                        disabled={busy}
                        key={template.id}
                      >
                        View {template.name} personalized website
                      </button>
                    ))}
                  </div>
                </div>
              </>
            ) : !generatedSiteError ? (
              <p style={{ color: "var(--muted)", margin: 0 }}>Loading generated website links…</p>
            ) : null}
          </div>

          <div className="card" style={{ display: "grid", gap: "0.65rem" }}>
            <h3 style={{ margin: 0, fontSize: "1rem" }}>How personalized website previews work</h3>
            <ol style={{ margin: 0, paddingLeft: "1.2rem", color: "var(--muted)", lineHeight: 1.65 }}>
              <li>
                <strong style={{ color: "var(--ink)" }}>View personalized website</strong> opens the
                selected Cinematic, Aurora, or Elysian restaurant website and starts an admin-preview session clock.
              </li>
              <li>
                <strong style={{ color: "var(--ink)" }}>Inspect payload</strong> shows the exact
                server-side public-safe data used by approved outreach; it does not expose a public publish control.
              </li>
            </ol>
          </div>

          <div className="card" style={{ display: "grid", gap: "0.6rem" }}>
            <h3 style={{ margin: 0, fontSize: "0.95rem" }}>Token-gated demo records</h3>
            {demoLinks.length === 0 ? (
              <p style={{ color: "var(--muted)", margin: 0 }}>
                No demo snapshot exists yet. Create one manually below when a reviewed public
                payload is needed.
              </p>
            ) : (
              <div className="table-wrap">
                <table className="data">
                  <thead>
                    <tr>
                      <th>Slug</th>
                      <th>Status</th>
                      <th>Created</th>
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
            </div>
            {demoId ? (
              <div style={{ color: "var(--muted)", fontSize: "0.82rem" }}>
                Selected demo: <code>{demoId}</code>
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

      {tab === "engagement" ? (
        <div style={{ display: "grid", gap: "1rem" }}>
          <div style={{ display: "flex", justifyContent: "space-between", gap: "0.75rem", alignItems: "center" }}>
            <p style={{ color: "var(--muted)", margin: 0 }}>
              Personalized-site and admin-preview visits, selected template, active time, and AI
              receptionist transcript turns for this restaurant.
            </p>
            <button className="btn btn-secondary" type="button" onClick={loadEngagement}>
              Refresh
            </button>
          </div>
          {demoSessions.length === 0 ? (
            <EmptyState message="No personalized website engagement has been recorded yet." />
          ) : (
            demoSessions.map((session) => (
              <div className="card" key={session.id} style={{ display: "grid", gap: "0.7rem" }}>
                <div style={{ display: "flex", gap: "1rem", flexWrap: "wrap", alignItems: "center" }}>
                  <strong>{formatDate(session.started_at)}</strong>
                  <StatusBadge status={templateName(session.template_id)} />
                  <span>Time on demo: {formatDuration(session.duration_seconds)}</span>
                  <span style={{ color: "var(--muted)" }}>
                    Last seen {formatDate(session.last_seen_at)}
                  </span>
                  <StatusBadge status={session.ended_at ? "ended" : "active"} />
                </div>
                {session.transcript.length === 0 ? (
                  <p style={{ color: "var(--muted)", margin: 0 }}>No AI receptionist transcript in this visit.</p>
                ) : (
                  <div style={{ display: "grid", gap: "0.45rem" }}>
                    {session.transcript.map((turn) => (
                      <div key={turn.id} style={{ borderLeft: "3px solid var(--line)", paddingLeft: "0.7rem" }}>
                        <div style={{ display: "flex", gap: "0.5rem", alignItems: "baseline" }}>
                          <strong style={{ textTransform: "capitalize" }}>{turn.role}</strong>
                          <span style={{ color: "var(--muted)", fontSize: "0.78rem" }}>
                            {formatDate(turn.occurred_at)}
                          </span>
                        </div>
                        <div style={{ whiteSpace: "pre-wrap", marginTop: "0.15rem" }}>{turn.content}</div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            ))
          )}
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

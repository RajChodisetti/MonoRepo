"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { RESTAURANT_STATUSES, formatDate } from "@/lib/constants";
import type {
  Restaurant,
  SharedEmailGroup,
  SharedEmailGroupListResponse,
} from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";

export default function RestaurantsPage() {
  const router = useRouter();
  const [items, setItems] = useState<Restaurant[]>([]);
  const [name, setName] = useState("");
  const [status, setStatus] = useState("");
  const [isContacted, setIsContacted] = useState("");
  const [shownInterest, setShownInterest] = useState("");
  const [includeArchived, setIncludeArchived] = useState(false);
  const [showSharedEmails, setShowSharedEmails] = useState(false);
  const [sharedEmailGroups, setSharedEmailGroups] = useState<SharedEmailGroup[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await adminFetch<{ items: Restaurant[] }>("restaurants", {
        query: {
          restaurant: name || undefined,
          status: status || undefined,
          is_contacted: isContacted || undefined,
          shown_interest: shownInterest || undefined,
          include_archived: includeArchived ? true : undefined,
        },
      });
      setItems(data.items || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load restaurants");
    } finally {
      setLoading(false);
    }
  }, [name, status, isContacted, shownInterest, includeArchived]);

  const loadSharedEmails = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await adminFetch<SharedEmailGroupListResponse>(
        "restaurants/shared-emails",
        { query: { limit: 100 } },
      );
      setSharedEmailGroups(data.groups || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load shared emails");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    const t = setTimeout(showSharedEmails ? loadSharedEmails : load, 200);
    return () => clearTimeout(t);
  }, [load, loadSharedEmails, showSharedEmails]);

  return (
    <div>
      <PageHeader
        title="Restaurants"
        subtitle="Filter and open leads for review and outreach"
      />
      <ErrorBanner message={error} />

      <div className="tabs" style={{ marginBottom: "1rem" }}>
        <button type="button" className="tab" data-active={!showSharedEmails} onClick={() => setShowSharedEmails(false)}>
          All restaurants
        </button>
        <button type="button" className="tab" data-active={showSharedEmails} onClick={() => setShowSharedEmails(true)}>
          Shared emails
        </button>
      </div>

      {!showSharedEmails ? <div
        className="card"
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(140px, 1fr))",
          gap: "0.75rem",
          marginBottom: "1rem",
          alignItems: "end",
        }}
      >
        <label style={{ display: "grid", gap: "0.35rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Name</span>
          <input
            className="input"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Search…"
          />
        </label>
        <label style={{ display: "grid", gap: "0.35rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Status</span>
          <select
            className="select"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            <option value="">All</option>
            {RESTAURANT_STATUSES.map((s) => (
              <option key={s} value={s}>
                {s}
              </option>
            ))}
          </select>
        </label>
        <label style={{ display: "grid", gap: "0.35rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Contacted</span>
          <select
            className="select"
            value={isContacted}
            onChange={(e) => setIsContacted(e.target.value)}
          >
            <option value="">Any</option>
            <option value="true">Yes</option>
            <option value="false">No</option>
          </select>
        </label>
        <label style={{ display: "grid", gap: "0.35rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Interest</span>
          <select
            className="select"
            value={shownInterest}
            onChange={(e) => setShownInterest(e.target.value)}
          >
            <option value="">Any</option>
            <option value="true">Yes</option>
            <option value="false">No</option>
          </select>
        </label>
        <label
          style={{
            display: "flex",
            alignItems: "center",
            gap: "0.45rem",
            paddingBottom: "0.55rem",
          }}
        >
          <input
            type="checkbox"
            checked={includeArchived}
            onChange={(e) => setIncludeArchived(e.target.checked)}
          />
          <span style={{ fontSize: "0.9rem" }}>Include archived</span>
        </label>
        <button className="btn btn-secondary" type="button" onClick={load}>
          Refresh
        </button>
      </div> : (
        <div className="card" style={{ marginBottom: "1rem" }}>
          <strong>Multi-restaurant contacts</strong>
          <p style={{ color: "var(--muted)", margin: "0.35rem 0 0" }}>
            Grouped by normalized email. Outreach is blocked when one address belongs to more than
            three restaurant records.
          </p>
        </div>
      )}

      {loading ? <EmptyState message="Loading restaurants…" /> : null}
      {!loading && !showSharedEmails && items.length === 0 ? (
        <EmptyState message="No restaurants match these filters." />
      ) : null}
      {!loading && showSharedEmails && sharedEmailGroups.length === 0 ? (
        <EmptyState message="No email address is shared by multiple restaurants." />
      ) : null}

      {!showSharedEmails && items.length > 0 ? (
        <div className="table-wrap">
          <table className="data">
            <thead>
              <tr>
                <th>Name</th>
                <th>Email</th>
                <th>Status</th>
                <th>Contacted</th>
                <th>Interest</th>
                <th>Sequence sends</th>
                <th>Last sent</th>
                <th>Updated</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {items.map((r) => (
                <tr
                  key={r.id}
                  onClick={() => router.push(`/restaurants/${r.id}`)}
                  style={{ cursor: "pointer" }}
                >
                  <td>{r.name}</td>
                  <td>{r.email || "—"}</td>
                  <td>
                    <StatusBadge status={r.status} />
                  </td>
                  <td>{r.is_contacted ? "Yes" : "No"}</td>
                  <td>{r.shown_interest ? "Yes" : "No"}</td>
                  <td>{r.email_send_count ?? 0}</td>
                  <td>{formatDate(r.last_email_sent_at)}</td>
                  <td>{formatDate(r.updated_at)}</td>
                  <td onClick={(e) => e.stopPropagation()}>
                    <Link href={`/restaurants/${r.id}`}>Open</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}

      {showSharedEmails && sharedEmailGroups.length > 0 ? (
        <div className="table-wrap">
          <table className="data">
            <thead>
              <tr>
                <th>Email</th>
                <th>Count</th>
                <th>Outreach</th>
                <th>Restaurants</th>
              </tr>
            </thead>
            <tbody>
              {sharedEmailGroups.map((group) => (
                <tr key={group.email}>
                  <td>{group.email}</td>
                  <td>{group.restaurant_count}</td>
                  <td>
                    <StatusBadge status={group.blocked_for_outreach ? "blocked" : "allowed"} />
                  </td>
                  <td>
                    <div style={{ display: "grid", gap: "0.3rem" }}>
                      {group.restaurants.map((restaurant) => (
                        <Link key={restaurant.id} href={`/restaurants/${restaurant.id}`}>
                          {restaurant.name} · {restaurant.status}
                        </Link>
                      ))}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
}

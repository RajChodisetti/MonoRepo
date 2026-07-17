"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { RESTAURANT_STATUSES, formatDate } from "@/lib/constants";
import type { Restaurant } from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";
import { SendPreviewModal } from "@/components/SendPreviewModal";

export default function RestaurantsPage() {
  const router = useRouter();
  const [items, setItems] = useState<Restaurant[]>([]);
  const [name, setName] = useState("");
  const [status, setStatus] = useState("");
  const [isContacted, setIsContacted] = useState("");
  const [shownInterest, setShownInterest] = useState("");
  const [includeArchived, setIncludeArchived] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [previewOpen, setPreviewOpen] = useState(false);

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

  useEffect(() => {
    const t = setTimeout(load, 200);
    return () => clearTimeout(t);
  }, [load]);

  return (
    <div>
      <PageHeader
        title="Restaurants"
        subtitle="Filter and open leads for review and outreach"
      />
      <ErrorBanner message={error} />

      <div
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
      </div>

      {selected.size > 0 ? (
        <div
          className="card"
          style={{
            marginBottom: "1rem",
            display: "flex",
            alignItems: "center",
            gap: "0.75rem",
            justifyContent: "space-between",
            position: "sticky",
            top: "0.5rem",
            zIndex: 10,
          }}
        >
          <span>{selected.size} selected</span>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button className="btn btn-secondary" type="button" onClick={() => setSelected(new Set())}>
              Clear
            </button>
            <button className="btn btn-primary" type="button" onClick={() => setPreviewOpen(true)}>
              Preview &amp; send {selected.size} selected
            </button>
          </div>
        </div>
      ) : null}

      {loading ? <EmptyState message="Loading restaurants…" /> : null}
      {!loading && items.length === 0 ? (
        <EmptyState message="No restaurants match these filters." />
      ) : null}

      {items.length > 0 ? (
        <div className="table-wrap">
          <table className="data">
            <thead>
              <tr>
                <th>
                  <input
                    type="checkbox"
                    checked={items.length > 0 && items.every((r) => selected.has(r.id))}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setSelected(new Set(items.map((r) => r.id)));
                      } else {
                        setSelected(new Set());
                      }
                    }}
                    aria-label="Select all"
                  />
                </th>
                <th>Name</th>
                <th>Email</th>
                <th>Status</th>
                <th>Contacted</th>
                <th>Interest</th>
                <th>Emailed</th>
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
                  <td onClick={(e) => e.stopPropagation()}>
                    <input
                      type="checkbox"
                      checked={selected.has(r.id)}
                      onChange={(e) => {
                        setSelected((prev) => {
                          const next = new Set(prev);
                          if (e.target.checked) next.add(r.id);
                          else next.delete(r.id);
                          return next;
                        });
                      }}
                      aria-label={`Select ${r.name}`}
                    />
                  </td>
                  <td>{r.name}</td>
                  <td>{r.email || "—"}</td>
                  <td>
                    <StatusBadge status={r.status} />
                  </td>
                  <td>{r.is_contacted ? "Yes" : "No"}</td>
                  <td>{r.shown_interest ? "Yes" : "No"}</td>
                  <td>
                    {r.email_sent
                      ? `Yes (${r.email_send_count ?? 1})`
                      : "No"}
                  </td>
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

      <SendPreviewModal
        open={previewOpen}
        restaurantIds={[...selected]}
        onClose={() => setPreviewOpen(false)}
        onSent={load}
      />
    </div>
  );
}

"use client";

import Link from "next/link";
import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import type { InboxListResponse, InboxThread } from "@/lib/types";
import { EmptyState, ErrorBanner, StatusBadge } from "@/components/ui";

export function OutreachInbox() {
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [data, setData] = useState<InboxListResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setError(null);
    try {
      const result = await adminFetch<InboxListResponse>("outreach/inbox", {
        query: { unread: unreadOnly ? true : undefined, limit: 50 },
      });
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load inbox");
    } finally {
      setLoading(false);
    }
  }, [unreadOnly]);

  useEffect(() => {
    setLoading(true);
    load();
    const timer = window.setInterval(load, 15_000);
    return () => window.clearInterval(timer);
  }, [load]);

  return (
    <div>
      <div
        className="card"
        style={{
          marginBottom: "1rem",
          display: "flex",
          gap: "0.75rem",
          alignItems: "center",
          justifyContent: "space-between",
          flexWrap: "wrap",
        }}
      >
        <div>
          <h2 style={{ margin: 0, fontSize: "1.05rem" }}>Outreach inbox</h2>
          <p style={{ color: "var(--muted)", margin: "0.3rem 0 0" }}>
            Sent snapshots and captured owner replies. A reply pauses that restaurant&apos;s campaign.
          </p>
        </div>
        <label style={{ display: "flex", gap: "0.4rem", alignItems: "center", fontSize: "0.9rem" }}>
          <input
            type="checkbox"
            checked={unreadOnly}
            onChange={(event) => setUnreadOnly(event.target.checked)}
          />
          Unread only
        </label>
      </div>
      <ErrorBanner message={error} />
      {loading && !data ? <EmptyState message="Loading inbox…" /> : null}
      {!loading && (data?.threads || []).length === 0 ? (
        <EmptyState message="No captured outreach mail yet." />
      ) : null}
      {(data?.threads || []).length > 0 ? (
        <div className="table-wrap">
          <table className="data">
            <thead>
              <tr>
                <th>Restaurant</th>
                <th>Latest</th>
                <th>When</th>
                <th>Unread</th>
              </tr>
            </thead>
            <tbody>
              {(data?.threads || []).map((thread) => (
                <InboxRow key={thread.last_message_id} thread={thread} />
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
}

function InboxRow({ thread }: { thread: InboxThread }) {
  const name = thread.restaurant_name || (thread.unmatched ? "Unmatched reply" : "Unknown restaurant");
  const title = thread.restaurant_id ? (
    <Link href={`/restaurants/${thread.restaurant_id}?tab=messages`}>{name}</Link>
  ) : (
    name
  );
  return (
    <tr>
      <td>
        {title}
        <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>{thread.email || "—"}</div>
      </td>
      <td>
        <StatusBadge status={thread.last_direction} />
        <div style={{ color: "var(--muted)", fontSize: "0.85rem", marginTop: "0.25rem" }}>
          {thread.last_snippet || "—"}
        </div>
      </td>
      <td>{formatDate(thread.last_at)}</td>
      <td>{thread.unread_count}</td>
    </tr>
  );
}

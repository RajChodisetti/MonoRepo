"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import type {
  OutreachRecipient,
  OutreachRecipientListResponse,
} from "@/lib/types";
import { EmptyState, ErrorBanner, StatusBadge } from "@/components/ui";

const PAGE_SIZE = 100;

type RecipientFilter = "all" | "followup" | "new" | "paused" | "completed";

function matchesFilter(recipient: OutreachRecipient, filter: RecipientFilter) {
  if (filter === "completed") return Boolean(recipient.completed_at);
  if (filter === "paused") {
    return !recipient.completed_at && (!recipient.eligible || recipient.suppressed);
  }
  if (filter === "followup") {
    return !recipient.completed_at && recipient.current_step > 0 && recipient.next_step != null;
  }
  if (filter === "new") {
    return !recipient.completed_at && recipient.current_step === 0 && recipient.next_step != null;
  }
  return true;
}

function holdLabel(recipient: OutreachRecipient) {
  if (recipient.completed_at) return "Sequence complete";
  if (recipient.suppressed) return "Opted out — permanently suppressed";
  if (recipient.hold_reason) return recipient.hold_reason;
  if (recipient.eligible) {
    return recipient.current_step > 0 ? "Due follow-up has priority" : "Eligible after due follow-ups";
  }
  return "Paused by outreach policy";
}

function humanize(value?: string) {
  if (!value) return "—";
  return value.replaceAll("_", " ");
}

export function OutreachRecipients() {
  const [items, setItems] = useState<OutreachRecipient[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [query, setQuery] = useState("");
  const [filter, setFilter] = useState<RecipientFilter>("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await adminFetch<OutreachRecipientListResponse>(
        "outreach/recipients",
        { query: { limit: PAGE_SIZE, offset } },
      );
      setItems(response.recipients || []);
      setTotal(response.total || 0);
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not load recipient progress");
    } finally {
      setLoading(false);
    }
  }, [offset]);

  useEffect(() => {
    load();
    const timer = window.setInterval(load, 30_000);
    return () => window.clearInterval(timer);
  }, [load]);

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    return items.filter((recipient) => {
      if (!matchesFilter(recipient, filter)) return false;
      if (!needle) return true;
      return `${recipient.restaurant_name} ${recipient.email}`.toLowerCase().includes(needle);
    });
  }, [filter, items, query]);

  const pageStart = total === 0 ? 0 : offset + 1;
  const pageEnd = Math.min(offset + items.length, total);

  return (
    <section aria-labelledby="recipient-progress-title">
      <div className="card" style={{ marginBottom: "1rem" }}>
        <h2 id="recipient-progress-title" style={{ margin: 0, fontSize: "1.05rem" }}>
          Recipient progress
        </h2>
        <p style={{ color: "var(--muted)", margin: "0.3rem 0 0", lineHeight: 1.5 }}>
          Confirmed sends advance the integer step. Failed or unknown provider outcomes do not.
          Due follow-ups are selected before any new restaurant.
        </p>
      </div>

      <div className="card recipient-filters">
        <label className="field-label" htmlFor="recipient-search">
          Search this page
          <input
            id="recipient-search"
            className="input"
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Restaurant or email"
          />
        </label>
        <label className="field-label" htmlFor="recipient-filter">
          Progress
          <select
            id="recipient-filter"
            className="select"
            value={filter}
            onChange={(event) => setFilter(event.target.value as RecipientFilter)}
          >
            <option value="all">All on this page</option>
            <option value="followup">Follow-ups</option>
            <option value="new">Not started</option>
            <option value="paused">Paused</option>
            <option value="completed">Completed</option>
          </select>
        </label>
        <div className="recipient-summary" aria-live="polite">
          {pageStart}–{pageEnd} of {total}
        </div>
      </div>

      <ErrorBanner message={error} />
      {loading && items.length === 0 ? <EmptyState message="Loading recipient progress…" /> : null}
      {!loading && !error && filtered.length === 0 ? (
        <EmptyState message="No recipients match this view." />
      ) : null}

      {filtered.length > 0 ? (
        <div className="table-wrap">
          <table className="data recipient-table">
            <thead>
              <tr>
                <th>Restaurant</th>
                <th>Lifecycle</th>
                <th>Consent</th>
                <th>Progress</th>
                <th>Last confirmed</th>
                <th>Next due</th>
                <th>State</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((recipient) => (
                <tr key={recipient.restaurant_id}>
                  <td data-label="Restaurant">
                    <div>
                      <Link className="recipient-name" href={`/restaurants/${recipient.restaurant_id}`}>
                        {recipient.restaurant_name}
                      </Link>
                      <div style={{ color: "var(--muted)", fontSize: "0.8rem", marginTop: "0.15rem" }}>
                        {recipient.email}
                      </div>
                    </div>
                  </td>
                  <td data-label="Lifecycle">
                    <StatusBadge status={humanize(recipient.lifecycle_status)} />
                  </td>
                  <td data-label="Consent">{humanize(recipient.consent_basis)}</td>
                  <td data-label="Progress">
                    <strong>{recipient.current_step}</strong> sent
                    <div style={{ color: "var(--muted)", fontSize: "0.8rem", marginTop: "0.15rem" }}>
                      {recipient.next_step != null
                        ? `Email ${recipient.next_step} next`
                        : "No remaining email"}
                      {recipient.email_send_count !== recipient.current_step
                        ? ` · ${recipient.email_send_count} recorded sends`
                        : ""}
                    </div>
                  </td>
                  <td data-label="Last confirmed">{formatDate(recipient.last_sent_at)}</td>
                  <td data-label="Next due">
                    {recipient.completed_at ? formatDate(recipient.completed_at) : formatDate(recipient.next_send_at)}
                  </td>
                  <td data-label="State">
                    <StatusBadge
                      status={
                        recipient.completed_at
                          ? "completed"
                          : recipient.suppressed
                            ? "suppressed"
                            : recipient.eligible
                              ? "eligible"
                              : "paused"
                      }
                    />
                    <div className="recipient-hold" style={{ marginTop: "0.3rem" }}>
                      {holdLabel(recipient)}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}

      <div style={{ display: "flex", justifyContent: "space-between", gap: "0.75rem", marginTop: "1rem", flexWrap: "wrap" }}>
        <button className="btn btn-secondary" type="button" onClick={load} disabled={loading}>
          {loading ? "Refreshing…" : "Refresh"}
        </button>
        <div style={{ display: "flex", gap: "0.5rem" }}>
          <button
            className="btn btn-secondary"
            type="button"
            onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
            disabled={offset === 0 || loading}
          >
            Previous
          </button>
          <button
            className="btn btn-secondary"
            type="button"
            onClick={() => setOffset(offset + PAGE_SIZE)}
            disabled={offset + items.length >= total || loading}
          >
            Next
          </button>
        </div>
      </div>
    </section>
  );
}

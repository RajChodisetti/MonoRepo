"use client";

import Link from "next/link";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import type { InboxListResponse, InboxThread } from "@/lib/types";
import { EmptyState, ErrorBanner, StatusBadge } from "@/components/ui";

const pageSize = 50;

export function OutreachInbox() {
  const [unreadOnly, setUnreadOnly] = useState(false);
  const [mailboxKey, setMailboxKey] = useState("");
  const [offset, setOffset] = useState(0);
  const [data, setData] = useState<InboxListResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setError(null);
    try {
      const result = await adminFetch<InboxListResponse>("outreach/inbox", {
        query: {
          unread: unreadOnly ? true : undefined,
          mailbox: mailboxKey || undefined,
          limit: pageSize,
          offset,
        },
      });
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load inbox");
    } finally {
      setLoading(false);
    }
  }, [mailboxKey, offset, unreadOnly]);

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
        <div style={{ display: "flex", gap: "0.75rem", alignItems: "center", flexWrap: "wrap" }}>
          <label style={{ display: "grid", gap: "0.25rem", fontSize: "0.85rem" }}>
            <span>Receiving mailbox</span>
            <select
              className="input"
              value={mailboxKey}
              onChange={(event) => {
                setMailboxKey(event.target.value);
                setOffset(0);
              }}
            >
              <option value="">All mailboxes</option>
              {(data?.mailboxes || []).map((mailbox) => (
                <option key={mailbox.mailbox_key} value={mailbox.mailbox_key}>
                  {mailbox.mailbox_key}
                </option>
              ))}
            </select>
          </label>
          <label style={{ display: "flex", gap: "0.4rem", alignItems: "center", fontSize: "0.9rem" }}>
            <input
              type="checkbox"
              checked={unreadOnly}
              onChange={(event) => {
                setUnreadOnly(event.target.checked);
                setOffset(0);
              }}
            />
            Unread only
          </label>
        </div>
      </div>
      <ErrorBanner message={error} />
      {(data?.mailboxes || []).map((mailbox) =>
        mailbox.last_error ? (
          <div className="alert alert-error" key={mailbox.mailbox_key} style={{ marginBottom: "0.75rem" }}>
            <strong>{mailbox.mailbox_key}</strong>: inbox access is unavailable. {mailbox.last_error}
          </div>
        ) : null,
      )}
      {loading && !data ? <EmptyState message="Loading inbox…" /> : null}
      {!loading && !error && (data?.threads || []).length === 0 ? (
        <EmptyState message="No inbox mail received in the last 10 days." />
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
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {(data?.threads || []).map((thread) => (
                <InboxRow key={thread.last_message_id} thread={thread} onReplied={load} />
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
      {data && data.total > 0 ? (
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            gap: "0.75rem",
            marginTop: "0.75rem",
            flexWrap: "wrap",
          }}
        >
          <span style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
            Showing {Math.min(offset + 1, data.total)}–{Math.min(offset + data.threads.length, data.total)} of {data.total}
          </span>
          <div style={{ display: "flex", gap: "0.5rem" }}>
            <button
              className="btn btn-secondary"
              type="button"
              disabled={offset === 0 || loading}
              onClick={() => setOffset((value) => Math.max(0, value - pageSize))}
            >
              Previous
            </button>
            <button
              className="btn btn-secondary"
              type="button"
              disabled={offset + pageSize >= data.total || loading}
              onClick={() => setOffset((value) => value + pageSize)}
            >
              Next
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function InboxRow({ thread, onReplied }: { thread: InboxThread; onReplied: () => Promise<void> }) {
  const [replying, setReplying] = useState(false);
  const [subject, setSubject] = useState("");
  const [bodyText, setBodyText] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const name = thread.restaurant_name || (thread.unmatched ? "Unmatched email" : "Unknown restaurant");
  const title = thread.restaurant_id ? (
    <Link href={`/restaurants/${thread.restaurant_id}?tab=messages`}>{name}</Link>
  ) : (
    name
  );
  async function sendReply(event: FormEvent) {
    event.preventDefault();
    if (!bodyText.trim()) return;
    if (!window.confirm(`Send this reply to ${thread.email || "the inbound sender"}?`)) return;
    setBusy(true);
    setError(null);
    try {
      await adminFetch(`outreach/messages/${thread.last_message_id}/reply`, {
        method: "POST",
        body: { subject: subject.trim() || undefined, body_text: bodyText },
      });
      setReplying(false);
      setSubject("");
      setBodyText("");
      await onReplied();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Reply failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <>
    <tr>
      <td>
        {title}
        <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>{thread.email || "—"}</div>
        <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
          Received by {thread.mailbox_email || thread.mailbox_key}
        </div>
      </td>
      <td>
        <StatusBadge status={thread.last_direction} />
        <div style={{ color: "var(--muted)", fontSize: "0.85rem", marginTop: "0.25rem" }}>
          {thread.last_snippet || "—"}
        </div>
      </td>
      <td>{formatDate(thread.last_at)}</td>
      <td>{thread.unread_count}</td>
      <td>
        {thread.last_direction === "inbound" ? (
          <button className="btn btn-secondary" type="button" onClick={() => setReplying((value) => !value)}>
            {replying ? "Cancel" : "Reply"}
          </button>
        ) : (
          <span style={{ color: "var(--muted)" }}>Waiting for reply</span>
        )}
      </td>
    </tr>
    {replying ? (
      <tr>
        <td colSpan={5}>
          <form onSubmit={sendReply} className="card" style={{ display: "grid", gap: "0.65rem", margin: "0.5rem 0" }}>
            <strong>Reply from the same outreach mailbox</strong>
            <label style={{ display: "grid", gap: "0.3rem" }}>
              <span>Subject / title (optional)</span>
              <input
                className="input"
                value={subject}
                onChange={(event) => setSubject(event.target.value)}
                maxLength={200}
                placeholder="Defaults to Re: original subject"
              />
            </label>
            <label style={{ display: "grid", gap: "0.3rem" }}>
              <span>Plain-text reply</span>
              <textarea
                className="textarea"
                rows={5}
                value={bodyText}
                onChange={(event) => setBodyText(event.target.value)}
                maxLength={10000}
                required
              />
            </label>
            {error ? <div className="alert alert-error">{error}</div> : null}
            <div>
              <button className="btn btn-primary" type="submit" disabled={busy || !bodyText.trim()}>
                {busy ? "Sending…" : "Review and send reply"}
              </button>
            </div>
          </form>
        </td>
      </tr>
    ) : null}
    </>
  );
}

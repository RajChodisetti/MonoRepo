"use client";

import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import type { BulkSendStatus } from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";

export default function OutreachPage() {
  const [status, setStatus] = useState<BulkSendStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);

  const load = useCallback(async () => {
    setError(null);
    try {
      const data = await adminFetch<BulkSendStatus>("outreach/bulk-send/status");
      setStatus(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load outreach status");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
    const t = setInterval(load, 5000);
    return () => clearInterval(t);
  }, [load]);

  async function trigger() {
    if (
      !confirm(
        "Start bulk outreach? This sends to eligible approved campaigns only.",
      )
    ) {
      return;
    }
    setSending(true);
    setError(null);
    setMessage(null);
    try {
      const res = await adminFetch<{
        job_id?: string;
        status?: string;
        pending_eligible_count?: number;
        max_sends?: number;
      }>("outreach/bulk-send", { method: "POST", body: {} });
      setMessage(
        `Bulk send accepted${res.job_id ? ` (job ${res.job_id})` : ""}. Status: ${
          res.status || "queued"
        }.`,
      );
      await load();
    } catch (err) {
      const e = err as Error & { status?: number };
      setError(
        e.message ||
          "Bulk send failed. Email may be disabled or accounts not configured.",
      );
    } finally {
      setSending(false);
    }
  }

  return (
    <div>
      <PageHeader
        title="Outreach"
        subtitle="Quota-managed bulk email for approved campaigns"
        actions={
          <button
            className="btn btn-primary"
            type="button"
            onClick={trigger}
            disabled={sending}
          >
            {sending ? "Starting…" : "Start bulk send"}
          </button>
        }
      />
      <ErrorBanner message={error} />
      {message ? (
        <div className="alert alert-info" style={{ marginBottom: "1rem" }}>
          {message}
        </div>
      ) : null}

      {loading && !status ? <EmptyState message="Loading outreach status…" /> : null}

      {status ? (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
            gap: "0.85rem",
            marginBottom: "1rem",
          }}
        >
          <div className="card">
            <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
              Pending eligible
            </div>
            <div style={{ fontSize: "1.8rem", fontWeight: 700 }}>
              {status.pending_eligible_count}
            </div>
          </div>
          <div className="card">
            <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
              Max sends / activation
            </div>
            <div style={{ fontSize: "1.8rem", fontWeight: 700 }}>
              {status.max_sends}
            </div>
          </div>
          <div className="card">
            <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
              Active job
            </div>
            <div style={{ marginTop: "0.35rem" }}>
              <StatusBadge status={status.active_job?.status || "none"} />
            </div>
            {status.active_job?.id ? (
              <div
                style={{
                  marginTop: "0.45rem",
                  fontSize: "0.8rem",
                  fontFamily: "monospace",
                  color: "var(--muted)",
                }}
              >
                {status.active_job.id}
              </div>
            ) : null}
          </div>
          <div className="card">
            <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
              Last completed
            </div>
            <div style={{ marginTop: "0.35rem" }}>
              <StatusBadge
                status={status.last_completed_job?.status || "none"}
              />
            </div>
            <div style={{ marginTop: "0.45rem", color: "var(--muted)", fontSize: "0.85rem" }}>
              {formatDate(status.last_completed_job?.created_at)}
            </div>
          </div>
        </div>
      ) : null}

      <div className="card">
        <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Eligibility checklist</h2>
        <ul style={{ margin: 0, paddingLeft: "1.1rem", color: "var(--muted)" }}>
          <li>OCR status verified</li>
          <li>Profile approved</li>
          <li>Demo published and unexpired</li>
          <li>Campaign approved</li>
          <li>Contact email present and not suppressed</li>
          <li>No prior confirmed send (`email_sent = false`)</li>
        </ul>
        <p style={{ marginBottom: 0, marginTop: "0.85rem", color: "var(--muted)" }}>
          If you see 503 errors, enable email sending and configure outreach
          accounts in VM <code>stack.env</code> (Zoho / Google Workspace JSON).
        </p>
        {status?.next_available_at ? (
          <p style={{ marginBottom: 0, marginTop: "0.5rem" }}>
            Next available: {formatDate(status.next_available_at)}
          </p>
        ) : null}
      </div>
    </div>
  );
}

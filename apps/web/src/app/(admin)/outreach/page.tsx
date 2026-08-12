"use client";

import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import type { BulkSendStatus, EmailAccountHealthResponse } from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";

export default function OutreachPage() {
  const [status, setStatus] = useState<BulkSendStatus | null>(null);
  const [emailHealth, setEmailHealth] = useState<EmailAccountHealthResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [settingJob, setSettingJob] = useState(false);

  const load = useCallback(async () => {
    setError(null);
    try {
      const [bulkData, healthData] = await Promise.all([
        adminFetch<BulkSendStatus>("outreach/bulk-send/status"),
        adminFetch<EmailAccountHealthResponse>("outreach/email-accounts/health"),
      ]);
      setStatus(bulkData);
      setEmailHealth(healthData);
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

  async function setEmailJob(enabled: boolean) {
    if (
      !confirm(
        enabled
          ? "Enable the email job now? This starts real Gmail outreach to OCR-verified leads that also have approved profiles, published demos, approved campaigns, and valid email addresses."
          : "Disable the email job? No new Gmail delivery will begin after the current provider request finishes.",
      )
    ) {
      return;
    }
    setSettingJob(true);
    setError(null);
    setMessage(null);
    try {
      const res = await adminFetch<{
        job_id?: string;
        status?: string;
        pending_eligible_count?: number;
        max_sends?: number;
      }>("outreach/email-job", { method: "PATCH", body: { enabled } });
      setMessage(
        enabled
          ? `Email job enabled${res.job_id ? ` (job ${res.job_id})` : ""}. Status: ${res.status || "queued"}.`
          : "Email job disabled. No additional leads will be sent until you enable it again.",
      );
      await load();
    } catch (err) {
      const e = err as Error & { status?: number };
      setError(
        e.message ||
          "Email job update failed. Gmail OAuth accounts may not be configured.",
      );
    } finally {
      setSettingJob(false);
    }
  }

  return (
    <div>
      <PageHeader
        title="Outreach"
        subtitle="UI-controlled, quota-managed Gmail outreach for approved leads"
        actions={
          <button
            className={status?.email_job.enabled ? "btn btn-danger" : "btn btn-primary"}
            type="button"
            onClick={() => setEmailJob(!status?.email_job.enabled)}
            disabled={settingJob || !status}
          >
            {settingJob
              ? "Updating…"
              : status?.email_job.enabled
                ? "Disable email job"
                : "Enable email job"}
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
              Email job control
            </div>
            <div style={{ marginTop: "0.35rem" }}>
              <StatusBadge status={status.email_job.enabled ? "enabled" : "disabled"} />
            </div>
            <div style={{ marginTop: "0.45rem", color: "var(--muted)", fontSize: "0.8rem" }}>
              {status.email_job.enabled_at
                ? `Enabled ${formatDate(status.email_job.enabled_at)}`
                : "Enable from this page to start a run"}
            </div>
          </div>
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
            {status.active_job?.job_id ? (
              <div
                style={{
                  marginTop: "0.45rem",
                  fontSize: "0.8rem",
                  fontFamily: "monospace",
                  color: "var(--muted)",
                }}
              >
                {status.active_job.job_id}
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
              {status.last_completed_job?.summary
                ? `${status.last_completed_job.summary.sent} sent · ${status.last_completed_job.summary.failed} failed`
                : "No completed run yet"}
            </div>
          </div>
        </div>
      ) : null}

      <div className="card" style={{ marginBottom: "1rem" }}>
        <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Gmail sender health</h2>
        <p style={{ color: "var(--muted)", marginTop: 0 }}>
          Each configured Gmail OAuth mailbox sends a real health-check message to {emailHealth?.recipient || "the configured recipient"} every {emailHealth?.interval_hours || 24} hours.
        </p>
        {!emailHealth || emailHealth.accounts.length === 0 ? (
          <EmptyState message="No Gmail outreach accounts are configured in OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON." />
        ) : (
          <div className="table-wrap">
            <table className="data">
              <thead>
                <tr>
                  <th>Sender</th>
                  <th>Status</th>
                  <th>Last checked</th>
                  <th>Next check</th>
                  <th>Result</th>
                </tr>
              </thead>
              <tbody>
                {emailHealth.accounts.map((account) => (
                  <tr key={account.provider_identity}>
                    <td>{account.from_email}</td>
                    <td><StatusBadge status={account.status} /></td>
                    <td>{formatDate(account.last_checked_at)}</td>
                    <td>{formatDate(account.next_check_at)}</td>
                    <td style={{ color: account.last_error ? "var(--danger)" : "var(--muted)" }}>
                      {account.last_error || (account.provider_message_id ? "Accepted by Gmail" : "Waiting for first check")}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      <div className="card">
        <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Eligibility checklist</h2>
        <ul style={{ margin: 0, paddingLeft: "1.1rem", color: "var(--muted)" }}>
          <li>OCR status verified</li>
          <li>Profile approved</li>
          <li>Demo published and unexpired</li>
          <li>Campaign approved</li>
          <li>Contact email present</li>
          <li>No prior confirmed send (`email_sent = false`)</li>
        </ul>
        <p style={{ marginBottom: 0, marginTop: "0.85rem", color: "var(--muted)" }}>
          Mailbox credentials stay in secret configuration through <code>OUTREACH_GOOGLE_WORKSPACE_ACCOUNTS_JSON</code>.
          The job itself is enabled and disabled only from this page. Each mailbox needs an OAuth client ID,
          client secret, and refresh token authorized for Gmail send access; Google API keys are not mailbox credentials.
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

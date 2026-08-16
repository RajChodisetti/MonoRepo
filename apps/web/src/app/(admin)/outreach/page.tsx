"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";
import { OutreachInbox } from "@/components/OutreachInbox";
import { OutreachEmailAccounts } from "@/components/OutreachEmailAccounts";
import { OutreachRecipients } from "@/components/OutreachRecipients";
import { OutreachSequenceEditor } from "@/components/OutreachSequenceEditor";
import { RestaurantSearch } from "@/components/RestaurantSearch";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import type {
  BulkSendStatus,
  EmailAccountHealthResponse,
  OutreachSequence,
  OutreachSequenceListResponse,
  OutreachTemplateTestSendResponse,
  Restaurant,
} from "@/lib/types";

type View = "sequence" | "recipients" | "operations" | "accounts" | "inbox";

export default function OutreachPage() {
  const [view, setView] = useState<View>("sequence");
  const [status, setStatus] = useState<BulkSendStatus | null>(null);
  const [emailHealth, setEmailHealth] = useState<EmailAccountHealthResponse | null>(null);
  const [sequences, setSequences] = useState<OutreachSequence[]>([]);
  const [activeSequenceId, setActiveSequenceId] = useState<string | undefined>();
  const [operationsError, setOperationsError] = useState<string | null>(null);
  const [sequenceError, setSequenceError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [testRecipientEmail, setTestRecipientEmail] = useState("");
  const [testRestaurantName, setTestRestaurantName] = useState("Tuvi Test Restaurant");
  const [testOwnerFirstName, setTestOwnerFirstName] = useState("");
  const [selectedTestRestaurant, setSelectedTestRestaurant] = useState<Restaurant | null>(null);
  const [testSendResult, setTestSendResult] = useState<OutreachTemplateTestSendResponse | null>(null);
  const [loadingOperations, setLoadingOperations] = useState(true);
  const [loadingSequences, setLoadingSequences] = useState(true);
  const [settingJob, setSettingJob] = useState(false);
  const [sendingTest, setSendingTest] = useState(false);

  const activeSequence = useMemo(
    () => sequences.find((sequence) => sequence.id === activeSequenceId || sequence.is_active),
    [activeSequenceId, sequences],
  );

  const loadOperations = useCallback(async () => {
    setOperationsError(null);
    const [bulkResult, healthResult] = await Promise.allSettled([
      adminFetch<BulkSendStatus>("outreach/bulk-send/status"),
      adminFetch<EmailAccountHealthResponse>("outreach/email-accounts/health"),
    ]);
    if (bulkResult.status === "fulfilled") {
      setStatus(bulkResult.value);
    } else {
      setOperationsError(
        bulkResult.reason instanceof Error
          ? bulkResult.reason.message
          : "Failed to load outreach status",
      );
    }
    if (healthResult.status === "fulfilled") {
      setEmailHealth(healthResult.value);
    }
    setLoadingOperations(false);
  }, []);

  const loadSequences = useCallback(async () => {
    setSequenceError(null);
    try {
      const result = await adminFetch<OutreachSequenceListResponse>("outreach/sequences");
      setSequences(result.sequences || []);
      setActiveSequenceId(result.active_sequence_id);
    } catch (reason) {
      setSequenceError(reason instanceof Error ? reason.message : "Failed to load sequences");
    } finally {
      setLoadingSequences(false);
    }
  }, []);

  useEffect(() => {
    void Promise.all([loadOperations(), loadSequences()]);
    const timer = window.setInterval(loadOperations, 15_000);
    return () => window.clearInterval(timer);
  }, [loadOperations, loadSequences]);

  async function setEmailJob(enabled: boolean) {
    if (enabled && !activeSequence) {
      setOperationsError("Approve an outreach sequence before enabling the email job.");
      setView("sequence");
      return;
    }
    if (
      !confirm(
        enabled
          ? `Enable real Gmail outreach using “${activeSequence?.name}”? Due follow-ups will be sent before new restaurants, subject to pacing, lifecycle, and consent checks.`
          : "Disable the email job? No new Gmail delivery will begin after the current provider request finishes.",
      )
    ) {
      return;
    }
    setSettingJob(true);
    setOperationsError(null);
    setMessage(null);
    try {
      const result = await adminFetch<{
        job_id?: string;
        status?: string;
      }>("outreach/email-job", { method: "PATCH", body: { enabled } });
      setMessage(
        enabled
          ? `Email job enabled${result.job_id ? ` (job ${result.job_id})` : ""}. Follow-ups remain first in the queue.`
          : "Email job disabled. Recipient progress and due dates are preserved.",
      );
      await loadOperations();
    } catch (reason) {
      setOperationsError(
        reason instanceof Error
          ? reason.message
          : "Email job update failed. Check the approved sequence and Gmail configuration.",
      );
    } finally {
      setSettingJob(false);
    }
  }

  async function sendTemplateTest(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const recipient = testRecipientEmail.trim();
    const restaurantName = selectedTestRestaurant?.name || testRestaurantName.trim() || "Tuvi Test Restaurant";
    if (!recipient) {
      setOperationsError("Enter a recipient email for the template test.");
      return;
    }
    if (!activeSequence) {
      setOperationsError("Make an outreach template active before sending an active-template test.");
      return;
    }
    if (
      !confirm(
        `Send every enabled email from the active template to ${recipient}? This uses the real configured sender account.`,
      )
    ) {
      return;
    }
    setSendingTest(true);
    setOperationsError(null);
    setMessage(null);
    setTestSendResult(null);
    try {
      const result = await adminFetch<OutreachTemplateTestSendResponse>("outreach/test-send", {
        method: "POST",
        body: {
          recipient_email: recipient,
          sequence_id: activeSequence.id,
          restaurant_id: selectedTestRestaurant?.id,
          restaurant_name: selectedTestRestaurant ? undefined : restaurantName,
          owner_first_name: selectedTestRestaurant
            ? undefined
            : testOwnerFirstName.trim() || undefined,
        },
      });
      setTestSendResult(result);
      setMessage(`Sent ${result.items.length} outreach test emails to ${result.recipient_email}.`);
    } catch (reason) {
      setOperationsError(
        reason instanceof Error
          ? reason.message
          : "Template test send failed. Check the active sequence and sender configuration.",
      );
    } finally {
      setSendingTest(false);
    }
  }

  const tabs: { id: View; label: string }[] = [
    { id: "sequence", label: "Email sequence" },
    { id: "recipients", label: "Recipient progress" },
    { id: "operations", label: "Sending & health" },
    { id: "accounts", label: "Email accounts" },
    { id: "inbox", label: "Inbox" },
  ];

  return (
    <div>
      <PageHeader
        title="Outreach"
        subtitle="Approved plain-text sequences with durable follow-up scheduling"
        actions={
          <button
            className={status?.email_job.enabled ? "btn btn-danger" : "btn btn-primary"}
            type="button"
            onClick={() => setEmailJob(!status?.email_job.enabled)}
            disabled={settingJob || !status || (!status.email_job.enabled && !activeSequence)}
          >
            {settingJob
              ? "Updating…"
              : status?.email_job.enabled
                ? "Disable email job"
                : "Enable email job"}
          </button>
        }
      />

      <ErrorBanner message={view === "sequence" ? sequenceError : operationsError} />
      {message ? (
        <div className="alert alert-info" style={{ marginBottom: "1rem" }} role="status" aria-live="polite">
          {message}
        </div>
      ) : null}

      <div className="outreach-view-tabs" role="tablist" aria-label="Outreach sections">
        {tabs.map((tab) => (
          <button
            className="outreach-view-tab"
            type="button"
            role="tab"
            id={`outreach-tab-${tab.id}`}
            aria-controls={`outreach-panel-${tab.id}`}
            aria-selected={view === tab.id}
            tabIndex={view === tab.id ? 0 : -1}
            onClick={() => setView(tab.id)}
            key={tab.id}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {view === "sequence" ? (
        <div role="tabpanel" id="outreach-panel-sequence" aria-labelledby="outreach-tab-sequence">
          <OutreachSequenceEditor
            sequences={sequences}
            activeSequenceId={activeSequenceId}
            loading={loadingSequences}
            onReload={loadSequences}
          />
        </div>
      ) : null}

      {view === "recipients" ? (
        <div role="tabpanel" id="outreach-panel-recipients" aria-labelledby="outreach-tab-recipients">
          <OutreachRecipients />
        </div>
      ) : null}

      {view === "inbox" ? (
        <div role="tabpanel" id="outreach-panel-inbox" aria-labelledby="outreach-tab-inbox">
          <OutreachInbox />
        </div>
      ) : null}

      {view === "accounts" ? (
        <div role="tabpanel" id="outreach-panel-accounts" aria-labelledby="outreach-tab-accounts">
          <OutreachEmailAccounts />
        </div>
      ) : null}

      {view === "operations" ? (
        <div role="tabpanel" id="outreach-panel-operations" aria-labelledby="outreach-tab-operations">
          {loadingOperations && !status ? <EmptyState message="Loading outreach status…" /> : null}

          {status ? (
            <div className="outreach-metrics">
              <div className="card">
                <div className="outreach-metric-label">Email job</div>
                <div style={{ marginTop: "0.4rem" }}>
                  <StatusBadge status={status.email_job.enabled ? "enabled" : "disabled"} />
                </div>
                <div className="field-help" style={{ marginTop: "0.45rem" }}>
                  {status.email_job.enabled_at
                    ? `Enabled ${formatDate(status.email_job.enabled_at)}`
                    : "Sending remains off until explicitly enabled"}
                </div>
              </div>
              <div className="card">
                <div className="outreach-metric-label">Due follow-ups</div>
                <div className="outreach-metric-value">{status.due_followup_count ?? "—"}</div>
              </div>
              <div className="card">
                <div className="outreach-metric-label">New eligible</div>
                <div className="outreach-metric-value">
                  {status.new_recipient_count ?? status.pending_eligible_count}
                </div>
              </div>
              <div className="card">
                <div className="outreach-metric-label">Active job</div>
                <div style={{ marginTop: "0.4rem" }}>
                  <StatusBadge status={status.active_job?.status || "none"} />
                </div>
              </div>
              <div className="card">
                <div className="outreach-metric-label">Last completed run</div>
                <div style={{ marginTop: "0.4rem" }}>
                  <StatusBadge status={status.last_completed_job?.status || "none"} />
                </div>
                <div className="field-help" style={{ marginTop: "0.45rem" }}>
                  {status.last_completed_job?.summary
                    ? `${status.last_completed_job.summary.sent} sent · ${status.last_completed_job.summary.failed} failed`
                    : "No completed run yet"}
                </div>
              </div>
            </div>
          ) : null}

          <div className="card" style={{ marginBottom: "1rem" }}>
            <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Active template</h2>
            {activeSequence ? (
              <div style={{ display: "flex", gap: "0.65rem", alignItems: "center", flexWrap: "wrap" }}>
                <StatusBadge status="active" />
                <strong>{activeSequence.name}</strong>
                <span style={{ color: "var(--muted)" }}>
                  {activeSequence.steps.filter((step) => step.enabled).length} enabled templates
                </span>
              </div>
            ) : (
              <p style={{ color: "var(--muted)", marginBottom: 0 }}>
                No active template. Create and make one active before enabling sending.
              </p>
            )}
          </div>

          <form className="card" style={{ marginBottom: "1rem" }} onSubmit={sendTemplateTest}>
            <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Send template test</h2>
            <RestaurantSearch
              label="Saved restaurant (optional)"
              selected={selectedTestRestaurant}
              onSelect={(restaurant) => {
                setSelectedTestRestaurant(restaurant);
                if (restaurant) {
                  setTestRestaurantName(restaurant.name);
                  setTestOwnerFirstName("");
                }
                setTestSendResult(null);
              }}
              help="When selected, the server ignores synthetic fields; Google fields are used only from a verified successful profile."
            />
            <div className="form-grid">
              <label>
                Recipient email
                <input
                  className="input"
                  type="email"
                  value={testRecipientEmail}
                  onChange={(event) => {
                    setTestRecipientEmail(event.target.value);
                    setTestSendResult(null);
                  }}
                  placeholder="name@example.com"
                  required
                />
              </label>
              <label>
                Synthetic restaurant name
                <input
                  className="input"
                  value={testRestaurantName}
                  onChange={(event) => {
                    setTestRestaurantName(event.target.value);
                    setTestSendResult(null);
                  }}
                  placeholder="Tuvi Test Restaurant"
                  disabled={selectedTestRestaurant !== null}
                />
              </label>
              <label>
                Synthetic owner first name
                <input
                  className="input"
                  value={testOwnerFirstName}
                  onChange={(event) => {
                    setTestOwnerFirstName(event.target.value);
                    setTestSendResult(null);
                  }}
                  placeholder="Optional"
                  disabled={selectedTestRestaurant !== null}
                />
              </label>
            </div>
            <div style={{ marginTop: "0.85rem", display: "flex", gap: "0.65rem", alignItems: "center", flexWrap: "wrap" }}>
              <button className="btn btn-primary" type="submit" disabled={sendingTest || !activeSequence}>
                {sendingTest ? "Sending..." : "Send test emails"}
              </button>
              <span className="field-help">
                Sends every enabled email from the active template to this address only.
              </span>
            </div>
            {testSendResult ? (
              <div style={{ marginTop: "1rem" }}>
                <div className="greeting-audit" aria-live="polite">
                  <span className="field-help">Server-rendered greeting01</span>
                  <pre>{testSendResult.greeting01}</pre>
                  <div className="facts-used">
                    <strong>Facts used</strong>
                    <span>
                      {testSendResult.facts_used.length > 0
                        ? testSendResult.facts_used.join(", ")
                        : "Fallback wording only"}
                    </span>
                  </div>
                </div>
                <div className="table-wrap" style={{ marginTop: "1rem" }}>
                <table className="data">
                  <thead>
                    <tr>
                      <th>Template</th>
                      <th>Subject</th>
                      <th>Provider ID</th>
                    </tr>
                  </thead>
                  <tbody>
                    {testSendResult.items.map((item, index) => (
                      <tr key={`${item.template}-${item.step || 0}-${index}`}>
                        <td>{`Template ${item.step}`}</td>
                        <td>{item.subject}</td>
                        <td style={{ fontFamily: "monospace", color: "var(--muted)" }}>
                          {item.provider_message_id || "accepted"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                </div>
              </div>
            ) : null}
          </form>

          <div className="card" style={{ marginBottom: "1rem" }}>
            <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Gmail sender health</h2>
            <p style={{ color: "var(--muted)", marginTop: 0 }}>
              Mailboxes remain quota-managed and paced. Credentials come from protected environment configuration or encrypted database storage.
            </p>
            {!emailHealth || emailHealth.accounts.length === 0 ? (
              <EmptyState message="No Gmail outreach accounts are configured." />
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
                        <td style={{ color: account.last_error ? "var(--bad)" : "var(--muted)", whiteSpace: "normal" }}>
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
            <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Eligibility and stop rules</h2>
            <ul style={{ margin: 0, paddingLeft: "1.1rem", color: "var(--muted)", lineHeight: 1.65 }}>
              <li>Restaurant name and a valid email are present.</li>
              <li>Imported business leads carry recorded <code>inferred_business</code> consent evidence.</li>
              <li>Lifecycle is <code>lead</code> or <code>emailed</code>.</li>
              <li>Interest pauses automation for human follow-up.</li>
              <li>Lost, archived, onboarding, and active-client restaurants are excluded.</li>
              <li>An approved active plain-text sequence is available.</li>
              <li>Each email contains only the Tuvi Solutions site link.</li>
              <li>Due follow-ups are exhausted before new recipients are selected.</li>
            </ul>
            {status?.next_available_at ? (
              <p style={{ marginBottom: 0, color: "var(--muted)" }}>
                Next mailbox slot: {formatDate(status.next_available_at)}
              </p>
            ) : null}
          </div>
        </div>
      ) : null}
    </div>
  );
}

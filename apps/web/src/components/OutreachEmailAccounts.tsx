"use client";

import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";
import { EmptyState, ErrorBanner, StatusBadge } from "@/components/ui";
import { adminFetch } from "@/lib/client-api";
import type {
  OutreachEmailAccount,
  OutreachEmailAccountListResponse,
} from "@/lib/types";

type AccountForm = {
  accountKey: string;
  mailboxEmail: string;
  fromEmail: string;
  clientID: string;
  clientSecret: string;
  refreshToken: string;
};

const emptyForm: AccountForm = {
  accountKey: "",
  mailboxEmail: "",
  fromEmail: "",
  clientID: "",
  clientSecret: "",
  refreshToken: "",
};

export function OutreachEmailAccounts() {
  const [data, setData] = useState<OutreachEmailAccountListResponse | null>(null);
  const [form, setForm] = useState<AccountForm>(emptyForm);
  const [editingAccount, setEditingAccount] = useState<OutreachEmailAccount | null>(null);
  const [replacement, setReplacement] = useState<AccountForm>(emptyForm);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [busyID, setBusyID] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      setData(await adminFetch<OutreachEmailAccountListResponse>("outreach/email-accounts"));
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Failed to load email accounts.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  async function addAccount(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError(null);
    setMessage(null);
    try {
      await adminFetch<OutreachEmailAccount>("outreach/email-accounts", {
        method: "POST",
        body: {
          account_key: form.accountKey.trim().toLowerCase(),
          mailbox_email: form.mailboxEmail.trim(),
          from_email: form.fromEmail.trim() || undefined,
          credentials: {
            client_id: form.clientID.trim(),
            client_secret: form.clientSecret.trim(),
            refresh_token: form.refreshToken.trim(),
          },
          enabled: true,
        },
      });
      setForm(emptyForm);
      setMessage("Email account added. It will be used by sending and inbox polling without a restart.");
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Failed to add email account.");
    } finally {
      setSaving(false);
    }
  }

  async function toggleAccount(account: OutreachEmailAccount) {
    if (!account.id || !account.editable) return;
    if (account.enabled) {
      const warning = account.overrides_environment
        ? `Disable ${account.mailbox_email}? The database override takes precedence, so its environment credentials will not be used as a fallback.`
        : `Disable ${account.mailbox_email}? It will stop participating in outreach sending and inbox polling.`;
      if (!window.confirm(warning)) return;
    }
    const busyKey = accountBusyKey(account);
    setBusyID(busyKey);
    setError(null);
    setMessage(null);
    try {
      await adminFetch<OutreachEmailAccount>(`outreach/email-accounts/${account.id}`, {
        method: "PATCH",
        body: { enabled: !account.enabled },
      });
      setMessage(`${account.mailbox_email} ${account.enabled ? "disabled" : "enabled"}.`);
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Failed to update email account.");
    } finally {
      setBusyID(null);
    }
  }

  function beginReplacement(account: OutreachEmailAccount) {
    setEditingAccount(account);
    setReplacement({ ...emptyForm, fromEmail: account.from_email });
    setError(null);
    setMessage(null);
  }

  async function replaceCredentials(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!editingAccount) return;
    const createsOverride = !editingAccount.id;
    const warning = createsOverride
      ? `Replace credentials for ${editingAccount.mailbox_email}? This creates a database override. Database values will take precedence over environment values.`
      : `Replace the stored OAuth credentials for ${editingAccount.mailbox_email}? Existing secret values cannot be recovered.`;
    if (!window.confirm(warning)) return;
    const busyKey = accountBusyKey(editingAccount);
    setBusyID(busyKey);
    setError(null);
    setMessage(null);
    try {
      const credentials = {
        client_id: replacement.clientID.trim(),
        client_secret: replacement.clientSecret.trim(),
        refresh_token: replacement.refreshToken.trim(),
      };
      if (editingAccount.id) {
        await adminFetch<OutreachEmailAccount>(`outreach/email-accounts/${editingAccount.id}`, {
          method: "PATCH",
          body: { from_email: replacement.fromEmail.trim(), credentials },
        });
      } else {
        await adminFetch<OutreachEmailAccount>("outreach/email-accounts", {
          method: "POST",
          body: {
            account_key: editingAccount.account_key,
            mailbox_email: editingAccount.mailbox_email,
            from_email: replacement.fromEmail.trim(),
            credentials,
            enabled: true,
          },
        });
      }
      setReplacement(emptyForm);
      setEditingAccount(null);
      setMessage(
        createsOverride
          ? "Database override saved. It now takes precedence for sending and inbox polling."
          : "Sender address and OAuth credentials replaced securely.",
      );
      await load();
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Failed to replace credentials.");
    } finally {
      setBusyID(null);
    }
  }

  return (
    <div style={{ display: "grid", gap: "1rem" }}>
      <ErrorBanner message={error} />
      {message ? (
        <div className="alert alert-info" role="status" aria-live="polite">
          {message}
        </div>
      ) : null}

      <section className="card" aria-labelledby="email-account-add-heading">
        <h2 id="email-account-add-heading" style={{ marginTop: 0, fontSize: "1.05rem" }}>
          Add Gmail account
        </h2>
        <p style={{ color: "var(--muted)", marginTop: 0, lineHeight: 1.5 }}>
          Add a Google Workspace mailbox once and it becomes available to outreach sending and the
          unified inbox. OAuth authorization must include <code>gmail.send</code> and{
          " "}<code>gmail.readonly</code>. Secrets are encrypted before database storage and are never
          returned to this page.
          If an account already comes from the environment, use its <strong>Replace credentials</strong>{" "}
          action below; the saved database override will take precedence.
        </p>
        {data && !data.encryption_ready ? (
          <div className="alert alert-error" role="alert" style={{ marginBottom: "1rem" }}>
            Secure database credential storage is not configured. Environment-only accounts remain
            active, but database overrides are unavailable and new credentials cannot be saved.
          </div>
        ) : null}
        <form onSubmit={addAccount}>
          <div className="form-grid">
            <label className="field-label">
              Account key
              <input
                className="input"
                value={form.accountKey}
                onChange={(event) => setForm({ ...form, accountKey: event.target.value })}
                placeholder="sales-us"
                pattern="[a-z0-9][a-z0-9_-]{1,63}"
                maxLength={64}
                autoComplete="off"
                required
              />
              <span className="field-help">Stable lowercase identifier; it cannot be changed later.</span>
            </label>
            <label className="field-label">
              Gmail mailbox
              <input
                className="input"
                type="email"
                value={form.mailboxEmail}
                onChange={(event) => setForm({ ...form, mailboxEmail: event.target.value })}
                placeholder="sales@example.com"
                autoComplete="off"
                required
              />
            </label>
            <label className="field-label">
              From address
              <input
                className="input"
                type="email"
                value={form.fromEmail}
                onChange={(event) => setForm({ ...form, fromEmail: event.target.value })}
                placeholder="Defaults to mailbox"
                autoComplete="off"
              />
            </label>
            <CredentialFields form={form} setForm={setForm} prefix="add" />
          </div>
          <div style={{ marginTop: "1rem", display: "flex", alignItems: "center", gap: "0.65rem", flexWrap: "wrap" }}>
            <button className="btn btn-primary" type="submit" disabled={saving || data?.encryption_ready === false}>
              {saving ? "Saving…" : "Add email account"}
            </button>
            <span className="field-help">New identities must be unique; partial account-key or mailbox conflicts are rejected.</span>
          </div>
        </form>
      </section>

      <section className="card" aria-labelledby="email-account-list-heading">
        <h2 id="email-account-list-heading" style={{ marginTop: 0, fontSize: "1.05rem" }}>
          Connected accounts
        </h2>
        <p style={{ color: "var(--muted)", marginTop: 0 }}>
          Credentials can be replaced for every account. Replacing an environment account creates an
          encrypted database override; database values always take precedence over environment values.
        </p>
        {loading && !data ? <EmptyState message="Loading connected email accounts…" /> : null}
        {data && data.accounts.length === 0 ? <EmptyState message="No Gmail accounts are configured." /> : null}
        {data && data.accounts.length > 0 ? (
          <div className="table-wrap">
            <table className="data">
              <thead>
                <tr>
                  <th>Mailbox</th>
                  <th>Account key</th>
                  <th>Source</th>
                  <th>Status</th>
                  <th>Credentials</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {data.accounts.map((account) => (
                  <tr key={`${account.source}-${account.id || account.account_key}`}>
                    <td>
                      <strong>{account.mailbox_email}</strong>
                      {account.from_email !== account.mailbox_email ? (
                        <div className="field-help">From {account.from_email}</div>
                      ) : null}
                    </td>
                    <td><code>{account.account_key}</code></td>
                    <td>
                      {account.source === "environment"
                        ? "Environment"
                        : account.overrides_environment
                          ? "Database override"
                          : "Database"}
                    </td>
                    <td>
                      <StatusBadge
                        status={
                          account.effective
                            ? "active"
                            : account.enabled
                              ? "unavailable"
                              : "disabled"
                        }
                      />
                      {account.overrides_environment && !account.effective ? (
                        <div className="field-help">Environment fallback blocked</div>
                      ) : null}
                    </td>
                    <td>{account.credentials_stored ? "Stored securely" : "Missing"}</td>
                    <td>
                      {account.editable ? (
                        <div style={{ display: "flex", gap: "0.45rem" }}>
                          <button
                            className={account.enabled ? "btn btn-danger btn-compact" : "btn btn-secondary btn-compact"}
                            type="button"
                            disabled={busyID === accountBusyKey(account)}
                            onClick={() => void toggleAccount(account)}
                          >
                            {account.enabled ? "Disable" : "Enable"}
                          </button>
                          <button
                            className="btn btn-secondary btn-compact"
                            type="button"
                            disabled={busyID === accountBusyKey(account)}
                            onClick={() => beginReplacement(account)}
                          >
                            Replace credentials
                          </button>
                        </div>
                      ) : (
                        <button
                          className="btn btn-secondary btn-compact"
                          type="button"
                          disabled={data.encryption_ready === false || busyID === accountBusyKey(account)}
                          onClick={() => beginReplacement(account)}
                        >
                          Replace credentials
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>

      {editingAccount ? (
        <form className="card" onSubmit={replaceCredentials} aria-labelledby="replace-credentials-heading">
          <h2 id="replace-credentials-heading" style={{ marginTop: 0, fontSize: "1.05rem" }}>
            Replace OAuth credentials
          </h2>
          <p style={{ color: "var(--muted)", marginTop: 0 }}>
            Enter the complete new credential set for <strong>{editingAccount.mailbox_email}</strong>. Existing
            secret values cannot be viewed. {editingAccount.source === "environment" ? (
              <>Saving creates an encrypted database override and leaves the environment configuration ignored.</>
            ) : null}
          </p>
          <div className="form-grid">
            <label className="field-label">
              From address
              <input
                className="input"
                type="email"
                value={replacement.fromEmail}
                onChange={(event) => setReplacement({ ...replacement, fromEmail: event.target.value })}
                autoComplete="off"
                required
              />
            </label>
            <CredentialFields form={replacement} setForm={setReplacement} prefix="replace" />
          </div>
          <div style={{ marginTop: "1rem", display: "flex", gap: "0.65rem" }}>
            <button className="btn btn-primary" type="submit" disabled={busyID === accountBusyKey(editingAccount)}>
              {busyID === accountBusyKey(editingAccount) ? "Replacing…" : "Replace credentials"}
            </button>
            <button className="btn btn-secondary" type="button" onClick={() => setEditingAccount(null)}>
              Cancel
            </button>
          </div>
        </form>
      ) : null}
    </div>
  );
}

function accountBusyKey(account: OutreachEmailAccount) {
  return account.id || `environment:${account.account_key}:${account.mailbox_email}`;
}

function CredentialFields({
  form,
  setForm,
  prefix,
}: {
  form: AccountForm;
  setForm: (next: AccountForm) => void;
  prefix: string;
}) {
  return (
    <>
      <label className="field-label" htmlFor={`${prefix}-client-id`}>
        OAuth client ID
        <input
          id={`${prefix}-client-id`}
          className="input"
          type="password"
          value={form.clientID}
          onChange={(event) => setForm({ ...form, clientID: event.target.value })}
          autoComplete="new-password"
          required
        />
      </label>
      <label className="field-label" htmlFor={`${prefix}-client-secret`}>
        OAuth client secret
        <input
          id={`${prefix}-client-secret`}
          className="input"
          type="password"
          value={form.clientSecret}
          onChange={(event) => setForm({ ...form, clientSecret: event.target.value })}
          autoComplete="new-password"
          required
        />
      </label>
      <label className="field-label" htmlFor={`${prefix}-refresh-token`}>
        OAuth refresh token
        <input
          id={`${prefix}-refresh-token`}
          className="input"
          type="password"
          value={form.refreshToken}
          onChange={(event) => setForm({ ...form, refreshToken: event.target.value })}
          autoComplete="new-password"
          required
        />
      </label>
    </>
  );
}

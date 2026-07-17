"use client";

import { useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import type { AdHocPreview, AdHocSendResult } from "@/lib/types";
import { Modal } from "@/components/Modal";
import { ErrorBanner } from "@/components/ui";

type PreviewState = {
  restaurantId: string;
  preview: AdHocPreview | null;
  error: string | null;
  loading: boolean;
};

export function SendPreviewModal({
  open,
  restaurantIds,
  onClose,
  onSent,
}: {
  open: boolean;
  restaurantIds: string[];
  onClose: () => void;
  onSent?: () => void;
}) {
  const [previews, setPreviews] = useState<PreviewState[]>([]);
  const [sending, setSending] = useState(false);
  const [results, setResults] = useState<AdHocSendResult[] | null>(null);
  const [sendError, setSendError] = useState<string | null>(null);

  useEffect(() => {
    if (!open || restaurantIds.length === 0) return;
    setResults(null);
    setSendError(null);
    setPreviews(
      restaurantIds.map((id) => ({ restaurantId: id, preview: null, error: null, loading: true })),
    );

    let cancelled = false;
    (async () => {
      const loaded = await Promise.all(
        restaurantIds.map(async (id) => {
          try {
            const preview = await adminFetch<AdHocPreview>(`restaurants/${id}/outreach/adhoc-preview`);
            return { restaurantId: id, preview, error: null, loading: false };
          } catch (err) {
            const e = err as Error & { body?: { error?: { message?: string } } };
            return {
              restaurantId: id,
              preview: null,
              error: e.body?.error?.message || e.message || "Preview failed",
              loading: false,
            };
          }
        }),
      );
      if (!cancelled) setPreviews(loaded);
    })();

    return () => {
      cancelled = true;
    };
  }, [open, restaurantIds]);

  const sendableIds = previews.filter((p) => p.preview && !p.error).map((p) => p.restaurantId);

  async function confirmSend() {
    if (sendableIds.length === 0) return;
    setSending(true);
    setSendError(null);
    try {
      if (sendableIds.length === 1) {
        const result = await adminFetch<AdHocSendResult>(
          `restaurants/${sendableIds[0]}/outreach/adhoc-send`,
          { method: "POST" },
        );
        setResults([result]);
      } else {
        const res = await adminFetch<{ items: AdHocSendResult[] }>("outreach/adhoc-send", {
          method: "POST",
          body: { restaurant_ids: sendableIds },
        });
        setResults(res.items);
      }
      onSent?.();
    } catch (err) {
      const e = err as Error & { body?: { error?: { message?: string } } };
      setSendError(e.body?.error?.message || e.message || "Send failed");
    } finally {
      setSending(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={`Send email — ${restaurantIds.length} lead${restaurantIds.length === 1 ? "" : "s"}`} width={720}>
      <ErrorBanner message={sendError} />

      {results ? (
        <div style={{ display: "grid", gap: "0.5rem" }}>
          <p style={{ margin: 0, fontWeight: 600 }}>
            {results.filter((r) => r.sent).length} sent, {results.filter((r) => !r.sent).length} blocked
          </p>
          {results.map((r) => (
            <div
              key={r.restaurant_id}
              className="alert"
              style={{
                background: r.sent ? "var(--ok-bg)" : "var(--bad-bg)",
                color: r.sent ? "var(--ok)" : "var(--bad)",
                borderColor: "transparent",
              }}
            >
              {r.sent ? "Sent" : `Blocked: ${r.error || "unknown error"}`} — {r.restaurant_id}
            </div>
          ))}
        </div>
      ) : (
        <div style={{ display: "grid", gap: "1rem" }}>
          {previews.map((p) => (
            <div key={p.restaurantId} className="card" style={{ display: "grid", gap: "0.5rem" }}>
              {p.loading ? (
                <p style={{ margin: 0, color: "var(--muted)" }}>Loading preview…</p>
              ) : p.error ? (
                <div className="alert alert-error">{p.error}</div>
              ) : p.preview ? (
                <>
                  <div>
                    <strong>{p.preview.restaurant_name}</strong>{" "}
                    <span style={{ color: "var(--muted)" }}>{p.preview.recipient_email}</span>
                  </div>
                  <div>
                    <strong>Subject:</strong> {p.preview.subject}
                  </div>
                  <div
                    style={{
                      border: "1px solid var(--line)",
                      padding: "0.75rem",
                      background: "var(--bg)",
                      maxHeight: 220,
                      overflow: "auto",
                    }}
                    dangerouslySetInnerHTML={{ __html: p.preview.body_html || "<em>No HTML body</em>" }}
                  />
                </>
              ) : null}
            </div>
          ))}
        </div>
      )}

      {!results ? (
        <div className="modal-footer" style={{ padding: 0, border: "none" }}>
          <button className="btn btn-secondary" type="button" onClick={onClose} disabled={sending}>
            Cancel
          </button>
          <button
            className="btn btn-primary"
            type="button"
            onClick={confirmSend}
            disabled={sending || sendableIds.length === 0}
          >
            {sending
              ? "Sending…"
              : `Confirm send (${sendableIds.length} of ${restaurantIds.length})`}
          </button>
        </div>
      ) : (
        <div className="modal-footer" style={{ padding: 0, border: "none" }}>
          <button className="btn btn-primary" type="button" onClick={onClose}>
            Done
          </button>
        </div>
      )}
    </Modal>
  );
}

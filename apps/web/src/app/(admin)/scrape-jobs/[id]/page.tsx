"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import type { ScrapeJob } from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";

export default function ScrapeJobDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const [job, setJob] = useState<ScrapeJob | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [retrying, setRetrying] = useState(false);

  const load = useCallback(async () => {
    try {
      const data = await adminFetch<ScrapeJob>(`scrape-jobs/${id}`);
      setJob(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load job");
    }
  }, [id]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!job) return;
    if (!["queued", "running", "waiting"].includes(job.status)) return;
    const t = setInterval(load, 3000);
    return () => clearInterval(t);
  }, [job, load]);

  async function retry() {
    setRetrying(true);
    setError(null);
    try {
      const data = await adminFetch<ScrapeJob>(`scrape-jobs/${id}/retry`, {
        method: "POST",
      });
      setJob(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Retry failed");
    } finally {
      setRetrying(false);
    }
  }

  if (!job && !error) {
    return <EmptyState message="Loading job…" />;
  }

  return (
    <div>
      <PageHeader
        title={job ? `${job.city} · ${job.niche}` : "Scrape job"}
        subtitle={job ? `Job ${job.id}` : undefined}
        actions={
          <>
            <Link className="btn btn-secondary" href="/scrape-jobs">
              All jobs
            </Link>
            {job?.status === "failed" ? (
              <button
                className="btn btn-primary"
                type="button"
                onClick={retry}
                disabled={retrying}
              >
                {retrying ? "Retrying…" : "Retry failed job"}
              </button>
            ) : null}
          </>
        }
      />
      <ErrorBanner message={error} />

      {job ? (
        <>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))",
              gap: "0.75rem",
              marginBottom: "1rem",
            }}
          >
            <div className="card">
              <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
                Status
              </div>
              <StatusBadge status={job.status} />
            </div>
            <div className="card">
              <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
                Requests (window)
              </div>
              <strong>
                {job.requests_used_window}/{job.max_requests_per_window}
              </strong>
            </div>
            <div className="card">
              <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
                Candidates imported
              </div>
              <strong>{job.progress?.candidates_imported ?? 0}</strong>
            </div>
            <div className="card">
              <div style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
                Cells done
              </div>
              <strong>
                {job.progress?.cells_completed ?? 0}/
                {job.progress?.cells_total ?? 0}
              </strong>
            </div>
          </div>

          {job.last_error ? (
            <div className="alert alert-error" style={{ marginBottom: "1rem" }}>
              <strong>Last error:</strong>
              <pre
                style={{
                  whiteSpace: "pre-wrap",
                  margin: "0.5rem 0 0",
                  fontSize: "0.85rem",
                }}
              >
                {job.last_error}
              </pre>
            </div>
          ) : null}

          <div className="card">
            <h2 style={{ marginTop: 0, fontSize: "1.05rem" }}>Progress</h2>
            <div className="table-wrap" style={{ border: "none" }}>
              <table className="data">
                <tbody>
                  {(
                    [
                      ["Candidates total", job.progress?.candidates_total],
                      ["Pending", job.progress?.candidates_pending],
                      ["Imported", job.progress?.candidates_imported],
                      ["Duplicate", job.progress?.candidates_duplicate],
                      ["Failed", job.progress?.candidates_failed],
                      ["Cells pending", job.progress?.cells_pending],
                      ["Cells failed", job.progress?.cells_failed],
                      ["Waiting reason", job.waiting_reason || "—"],
                      ["Resume at", formatDate(job.resume_at)],
                      ["Updated", formatDate(job.updated_at)],
                    ] as [string, string | number | undefined][]
                  ).map(([label, value]) => (
                    <tr key={label}>
                      <td style={{ color: "var(--muted)" }}>{label}</td>
                      <td>{value ?? "—"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
}

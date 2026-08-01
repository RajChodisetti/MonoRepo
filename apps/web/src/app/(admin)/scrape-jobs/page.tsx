"use client";

import Link from "next/link";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { withBasePath } from "@/lib/base-path";
import { CITIES, NICHES, formatDate } from "@/lib/constants";
import type { ScrapeJob } from "@/lib/types";
import { EmptyState, ErrorBanner, PageHeader, StatusBadge } from "@/components/ui";

export default function ScrapeJobsPage() {
  const [jobs, setJobs] = useState<ScrapeJob[]>([]);
  const [city, setCity] = useState<string>("Melbourne");
  const [niche, setNiche] = useState<string>("restaurant");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const load = useCallback(async () => {
    setError(null);
    try {
      const data = await adminFetch<{ jobs: ScrapeJob[] }>("scrape-jobs", {
        query: { limit: 50 },
      });
      setJobs(data.jobs || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load jobs");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
    const id = setInterval(load, 8000);
    return () => clearInterval(id);
  }, [load]);

  async function onTrigger(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      const res = await adminFetch<{ created: boolean; job: ScrapeJob }>(
        "scrape-jobs",
        {
          method: "POST",
          body: { city, niche },
        },
      );
      await load();
      if (res.job?.id) {
        window.location.assign(withBasePath(`/scrape-jobs/${res.job.id}`));
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to trigger job");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div>
      <PageHeader
        title="Scrape jobs"
        subtitle="Trigger and monitor Google Places city scrapes"
      />
      <ErrorBanner message={error} />

      <form
        onSubmit={onTrigger}
        className="card"
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))",
          gap: "0.75rem",
          alignItems: "end",
          marginBottom: "1.1rem",
        }}
      >
        <label style={{ display: "grid", gap: "0.35rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>City</span>
          <select
            className="select"
            value={city}
            onChange={(e) => setCity(e.target.value)}
          >
            {CITIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
        </label>
        <label style={{ display: "grid", gap: "0.35rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Niche</span>
          <select
            className="select"
            value={niche}
            onChange={(e) => setNiche(e.target.value)}
          >
            {NICHES.map((n) => (
              <option key={n} value={n}>
                {n}
              </option>
            ))}
          </select>
        </label>
        <button className="btn btn-primary" type="submit" disabled={submitting}>
          {submitting ? "Starting…" : "Trigger scrape"}
        </button>
      </form>

      {loading ? <EmptyState message="Loading jobs…" /> : null}

      {!loading && jobs.length === 0 ? (
        <EmptyState message="No scrape jobs yet. Trigger one above." />
      ) : null}

      {jobs.length > 0 ? (
        <div className="table-wrap">
          <table className="data">
            <thead>
              <tr>
                <th>City</th>
                <th>Niche</th>
                <th>Status</th>
                <th>Imported</th>
                <th>Requests</th>
                <th>Updated</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {jobs.map((job) => (
                <tr key={job.id}>
                  <td>{job.city}</td>
                  <td>{job.niche}</td>
                  <td>
                    <StatusBadge status={job.status} />
                  </td>
                  <td>{job.progress?.candidates_imported ?? 0}</td>
                  <td>
                    {job.requests_used_window}/{job.max_requests_per_window}
                  </td>
                  <td>{formatDate(job.updated_at)}</td>
                  <td>
                    <Link href={`/scrape-jobs/${job.id}`}>Open</Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : null}
    </div>
  );
}
